package router

import (
	"github.com/gin-gonic/gin"
)

// registerFitRoutes 注册 fit 相关路由（Basic 认证，直连外部系统）
func registerFitRoutes(r *gin.RouterGroup, controller *Controller) {
	fit := r.Group("/fit")
	{
		fit.GET("/user", controller.User.Current)

		repos := fit.Group("/repos")
		{
			reposWithID := repos.Group(":id")
			{
				reposWithID.GET("/", controller.Repo.Get)
				reposWithID.GET("/settings/advanced", controller.Fit.GetBranchPermission)
			}
		}

		spaces := fit.Group("/spaces")
		{
			spaces.GET("/:code", controller.Fit.GetRepoByName)
		}
	}
}
