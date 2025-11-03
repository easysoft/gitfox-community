package resp

import (
	"fmt"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ListResponse represents a response structure for list data
type ListResponse struct {
	ObjectResponse
	ListArgs
}

type ListArgs struct {
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

// ObjectResponse represents a response structure for single object data
type ObjectResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	TraceID string      `json:"traceID"`
	Data    interface{} `json:"data"`
}

// ObjectOK renders a successful object response (200 OK)
func OK(c *gin.Context, code string, data interface{}) {
	objectOk(c, code, http.StatusOK, data)
}

// Created renders an object created response (201 Created)
func Created(c *gin.Context, code string, data interface{}) {
	objectOk(c, code, http.StatusCreated, data)
}

// ObjectError renders an object error response (400 Bad Request or 500 Internal Server Error)
func Error(c *gin.Context, err error, args ...interface{}) {
	httpStatus := errors.GetHttpStatus(err)
	objectErr(c, err, httpStatus)
}

// BadRequest renders an object bad request response (400 Bad Request)
func BadRequest(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusBadRequest)
}

// NotFound renders an object not found response (404 Not Found)
func NotFound(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusNotFound)
}

// InternalError renders an object internal server error response (500 Internal Server Error)
func InternalError(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusInternalServerError)
}

// ListSuccess renders a successful list response (200 OK)
func ListSuccess(c *gin.Context, data interface{}, listArgs ListArgs, code string, args ...interface{}) {
	list(c, data, listArgs, code, args)
}

// list renders a list response
// Used for GET list requests or POST/PATCH/DELETE batch operations (partial or full failure)
// data: list data
// listArgs: list arguments, e.g. total count, page, pageSize
// code: i18n key, e.g. "repo.branch.list.success"
func list(c *gin.Context, data interface{}, listArgs ListArgs, code string, args ...interface{}) {
	traceID := getTraceID(c)
	message := i18n.TGin(c, code, args...)

	c.JSON(http.StatusOK, ListResponse{
		ObjectResponse: ObjectResponse{
			Code:    code,
			Message: message,
			TraceID: traceID,
			Data:    data,
		},
		ListArgs: listArgs,
	})
}

// object renders an object response
// Used for GET/POST/PATCH/DELETE operations on single objects
// code: i18n key, e.g. "repo.branch.create.success"
// httpStatus: HTTP status code
// data: object data
func objectErr(c *gin.Context, err error, httpStatus int, data ...interface{}) {
	code := errors.GetI18nKey(err)
	message := i18n.TGin(c, code) + fmt.Sprintf(";%s", err.Error())
	traceID := getTraceID(c)
	c.JSON(httpStatus, ObjectResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Data:    data,
	})
}

func objectOk(c *gin.Context, code string, httpStatus int, data ...interface{}) {
	message := i18n.TGin(c, code)
	traceID := getTraceID(c)
	c.JSON(httpStatus, ObjectResponse{
		Code:    code,
		Message: message,
		TraceID: traceID,
		Data:    data,
	})
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
