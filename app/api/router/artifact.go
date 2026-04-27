package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/middleware"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"
	enum "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	"github.com/gin-gonic/gin"
)

// registerArtifactRoutes 注册制品库相关路由
func registerArtifactRoutes(group *gin.RouterGroup, controller *Controller, authz authorizer.Authorizer) {
	artifact := group.Group("/artifacts")

	// 读操作 — 拉取权限
	pullAuth := middleware.ArtifactAuthMiddleware(authz, enum.CasbinActionArtifactPull)
	// 写操作 — 推送权限
	pushAuth := middleware.ArtifactAuthMiddleware(authz, enum.CasbinActionArtifactPush)
	// 创建操作
	createAuth := middleware.ArtifactAuthMiddleware(authz, enum.CasbinActionArtifactCreate)
	// 删除操作
	deleteAuth := middleware.ArtifactAuthMiddleware(authz, enum.CasbinActionArtifactDelete)
	{
		// 空间级操作（读）
		artifact.POST("/views", createAuth, controller.Artifact.CreateView)
		artifact.POST("/packages", pullAuth, controller.Artifact.ListPackages)
		artifact.POST("/formats", pullAuth, controller.Artifact.ListFormats)

		// 创建 group
		artifact.POST("/groups", createAuth, controller.Artifact.CreateGroup)

		// 上传
		artifact.POST("/upload/:format", pushAuth, controller.Artifact.Upload)

		// 包版本操作（读）
		artifact.POST("/versions/list", pullAuth, controller.Artifact.ListVersions)

		// 版本级操作
		artifact.GET("/versions/:versionId/assets", pullAuth, controller.Artifact.ListAssets)
		artifact.DELETE("/versions/:versionId", deleteAuth, controller.Artifact.DeleteVersion)

		// group 级操作（读）
		artifact.GET("/groups/:groupId/assets", pullAuth, controller.Artifact.ListGroupAssets)
		artifact.GET("/groups/:groupId/archive", pullAuth, controller.Artifact.ArchiveGroup)

		// asset 级操作（读）
		artifact.GET("/assets/:assetId", pullAuth, controller.Artifact.GetAsset)
		artifact.GET("/assets/:assetId/history", pullAuth, controller.Artifact.GetBlobHistory)
		artifact.GET("/assets/:assetId/download", pullAuth, controller.Artifact.DownloadAsset)

		// 包级删除
		artifact.DELETE("/packages/:packageId", deleteAuth, controller.Artifact.DeletePackage)
		artifact.DELETE("/entities", deleteAuth, controller.Artifact.SoftRemoveByEntity)

		// 树节点（读）
		artifact.GET("/nodes/:nodeId/children", pullAuth, controller.Artifact.ListTreeNodes)

		// 统计
		artifact.GET("/capacity/all", pullAuth, controller.Artifact.StatisticCapacity)
		artifact.GET("/capacity/container", pullAuth, controller.Artifact.StatisticContainerCapacity)
		artifact.GET("/statistic", pullAuth, controller.Artifact.StatisticSummary)

		// 管理操作
		artifact.GET("/garbage/container", deleteAuth, controller.Artifact.GarbageCollectContainer)
		artifact.GET("/garbage/soft-remove", deleteAuth, controller.Artifact.GarbageCollectSoftRemove)
		artifact.POST("/tree/refresh", createAuth, controller.Artifact.RefreshTree)

		// 权限管理
		permAuth := middleware.ArtifactAuthMiddleware(authz, enum.CasbinActionPermission)
		artifact.POST("/entities/:entityId/permissions", permAuth, controller.Artifact.SetEntityPermission)
		artifact.GET("/entities/:entityId/permissions", permAuth, controller.Artifact.GetEntityPermissions)
		artifact.POST("/permissions", permAuth, controller.Artifact.SetPermissions)
		artifact.GET("/permissions", permAuth, controller.Artifact.GetPermissions)
		artifact.GET("/members", pullAuth, controller.Artifact.GetMembers)
	}
}
