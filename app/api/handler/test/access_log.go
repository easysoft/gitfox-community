package test

import (
	"time"

	"github.com/gin-gonic/gin"
)

// AccessLogTest 固定请求值、固定返回值、固定耗时，并回传 trace_id
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
