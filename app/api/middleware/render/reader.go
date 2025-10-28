package render

import (
	pkg_errors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Done 结束响应
func Done(ctx *gin.Context, err error, args ...interface{}) error {
	i18nKey := pkg_errors.GetI18nKey(err)
	message := i18n.TGin(ctx, i18nKey, args...)
	traceID := getTraceID(ctx)
	httpStatus := pkg_errors.GetHttpStatus(err)

	ctx.JSON(httpStatus, Response{
		Code:    i18nKey,
		Message: message,
		TraceID: traceID,
	})
	return nil
}

// JSON 自定义响应
func JSON(ctx *gin.Context, data interface{}, err error, args ...interface{}) error {
	i18nKey := pkg_errors.GetI18nKey(err)
	message := i18n.TGin(ctx, i18nKey, args...)
	traceID := getTraceID(ctx)
	httpStatus := pkg_errors.GetHttpStatus(err)

	ctx.JSON(httpStatus, Response{
		Code:    i18nKey,
		Message: message,
		TraceID: traceID,
		Data:    data,
	})
	return nil
}

// List 列表响应
func List(ctx *gin.Context, data interface{}, listArgs ListArgs, err error, args ...interface{}) error {
	i18nKey := pkg_errors.GetI18nKey(err)
	message := i18n.TGin(ctx, i18nKey, args...)
	traceID := getTraceID(ctx)
	httpStatus := pkg_errors.GetHttpStatus(err)

	ctx.JSON(httpStatus, ListResponse{
		Response: Response{
			Code:    i18nKey,
			Message: message,
			TraceID: traceID,
			Data:    data,
		},
		ListArgs: listArgs,
	})
	return nil
}
