package router

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"github.com/gin-gonic/gin"
)

// registerPipelineRoutes 注册流水线相关路由
func registerPipelineRoutes(r *gin.RouterGroup, controller *Controller) {
	schema := r.Group("/pipeline/steps")
	{
		schema.GET("/grouped", controller.Pipeline.GroupSteps)
		schema.GET(fmt.Sprintf("/:%s/schema", req.PathParamSchemaName), controller.Pipeline.GetSchema)
	}

	execution := r.Group("/pipeline/executions")
	{
		execution.POST("", controller.Execution.Create)
	}

	trigger := r.Group("/pipeline/triggers")
	{
		trigger.POST("", controller.Pipeline.CreateTrigger)
		trigger.GET("", controller.Pipeline.ListTriggers)
		trigger.GET("/:id", controller.Pipeline.GetTriggerByID)
		trigger.GET(fmt.Sprintf("/by-pipeline/:%s", req.PathParamPipelineID), controller.Pipeline.GetTriggerByPipelineID)
		trigger.DELETE("/:id", controller.Pipeline.DeleteTrigger)
	}
}
