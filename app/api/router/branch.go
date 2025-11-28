package router

import (
	"github.com/gin-gonic/gin"
)

// registerBranchRoutes 注册分支相关路由
func registerBranchRoutes(r *gin.RouterGroup, controller *Controller) {
	branches := r.Group("/repos/:id/branches")
	{
		branches.POST("", controller.Branch.Create)               // 创建分支
		branches.POST("/list", controller.Branch.List)            // 查看分支列表
		branches.GET("/prefixes", controller.Branch.ListPrefixes) // 查看分支前缀列表
	}
}
