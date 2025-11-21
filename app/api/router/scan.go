package router

import (
	"github.com/gin-gonic/gin"
)

// registerScanRoutes 注册扫描相关路由
func registerScanRoutes(r *gin.RouterGroup, controller *Controller) {
	ruleSets := r.Group("/scan/rulesets")
	{
		ruleSets.POST("/", controller.RuleSet.Create)
		ruleSets.POST("/list", controller.RuleSet.List)
		ruleSetsWithID := ruleSets.Group(":id")
		{
			ruleSetsWithID.GET("", controller.RuleSet.Get)
			ruleSetsWithID.PUT("", controller.RuleSet.Update)
			ruleSetsWithID.DELETE("", controller.RuleSet.Delete)
		}
	}
}
