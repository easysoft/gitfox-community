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
		repos.Group(fmt.Sprintf("{%s}", req.QueryParamRepoID))
		{
			repos.GET(fmt.Sprintf("/content/{%s}", req.PathParamOther), controller.Repo.GetContent)
		}
	}
}
