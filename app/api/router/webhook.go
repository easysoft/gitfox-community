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
