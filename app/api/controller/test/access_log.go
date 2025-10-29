package test

import (
	"time"

	"github.com/gin-gonic/gin"
)

// @Summary 访问日志测试接口
// @Description 用于测试访问日志中间件功能，返回固定响应并包含 trace_id
// @Tags test
// @Accept application/json
// @Produce application/json
// @Param req query string false "请求参数" default(fixed-req)
// @Success 200 {object} map[string]interface{} "测试成功"
// @Router /public/test_accessLog [get]
func AccessLogTest(c *gin.Context) {
	// /public/test_accessLog?req=fixed-req
	reqVal := c.DefaultQuery("req", "fixed-req")

	// 固定耗时
	time.Sleep(200 * time.Millisecond)

	// 返回固定值，并回传 trace_id 便于对照日志
	c.JSON(200, gin.H{
		"request_value":  reqVal,
		"response_value": "fixed-resp",
		"trace_id":       c.GetString("trace_id"),
	})
}
