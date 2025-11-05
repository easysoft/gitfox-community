package router

import (
	"github.com/gin-gonic/gin"
)

// registerBranchRoutes 注册分支相关路由
func registerBranchRoutes(r *gin.RouterGroup, controller *Controller) {
	branches := r.Group("/branches")
	{
		branches.POST("", controller.Branch.Create) // 创建分支
	}
}
