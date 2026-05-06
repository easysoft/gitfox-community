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
