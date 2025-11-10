package router

import (
	"github.com/gin-gonic/gin"
)

// registerSpaceRoutes 注册空间相关路由
func registerRepoRoutes(r *gin.RouterGroup, controller *Controller) {
	repos := r.Group("/repos")
	{
		repos.POST("/", controller.Repo.Create)
	}
}
