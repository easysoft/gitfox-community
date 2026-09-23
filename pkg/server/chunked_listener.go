// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

// Package server 提供 net.Listener / http.Server 层的协议级兼容工具。
package server

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"net"
	"strings"
)

// WrapChunkedFix 包装 net.Listener，对 docker 28 客户端在 OCI Registry
// /v2/.../blobs/uploads/{uuid} 路径上发 PATCH/PUT 时漏带 Transfer-Encoding 头
// 但 body 走 chunked 字节流的请求做兼容：在 net/http 解析请求前，
// 在请求头注入 "Transfer-Encoding: chunked"，让 net/http 用 chunked
// reader 正确解析 body。
//
// 不命中目标路径或已带 Content-Length / Transfer-Encoding 的请求原样透传。
//
// 仅用于明文 HTTP 监听端口；如果对 listener 启用了 TLS，应在 TLS 解密后再包装，
// 否则会破坏密文流。
func WrapChunkedFix(ln net.Listener) net.Listener {
	return &chunkedFixListener{Listener: ln}
}

type chunkedFixListener struct {
	net.Listener
}

func (l *chunkedFixListener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return newChunkedFixConn(c), nil
}

// chunkedFixConn 包装 net.Conn，对每条 HTTP 请求的请求行 + 请求头做嗅探与改写。
// 只改写 header，body 原样透传。支持 keep-alive 连接复用。
type chunkedFixConn struct {
	net.Conn

	br *bufio.Reader

	// 当前请求改写后的 header 缓冲。
	pending bytes.Buffer

	// 当前阶段：true 表示 header 已处理完毕，正在透传 body。
	inBody bool

	// body 剩余字节数（按 Content-Length 计）。
	// -1 表示 chunked / 不定长，透传到连接关闭。
	bodyRemaining int64
}

func newChunkedFixConn(c net.Conn) *chunkedFixConn {
	return &chunkedFixConn{
		Conn: c,
		br:   bufio.NewReaderSize(c, 8192),
	}
}

// header 总大小上限，超过则放弃改写、原样透传，避免内存放大。
const maxPatchedHeaderBytes = 64 * 1024

func (c *chunkedFixConn) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// 1) 先排空已落定的 header 缓冲
	if c.pending.Len() > 0 {
		return c.pending.Read(p)
	}

	// 2) body 阶段透传
	if c.inBody {
		return c.readBody(p)
	}

	// 3) header 阶段：捕获并改写
	if err := c.captureAndPatchHeader(); err != nil {
		if c.pending.Len() > 0 {
			n, _ := c.pending.Read(p)
			return n, nil
		}
		return 0, err
	}

	if c.pending.Len() > 0 {
		return c.pending.Read(p)
	}
	return c.br.Read(p)
}

// captureAndPatchHeader 从 br 读完一段完整 HTTP 请求头（直到 \r\n\r\n），
// 必要时注入 Transfer-Encoding，写入 c.pending；
// 同时根据 Content-Length / Transfer-Encoding 推算 body 透传策略。
func (c *chunkedFixConn) captureAndPatchHeader() error {
	var head bytes.Buffer

	for {
		if head.Len() > maxPatchedHeaderBytes {
			c.pending.Write(head.Bytes())
			c.inBody = true
			c.bodyRemaining = -1
			return nil
		}

		line, err := c.br.ReadBytes('\n')
		if len(line) > 0 {
			head.Write(line)
		}
		if err != nil {
			if head.Len() > 0 {
				c.pending.Write(head.Bytes())
				c.inBody = true
				c.bodyRemaining = -1
			}
			return err
		}
		if bytes.Equal(line, []byte("\r\n")) {
			break
		}
	}

	patched := patchHeader(head.Bytes())
	c.pending.Write(patched)
	c.inBody = true
	c.bodyRemaining = parseBodyLen(patched)
	return nil
}

// readBody 透传 body 字节，按 Content-Length 计数；归零后切回 header 阶段。
// chunked / 未知长度（bodyRemaining == -1）则一直透传，由对端关闭连接结束。
func (c *chunkedFixConn) readBody(p []byte) (int, error) {
	if c.bodyRemaining == 0 {
		c.inBody = false
		return c.Read(p)
	}

	if c.bodyRemaining > 0 && int64(len(p)) > c.bodyRemaining {
		p = p[:c.bodyRemaining]
	}

	n, err := c.br.Read(p)
	if c.bodyRemaining > 0 {
		c.bodyRemaining -= int64(n)
		if c.bodyRemaining < 0 {
			c.bodyRemaining = 0
		}
	}
	return n, err
}

// patchHeader 检查请求行 + 请求头，必要时注入 "Transfer-Encoding: chunked"。
// 命中条件（全部满足才改写）：
//   - 请求方法是 PATCH 或 PUT
//   - URI 同时包含 "/v2/" 与 "/blobs/uploads/"
//   - header 中无 Transfer-Encoding
//   - header 中无 Content-Length
func patchHeader(head []byte) []byte {
	firstLine, _, ok := bytes.Cut(head, []byte("\r\n"))
	if !ok {
		return head
	}
	first := string(firstLine)
	if !((strings.HasPrefix(first, "PATCH ") || strings.HasPrefix(first, "PUT ")) &&
		strings.Contains(first, "/v2/") &&
		strings.Contains(first, "/blobs/uploads/")) {
		return head
	}

	lower := bytes.ToLower(head)
	if bytes.Contains(lower, []byte("\r\ntransfer-encoding:")) {
		return head
	}
	if bytes.Contains(lower, []byte("\r\ncontent-length:")) {
		return head
	}

	end := bytes.LastIndex(head, []byte("\r\n\r\n"))
	if end < 0 {
		return head
	}

	out := make([]byte, 0, len(head)+32)
	out = append(out, head[:end+2]...) // 含倒数第二个 \r\n
	out = append(out, "Transfer-Encoding: chunked\r\n"...)
	out = append(out, "\r\n"...)
	return out
}

// parseBodyLen 从落定后的请求头里推算 body 透传字节数。
//   - 命中 Transfer-Encoding：返回 -1，透传到连接关闭
//   - 命中 Content-Length：返回数值（解析失败按 -1 处理）
//   - 都没有：按 RFC 7230 视为无 body，返回 0
func parseBodyLen(head []byte) int64 {
	rest := head
	// 跳过请求行
	if _, after, ok := bytes.Cut(rest, []byte("\r\n")); ok {
		rest = after
	}
	for {
		line, after, ok := bytes.Cut(rest, []byte("\r\n"))
		if !ok || len(line) == 0 {
			break
		}
		name, val, ok := bytes.Cut(line, []byte(":"))
		if ok {
			n := strings.ToLower(strings.TrimSpace(string(name)))
			v := strings.TrimSpace(string(val))
			switch n {
			case "transfer-encoding":
				return -1
			case "content-length":
				cl, err := parseUint64(v)
				if err != nil {
					return -1
				}
				return cl
			}
		}
		rest = after
	}
	return 0
}

var errInvalidInt = errors.New("invalid int")

func parseUint64(s string) (int64, error) {
	if s == "" {
		return 0, errInvalidInt
	}
	var n int64
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch < '0' || ch > '9' {
			return 0, errInvalidInt
		}
		n = n*10 + int64(ch-'0')
	}
	return n, nil
}

// 编译期断言：chunkedFixConn 仍然是合法的 net.Conn。
var _ net.Conn = (*chunkedFixConn)(nil)
var _ io.Reader = (*chunkedFixConn)(nil)
