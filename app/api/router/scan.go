package router

import (
	"github.com/gin-gonic/gin"
)

// registerScanRoutes 注册扫描相关路由
func registerScanRoutes(r *gin.RouterGroup, controller *Controller) {
	rule := r.Group("/scan")
	{

		rule.GET("/types", controller.Scan.AllTypes)
		rule.GET("/plugins", controller.Scan.AllPlugins)
		rule.GET("/langs", controller.Scan.AllLangs)
		rule.GET("/tags", controller.Scan.AllTags)
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

		// 扫描方案相关路由
		solutions := rule.Group("/solutions")
		solutions.POST("", controller.Scan.SolutionCreate)
		solutions.POST("/list", controller.Scan.SolutionList)
		solutionsWithID := solutions.Group(":id")
		{
			solutionsWithID.GET("", controller.Scan.SolutionGet)
			solutionsWithID.PUT("", controller.Scan.SolutionUpdate)
			solutionsWithID.DELETE("", controller.Scan.SolutionDelete)
			solutionsWithID.PATCH("/status", controller.Scan.SolutionUpdateStatus)
			solutionsWithID.PATCH("/draft", controller.Scan.SolutionUpdateStatus)
			solutionsWithID.GET("/rulesets", controller.Scan.SolutionListRulesets)
			solutionsWithID.GET("/rulesets/unbind", controller.Scan.SolutionListUnbindRulesets)
			solutionsWithID.POST("/rulesets/bind", controller.Scan.SolutionBindRulesets)
			solutionsWithID.POST("/rulesets/unbind", controller.Scan.SolutionUnbindRulesets)
		}
	}
}
