// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
	// 启动阶段
	runner.POST("/ping", controller.Runner.Ping)
	runner.POST("/register", controller.Runner.Register)
	runner.POST("/keep-alive", controller.Runner.UpdateHeartBeat)

	// 拉取执行
	runner.POST("/stage", controller.Runner.RequestStage)
	runner.POST("/stage/:stageID", controller.Runner.AcceptStage)
	runner.GET("/stage/:stageID", controller.Runner.DetailStage)
	runner.PUT("/stage/:stageID", controller.Runner.UpdateStage)
	runner.PUT("/step/:stepID", controller.Runner.UpdateStep)
	runner.POST("/build/:buildID/watch", controller.Runner.WatchBuild)
	runner.POST("/step/:stepID/logs/batch", controller.Runner.BatchLogs)
	runner.POST("/step/:stepID/logs/upload", controller.Runner.UploadLogs)
}
