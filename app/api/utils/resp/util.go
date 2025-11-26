package resp

import (
	"fmt"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ListResponse represents a response structure for list data
// ListResponse 列表类型统一响应结构体
// @Schema
// @Description 列表查询接口的统一响应结构，继承基础响应字段并补充分页元数据
// @allOf                       // 合并基础结构体（Swagger 3.0+ 支持）
// @property {} ObjectResponse  // 继承基础响应的所有字段
// @property listArgs ListArgs  `json:"listArgs" description:"分页元数据" $ref:"#/components/schemas/ListArgs"`
type ListResponse struct {
	ObjectResponse
	ListArgs any
}

// ListArgs 列表分页元数据（描述列表的总数、页码、每页条数）
// @Schema
// @Description 列表类接口的分页参数集合，用于前端分页展示
// @property total int64 `json:"total" description:"符合条件的总记录数" example:"100"`
// @property page int64 `json:"page" description:"当前页码（从1开始）" example:"1"`
// @property pageSize int64 `json:"pageSize" description:"每页显示的记录数" example:"20"`
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

// ObjectResponse represents a response structure for single object data
// ObjectResponse 基础响应结构体（所有接口响应的公共结构）
// @Schema
// @Description 所有接口的基础响应模板，包含状态码、提示信息、追踪ID和业务数据
// @property code constant.RespCode `json:"code" description:"状态码（success,failure）" example:"success"`
// @property message string `json:"message" description:"响应提示信息"`
// @property traceID string `json:"traceID" description:"请求追踪ID"`
// @property data interface{} `json:"data" description:"业务核心数据（object 或 list）"`
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
	message := i18n.TGin(c, code, args...) + fmt.Sprintf(";%s", err.Error())
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
	message := i18n.TGin(c, code, args...) + fmt.Sprintf(";%s", err.Error())
	traceID := getTraceID(c)
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
