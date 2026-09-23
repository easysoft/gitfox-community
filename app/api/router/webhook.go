// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerWebhookRoutes 注册webhook相关路由
func registerWebhookRoutes(r *gin.RouterGroup, controller *Controller) {
	webhooks := r.Group("/repos/:id/webhooks")
	{
		webhooks.GET("/", controller.Webhook.List)
		webhooks.POST("/", controller.Webhook.Create)
		webhooksWithID := webhooks.Group("/:webhookID")
		{
			webhooksWithID.GET("/", controller.Webhook.Get)
			webhooksWithID.PUT("/", controller.Webhook.Update)
			webhooksWithID.DELETE("/", controller.Webhook.Delete)
			executions := webhooksWithID.Group("/executions")
			{
				executions.GET("/", controller.Webhook.ListExecutions)
				executions.GET("/:executionID", controller.Webhook.GetExecution)
				executions.POST("/:executionID/retrigger", controller.Webhook.RetriggerExecution)
			}
		}
	}
}
