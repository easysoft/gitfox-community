package router

import (
	"github.com/gin-gonic/gin"
)

// registerCommitRoutes 注册提交相关路由
func registerCommitRoutes(r *gin.RouterGroup, controller *Controller) {
	commits := r.Group("/commits")
	{
		commits.POST("", controller.Commit.Create)                   // 创建提交
		commits.POST("/list", controller.Commit.List)                // 查看提交列表
		commits.GET("", controller.Commit.Get)                       // 查看单个提交信息
		commits.POST("/divergence", controller.Commit.GetDivergence) // 查看提交分歧度
		commits.POST("/diff", controller.Commit.GetDiff)             // 查看提交 diff
	}
}
