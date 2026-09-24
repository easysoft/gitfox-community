// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package router

import (
	"fmt"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
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
		execution.POST("/list-by-pr", controller.Execution.ListByPr)
		executionWithID := execution.Group(":id")
		{
			executionWithID.GET("", controller.Execution.Get)
			executionWithID.POST("/cancel", controller.Execution.Cancel)

		}
	}

	logs := r.Group(fmt.Sprintf("/pipelines/:%s/executions/:%s", req.PathParamPipelineID, req.PathParamExecutionID))
	{
		logs.GET(fmt.Sprintf("logs/:%s/:%s", req.PathParamStageID, req.PathParamStepID), controller.Execution.GetFileLogs)
		logs.POST(fmt.Sprintf("logs/:%s/:%s", req.PathParamStageID, req.PathParamStepID), controller.Execution.CreateFileLogs)
	}

	trigger := r.Group(fmt.Sprintf("/pipelines/:%s/triggers", req.PathParamPipelineID))
	{
		trigger.POST("", controller.Pipeline.CreateTrigger)
		trigger.GET("", controller.Pipeline.ListTriggers)
		trigger.GET(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.GetTriggerByID)
		trigger.PUT(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.UpdateTrigger)
		trigger.DELETE(fmt.Sprintf("/:%s", req.PathParamTriggerID), controller.Pipeline.DeleteTrigger)
	}
}
