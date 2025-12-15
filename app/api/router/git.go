package router

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"github.com/gin-gonic/gin"
)

// registerGitRoutes 注册 Git 相关路由
func registerGitRoutes(r *gin.RouterGroup, controller *Controller) {
	git := r.Group(fmt.Sprintf("git/:%s/:%s", req.QueryParamSpaceID, req.PathParamRepoRef))
	{
		git.POST("/git-upload-pack", controller.Git.Clone)
		git.POST("/git-receive-pack", controller.Git.Push)
		git.GET("/info/refs", controller.Git.InfoRefs)
	}
}

func registerGitHookRoutes(r *gin.RouterGroup, controller *Controller) {
	git := r.Group("/git-hooks")
	{
		git.POST("/pre-receive", controller.Git.PreReceiveClone)
		git.POST("/post-receive", controller.Git.PostReceiveClone)
		git.GET("/update", controller.Git.UpdateClone)
	}
}
