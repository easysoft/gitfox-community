package router

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"github.com/gin-gonic/gin"
)

// registerSpaceRoutes 注册空间相关路由
func registerRepoRoutes(r *gin.RouterGroup, controller *Controller) {
	repos := r.Group("/repos")
	{
		repos.POST("/", controller.Repo.Create)
		reposWithID := repos.Group(":id")
		{
			reposWithID.GET("/", controller.Repo.Get)
			reposWithID.PUT("/", controller.Repo.Update)
			reposWithID.POST("/diff", controller.Repo.GetDiff)
			reposWithID.GET("/paths", controller.Repo.GetPaths)
			reposWithID.POST("/paths/details", controller.Repo.GetPathsDetails)
			reposWithID.GET(fmt.Sprintf("/content/{%s}", req.PathParamFilePath), controller.Repo.GetContent)
			reposWithID.GET(fmt.Sprintf("/blame/{%s}", req.PathParamFilePath), controller.Repo.Blame)
		}
	}
}
