package router

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"github.com/gin-gonic/gin"
)

// registerRepoRoutes 注册仓库相关路由
func registerRepoRoutes(r *gin.RouterGroup, controller *Controller) {
	repos := r.Group("/repos")
	{
		repos.POST("/list", controller.Repo.List)
		repos.POST("/", controller.Repo.Create)
		reposWithID := repos.Group(":id")
		{
			reposWithID.GET("/", controller.Repo.Get)
			reposWithID.PUT("/", controller.Repo.Update)
			reposWithID.DELETE("/", controller.Repo.Delete)
			reposWithID.POST("/diff", controller.Repo.GetDiff)
			reposWithID.GET("/paths", controller.Repo.GetPaths)
			reposWithID.POST("/paths/details", controller.Repo.GetPathsDetails)
			reposWithID.GET(fmt.Sprintf("/content/*%s", req.PathParamFilePath), controller.Repo.GetContent)
			reposWithID.GET(fmt.Sprintf("/blame/*%s", req.PathParamFilePath), controller.Repo.Blame)

			// pullrequest 相关路由
			reposWithID.GET(fmt.Sprintf("/diff-stats/*%s", req.SourcePathToTargetPath), controller.Repo.GetDiffStats)
			reposWithID.POST(fmt.Sprintf("/merge-check/*%s", req.SourcePathToTargetPath), controller.Repo.MergeCheck)
			repopull := reposWithID.Group("/pullreq")
			{
				repopull.GET(fmt.Sprintf("/:%s/metadata", req.PullrequestID), controller.Repo.Metadata)
				repopull.GET(fmt.Sprintf("/:%s/diff", req.PullrequestID), controller.Repo.DiffSimp)
				repopull.POST(fmt.Sprintf("/:%s/diff", req.PullrequestID), controller.Repo.Diff)
				repopull.GET(fmt.Sprintf("/:%s/commits", req.PullrequestID), controller.Repo.Commits)
				repopull.POST(fmt.Sprintf("/:%s/merge", req.PullrequestID), controller.Repo.Merge)
			}
		}
	}
}
