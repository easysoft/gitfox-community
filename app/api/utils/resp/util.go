package resp

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

type ListResponse struct {
	ObjectResponse
	ListArgs any `json:"listArgs"`
}

type ListArgs struct {
	Total    int64 `json:"total"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

type ListNoTotalArgs struct {
	IsLast   bool  `json:"isLast"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"pageSize"`
}

type ObjectResponse struct {
	Code    constant.RespCode `json:"code"`
	Message string            `json:"message"`
	TraceID string            `json:"traceID"`
	Data    interface{}       `json:"data"`
}

// ObjectOK renders a successful object response (200 OK)
func OK(c *gin.Context, code i18n.I18nKey, data interface{}) {
	objectOk(c, code, http.StatusOK, data)
}

// Created renders an object created response (201 Created)
func Created(c *gin.Context, code i18n.I18nKey, data interface{}) {
	objectOk(c, code, http.StatusCreated, data)
}

// Accepted renders an object accepted response (202 Accepted)
func Accepted(c *gin.Context, code i18n.I18nKey, data interface{}) {
	objectOk(c, code, http.StatusAccepted, data)
}

// ObjectError renders an object error response (400 Bad Request or 500 Internal Server Error)
func Error(c *gin.Context, err error, args ...interface{}) {
	httpStatus := errors.GetHttpStatus(err)
	objectErr(c, err, httpStatus, args...)
}

// BadRequest renders an object bad request response (400 Bad Request)
func BadRequest(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusBadRequest, args...)
}

// NotFound renders an object not found response (404 Not Found)
func NotFound(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusNotFound, args...)
}

// InternalError renders an object internal server error response (500 Internal Server Error)
func InternalError(c *gin.Context, err error, args ...interface{}) {
	objectErr(c, err, http.StatusInternalServerError, args...)
}

func Unauthorized(c *gin.Context, err error, args ...interface{}) {
	code := errors.GetI18nKey(err)
	message := i18n.TGin(c, code, args...)
	traceID := getTraceID(c)
	c.AbortWithStatusJSON(http.StatusUnauthorized, ObjectResponse{
		Code:    constant.RespFailCode,
		Message: message,
		TraceID: traceID,
		Data:    nil,
	})
}

// ListSuccess renders a successful list response (200 OK)
func ListSuccess(c *gin.Context, data interface{}, listArgs interface{}, code i18n.I18nKey, args ...interface{}) {
	listSuccess(c, data, listArgs, code, args)
}

// list renders a list response
// Used for GET list requests or POST/PATCH/DELETE batch operations (partial or full failure)
// data: list data
// listArgs: list arguments, e.g. total count, page, pageSize
// code: i18n key, e.g. "repo.branch.list.success"
func listSuccess(c *gin.Context, data interface{}, listArgs interface{}, code i18n.I18nKey, args ...interface{}) {
	traceID := getTraceID(c)
	message := i18n.TGin(c, code, args...)

	c.JSON(http.StatusOK, ListResponse{
		ObjectResponse: ObjectResponse{
			Code:    constant.RespSuccessCode,
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
func objectErr(c *gin.Context, err error, httpStatus int, args ...interface{}) {
	code := errors.GetI18nKey(err)
	var message string
	if code == i18n.ErrorDefault {
		message = err.Error()
	} else {
		message = i18n.TGin(c, code, args...)
	}
	traceID := getTraceID(c)
	log := logger.GetLoggerFromGin(c)
	log.Errorw("API error response",
		"error", err,
		"code", code,
		"message", message,
	)
	c.JSON(httpStatus, ObjectResponse{
		Code:    constant.RespFailCode,
		Message: message,
		TraceID: traceID,
		Data:    nil,
	})
}

func objectOk(c *gin.Context, code i18n.I18nKey, httpStatus int, data interface{}) {
	message := i18n.TGin(c, code)
	traceID := getTraceID(c)
	c.JSON(httpStatus, ObjectResponse{
		Code:    constant.RespSuccessCode,
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

// GitError writes an error response for git protocol endpoints.
//
// Git's internal curl does NOT read the response body for HTTP errors (4xx/5xx),
// so we return HTTP 200 and write the error in git protocol format.
//
// Protocol format depends on the git service:
//   - upload-pack (clone/fetch/info-refs): pkt-line ERR packet
//   - receive-pack (push): sideband band 3 (fatal error) packet
//
// The httpStatus parameter is only used for server-side logging.
func GitError(c *gin.Context, err error, httpStatus int, isPush bool) {
	code := errors.GetI18nKey(err)
	var message string
	if code == i18n.ErrorDefault {
		message = err.Error()
	} else {
		message = i18n.TGin(c, code)
	}
	log := logger.GetLoggerFromGin(c)
	log.Errorw("git API error response",
		"error", err,
		"code", code,
		"message", message,
		"http_status", httpStatus,
		"is_push", isPush,
	)
	// Must use HTTP 200 so git/curl reads the response body.
	// If headers were already sent (e.g. ServicePack stream), this is a no-op.
	c.Status(http.StatusOK)

	if isPush {
		// receive-pack expects sideband-multiplexed output.
		// Band 3 = fatal error (displayed to user via git's stderr).
		// Format: <4-hex-len><\x03><message>
		pkt := "\x03" + message
		c.Writer.Write(gitPacketWrite(pkt))
	} else {
		// upload-pack expects pkt-line format.
		// ERR packet tells git to display the message as an error.
		// Format: <4-hex-len>ERR <message>
		errPkt := "ERR " + message
		c.Writer.Write(gitPacketWrite(errPkt))
	}
	// Flush packet terminates the stream
	c.Writer.Write([]byte("0000"))
}

// gitPacketWrite encodes a string as a git pkt-line packet.
// Format: 4 hex chars (total length including the 4 chars) + payload.
func gitPacketWrite(str string) []byte {
	s := strconv.FormatInt(int64(len(str)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}
