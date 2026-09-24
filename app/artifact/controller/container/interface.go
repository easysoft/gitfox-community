// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package container

import "net/http"

// HttpResponseWriter 封装 OCI 协议的 HTTP 响应写入逻辑。
type HttpResponseWriter interface {
	Write(w http.ResponseWriter)
}

type commonHttpResponseWriter struct {
	writeFn func(w http.ResponseWriter)
}

func (c *commonHttpResponseWriter) Write(w http.ResponseWriter) {
	c.writeFn(w)
}

func NewResponseWriter(fn func(w http.ResponseWriter)) HttpResponseWriter {
	return &commonHttpResponseWriter{writeFn: fn}
}
