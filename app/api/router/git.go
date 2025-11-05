package router

import (
	"github.com/gin-gonic/gin"
)

// registerGitRoutes 注册 Git 相关路由
func registerGitRoutes(r *gin.RouterGroup, controller *Controller) {
	git := r.Group("/git")
	{
		git.POST("/git-upload-pack", controller.Git.Clone)
		git.POST("/git-receive-pack", controller.Git.Push)
		git.GET("/info/refs", controller.Git.InfoRefs)
	}
}
