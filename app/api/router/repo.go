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
			reposWithID.POST("/diff", controller.Repo.GetDiff)
			reposWithID.GET("/paths", controller.Repo.Paths)
			reposWithID.POST("/paths/details", controller.Repo.PathsDetails)
		}
		reposWithRepoID := repos.Group(fmt.Sprintf("{%s}", req.QueryParamRepoID))
		{
			reposWithRepoID.GET("/", controller.Repo.Get)
			reposWithRepoID.GET("/content", controller.Repo.GetContent)
			reposWithRepoID.GET(fmt.Sprintf("/blame/{%s}", req.PathParamFilePath), controller.Repo.Blame)
		}

	}
}
