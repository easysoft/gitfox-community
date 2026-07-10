package router

import (
	"github.com/gin-gonic/gin"
)

// registerCronRoutes 注册定时任务相关路由
func registerCronRoutes(r *gin.RouterGroup, controller *Controller) {
	cron := r.Group("/cron/jobs")
	{
		cron.POST("/", controller.Cron.CreateOrUpdate)
		cron.DELETE("/", controller.Cron.Delete)
	}
}
