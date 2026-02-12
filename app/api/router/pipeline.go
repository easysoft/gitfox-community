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
		executionWithID := execution.Group(":id")
		{
			executionWithID.GET("", controller.Execution.Get)
			executionWithID.POST("/cancel", controller.Execution.Cancel)

		}
	}

	logPath := fmt.Sprintf("/:%s/:%s/:%s/:%s", req.PathParamPipelineID, req.PathParamExecutionID, req.PathParamStageID, req.PathParamStepID)
	r.GET(logPath+"/stream", controller.Execution.LogStream)
	r.POST(logPath+"/logs", controller.Execution.CreateLog)

	trigger := r.Group(fmt.Sprintf("/pipelines/:%s/triggers", req.PathParamPipelineID))
	{
		trigger.POST("", controller.Pipeline.CreateTrigger)
		trigger.GET("", controller.Pipeline.ListTriggers)
		trigger.GET(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.GetTriggerByID)
		trigger.PUT(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.UpdateTrigger)
		trigger.DELETE(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.DeleteTrigger)
	}
}
