// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package request

import "net/http"

const (
	ContentType    = "Content-Type"
	ContentLength  = "Content-Length"
	Disposition    = "Content-Disposition"
	Etag           = "ETag"
	LastModify     = "Last-Modified"
	Location       = "Location"
	HeaderRange    = "Range"
	HeadDockerUUID = "Docker-Upload-UUID"
)

type ResponseHeaderWriter interface {
	Write(w http.ResponseWriter)
}

type commonHttpResponseWriter struct {
	writeFn func(w http.ResponseWriter)
}

func (c *commonHttpResponseWriter) Write(w http.ResponseWriter) {
	c.writeFn(w)
}

func NewResponseWriter(fn func(w http.ResponseWriter)) ResponseHeaderWriter {
	return &commonHttpResponseWriter{writeFn: fn}
}
