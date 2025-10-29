package test

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// TestController 处理测试相关的 HTTP 请求
type TestController struct{}

// NewTestController 创建测试控制器实例
func NewTestController() *TestController {
	return &TestController{}
}

// @Summary 访问日志测试接口
// @Description 用于测试访问日志中间件功能，返回固定响应并包含 trace_id
// @Tags test
// @Accept application/json
// @Produce application/json
// @Param req query string false "请求参数" default(fixed-req)
// @Success 200 {object} map[string]interface{} "测试成功"
// @Router /public/test_accessLog [get]
func (tc *TestController) AccessLogTest(c *gin.Context) {
	// /public/test_accessLog?req=fixed-req
	reqVal := c.DefaultQuery("req", "fixed-req")

	// 固定耗时
	time.Sleep(201 * time.Millisecond)

	c.JSON(200, gin.H{
		"request_value":  reqVal,
		"response_value": i18n.TGin(c, "test.fixed-resp", reqVal),
		"trace_id":       c.GetString("trace_id"),
	})
}
