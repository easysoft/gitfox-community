package render

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ListResponse represents a response structure for list data
type ListResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	TraceID string      `json:"traceID"`
	Total   int64       `json:"total"`
	Data    interface{} `json:"data"`
}

// ObjectResponse represents a response structure for single object data
type ObjectResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	TraceID string      `json:"traceID"`
	Data    interface{} `json:"data"`
}

// getTraceID retrieves traceID from Gin Context
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if traceIDStr, ok := traceID.(string); ok {
			return traceIDStr
		}
	}
	return ""
}

// list renders a list response
// Used for GET list requests or POST/PATCH/DELETE batch operations (partial or full failure)
// code: i18n key, e.g. "repo.branch.list.success"
// data: list data
// total: total count of data
// httpStatus: HTTP status code, e.g. http.StatusOK
func list(c *gin.Context, code string, data interface{}, total int64, httpStatus int) {
	message := i18n.TGin(c, code)
	traceID := getTraceID(c)

	c.JSON(httpStatus, ListResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Total:   total,
		Data:    data,
	})
}

// ListSuccess renders a successful list response (200 OK)
func ListSuccess(c *gin.Context, code string, data interface{}, total int64) {
	list(c, code, data, total, http.StatusOK)
}

// ListError renders a list error response (400 Bad Request)
func ListError(c *gin.Context, code string, data interface{}, total int64) {
	list(c, code, data, total, http.StatusBadRequest)
}

// object renders an object response
// Used for GET/POST/PATCH/DELETE operations on single objects
// code: i18n key, e.g. "repo.branch.create.success"
// data: object data
// httpStatus: HTTP status code
func object(c *gin.Context, code string, data interface{}, httpStatus int) {
	message := i18n.TGin(c, code)
	traceID := getTraceID(c)

	c.JSON(httpStatus, ObjectResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Data:    data,
	})
}

// ObjectCreated renders an object created response (201 Created)
func ObjectCreated(c *gin.Context, code string, data interface{}) {
	object(c, code, data, http.StatusCreated)
}

// ObjectOK renders a successful object response (200 OK)
func ObjectOK(c *gin.Context, code string, data interface{}) {
	object(c, code, data, http.StatusOK)
}

// ObjectBadRequest renders an object bad request response (400 Bad Request)
func ObjectBadRequest(c *gin.Context, code string, data interface{}) {
	object(c, code, data, http.StatusBadRequest)
}

// ObjectNotFound renders an object not found response (404 Not Found)
func ObjectNotFound(c *gin.Context, code string, data interface{}) {
	object(c, code, data, http.StatusNotFound)
}

// ObjectInternalError renders an object internal server error response (500 Internal Server Error)
func ObjectInternalError(c *gin.Context, code string, data interface{}) {
	object(c, code, data, http.StatusInternalServerError)
}
