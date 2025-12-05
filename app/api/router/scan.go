package router

import (
	"github.com/gin-gonic/gin"
)

// registerScanRoutes 注册扫描相关路由
func registerScanRoutes(r *gin.RouterGroup, controller *Controller) {
	rule := r.Group("/scan")
	{
		// 规则相关路由
		rules := rule.Group("/rules")
		rules.POST("/migrate", controller.Scan.MigrateRules)
		rules.POST("/list", controller.Scan.ListRules)
		rulesWithID := rules.Group(":id")
		{
			rulesWithID.GET("", controller.Scan.GetRule)
			rulesWithID.PUT("/status", controller.Scan.UpdateRuleStatus)
		}

		// 规则集相关路由
		ruleSets := rule.Group("/rulesets")
		ruleSets.POST("", controller.Scan.Create)
		ruleSets.POST("/list", controller.Scan.List)
		ruleSetsWithID := ruleSets.Group(":id")
		{
			ruleSetsWithID.GET("", controller.Scan.Get)
			ruleSetsWithID.PUT("", controller.Scan.Update)
			ruleSetsWithID.DELETE("", controller.Scan.Delete)
			ruleSetsWithID.GET("/rules/bind", controller.Scan.GetBindRule)
			ruleSetsWithID.GET("/rules/unbind", controller.Scan.GetUnBindRule)
			ruleSetsWithID.POST("/rules/bind", controller.Scan.BindRule)
			ruleSetsWithID.POST("/rules/unbind", controller.Scan.UnBindRule)
		}
	}
}
