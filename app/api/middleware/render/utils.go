package render

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"github.com/gin-gonic/gin"
)

func getTraceID(ctx *gin.Context) string {
	return ctx.GetString(constant.TraceIDKey)
}
