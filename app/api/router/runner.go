package router

import (
	"github.com/gin-gonic/gin"
)

// registerRunnerRoutes 注册 Runner 管理相关路由
func registerRunnerRoutes(r *gin.RouterGroup, controller *Controller) {
	runner := r.Group("/runners")
	{
		runner.GET("/token", controller.Runner.GenerateToken)
		runner.POST("/list", controller.Runner.List)
		runnerWithID := runner.Group(":id")
		{
			runnerWithID.GET("/", controller.Runner.GetByID)
			runnerWithID.PUT("/", controller.Runner.UpdateBasicInfo)
			runnerWithID.PUT("/status", controller.Runner.UpdateStatus)
			runnerWithID.DELETE("/", controller.Runner.Delete)
		}
	}
}

// registerRunnerClientRoutes 注册 Runner 客户端相关路由
func registerRunnerClientRoutes(runner *gin.RouterGroup, controller *Controller) {
	runner.POST("/register", controller.Runner.Register)
	runner.POST("/keep-alive", controller.Runner.UpdateHeartBeat)
}
