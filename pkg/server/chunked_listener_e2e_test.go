// Copyright (c) 2023-2024 北京渠成软件有限公司(Beijing Qucheng Software Co., Ltd. www.qucheng.com) All rights reserved.
// Use of this source code is covered by the following dual licenses:
// (1) Z PUBLIC LICENSE 1.2 (ZPL 1.2)
// (2) Affero General Public License 3.0 (AGPL 3.0)
// license that can be found in the LICENSE file.

package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

// TestEndToEnd_ChunkedPatch_NoTE: 端到端复现 docker 28 行为：
// client 直接在 wire 上发 chunked body 但不带 Transfer-Encoding 头。
// 经 WrapChunkedFix 包装后,net/http server 应能正确读出 body。
func TestEndToEnd_ChunkedPatch_NoTE(t *testing.T) {
	const wantPayload = "hello-from-docker-28-chunked-body"

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	wrapped := WrapChunkedFix(ln)

	gotBody := make(chan []byte, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/v2/global/foo/bar/blobs/uploads/abc", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody <- b
		w.WriteHeader(http.StatusAccepted)
	})

	srv := &http.Server{Handler: mux}
	var serveWG sync.WaitGroup
	serveWG.Add(1)
	go func() {
		defer serveWG.Done()
		_ = srv.Serve(wrapped)
	}()
	defer func() {
		_ = srv.Close()
		serveWG.Wait()
	}()

	// 等 server 起来
	time.Sleep(50 * time.Millisecond)

	// 直接拨 TCP，手写畸形请求：chunked body + 无 TE 头
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	chunkSize := fmt.Sprintf("%x", len(wantPayload))
	wire := "PATCH /v2/global/foo/bar/blobs/uploads/abc HTTP/1.1\r\n" +
		"Host: " + addr + "\r\n" +
		"User-Agent: docker/28.4.0 UpstreamClient(Docker-Client/28.4.0)\r\n" +
		"Content-Type: application/octet-stream\r\n" +
		"Accept-Encoding: gzip\r\n" +
		"\r\n" +
		chunkSize + "\r\n" + wantPayload + "\r\n" +
		"0\r\n\r\n"

	if _, err := conn.Write([]byte(wire)); err != nil {
		t.Fatal(err)
	}

	// 读响应避免连接阻塞
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	select {
	case got := <-gotBody:
		if string(got) != wantPayload {
			t.Fatalf("body mismatch: want %q, got %q (len=%d)", wantPayload, got, len(got))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler not invoked or body not received in time")
	}
}

// TestEndToEnd_NormalPostPassThrough: 普通带 Content-Length 的 POST 必须原样透传，
// 不能被 listener 误改。
func TestEndToEnd_NormalPostPassThrough(t *testing.T) {
	const wantPayload = `{"x":1}`

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	addr := ln.Addr().String()

	wrapped := WrapChunkedFix(ln)

	gotBody := make(chan []byte, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/foo", func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		gotBody <- b
		w.WriteHeader(http.StatusOK)
	})
	srv := &http.Server{Handler: mux}
	var serveWG sync.WaitGroup
	serveWG.Add(1)
	go func() {
		defer serveWG.Done()
		_ = srv.Serve(wrapped)
	}()
	defer func() {
		_ = srv.Close()
		serveWG.Wait()
	}()
	time.Sleep(50 * time.Millisecond)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	wire := fmt.Sprintf(
		"POST /api/v2/foo HTTP/1.1\r\nHost: %s\r\nContent-Type: application/json\r\nContent-Length: %d\r\n\r\n%s",
		addr, len(wantPayload), wantPayload,
	)
	if _, err := conn.Write([]byte(wire)); err != nil {
		t.Fatal(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()

	select {
	case got := <-gotBody:
		if string(got) != wantPayload {
			t.Fatalf("body mismatch: want %q, got %q", wantPayload, got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler not invoked")
	}
}
