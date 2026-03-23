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

		// 扫描任务相关路由
		tasks := rule.Group("/tasks")
		tasks.POST("", controller.Scan.ScanTaskCreate)
		tasks.POST("/list", controller.Scan.ScanTaskList)
		tasks.GET("/:id", controller.Scan.ScanTaskGet)
		tasks.POST("/:id/report", controller.Scan.ScanTaskUploadReport)
		tasks.POST("/:id/issues", controller.Scan.ScanTaskIssueList)
		tasks.POST("/:id/retry", controller.Scan.ScanTaskRetry)

		// 扫描问题相关路由
		issues := rule.Group("/issues")
		issues.GET("/:id", controller.Scan.ScanIssueGet)
		issues.POST("/list", controller.Scan.ScanIssueList)
		issues.POST("/batch", controller.Scan.ScanIssueListByIds)
		issues.PUT("/status", controller.Scan.ScanIssueUpdateStatus)
		issues.GET("/file-tree", controller.Scan.ScanIssueFileTree)
		issues.GET("/rule-tree", controller.Scan.ScanIssueRuleTree)

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
		// 仓库维度扫描计划路由：/scan/repos/:repoID/scanplans/...
		repos := rule.Group("/repos/:repoID")
		{
			scanPlans := repos.Group("/scanplans")
			scanPlans.POST("", controller.Scan.ScanPlanCreate)
			scanPlans.POST("/list", controller.Scan.ScanPlanList)

			planWithID := scanPlans.Group("/:planID")
			{
				planWithID.GET("", controller.Scan.ScanPlanGet)
				planWithID.GET("/rules", controller.Scan.ScanPlanListRules)
				planWithID.GET("/plan-rules", controller.Scan.ScanPlanRules)
				planWithID.PATCH("", controller.Scan.ScanPlanUpdate)
				planWithID.DELETE("", controller.Scan.ScanPlanDelete)
				planWithID.POST("/bind", controller.Scan.ScanPlanBindSolution)
				planWithID.POST("/unbind", controller.Scan.ScanPlanUnbindSolution)
				planWithID.GET("/unbind", controller.Scan.ScanPlanListUnbindSolution)

				// 扫描计划级条件路由（triggerID=0）
				conditions := planWithID.Group("/conditions")
				conditions.POST("", controller.Scan.ScanPlanConditionCreate)
				conditions.GET("", controller.Scan.ScanPlanConditionList)
				conditions.POST("/batch/new", controller.Scan.ScanPlanConditionBatchCreate)
				conditions.POST("/batch/delete", controller.Scan.ScanPlanConditionBatchDelete)

				conditionWithID := conditions.Group("/:conditionID")
				{
					conditionWithID.GET("", controller.Scan.ScanPlanConditionFind)
					conditionWithID.PATCH("", controller.Scan.ScanPlanConditionUpdate)
					conditionWithID.DELETE("", controller.Scan.ScanPlanConditionDelete)
				}

				// 触发器路由
				triggers := planWithID.Group("/triggers")
				triggers.GET("", controller.Scan.ScanTriggerList)
				triggers.POST("", controller.Scan.ScanTriggerCreate)

				triggerWithID := triggers.Group("/:triggerID")
				{
					triggerWithID.GET("", controller.Scan.ScanTriggerGet)
					triggerWithID.PATCH("", controller.Scan.ScanTriggerUpdate)
					triggerWithID.DELETE("", controller.Scan.ScanTriggerDelete)

					// 触发器条件路由
					conditions := triggerWithID.Group("/conditions")
					conditions.GET("", controller.Scan.ScanConditionList)
					conditions.POST("/new", controller.Scan.ScanConditionBatchCreate)
					conditions.POST("/delete", controller.Scan.ScanConditionBatchDelete)
				}
			}
		}
	}
}
