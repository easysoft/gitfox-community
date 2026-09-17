package router

import (
	"github.com/gin-gonic/gin"
)

// registerStatisticsRoutes 注册统计相关路由
func registerStatisticsRoutes(r *gin.RouterGroup, controller *Controller) {
	statistics := r.Group("/statistics")
	{
		statistics.GET("/commits", controller.Statistics.GetCommitStatistics) // 遍历代码库执行git命令统计提交次数
	}
}
