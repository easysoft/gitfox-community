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
	pullAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactPull)
	// 写操作 — 推送权限
	pushAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactPush)
	// 创建操作
	createAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactCreate)
	// 删除操作
	deleteAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactDelete)
	{

		// 创建制品库
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

		// group 级操作（读）
		artifact.GET("/groups/:groupId/assets", pullAuth, controller.Artifact.ListGroupAssets)
		artifact.GET("/groups/:groupId/archive", pullAuth, controller.Artifact.ArchiveGroup)

		// asset 级操作（读）
		artifact.GET("/assets/:assetId", pullAuth, controller.Artifact.GetAsset)
		artifact.GET("/assets/:assetId/history", pullAuth, controller.Artifact.GetBlobHistory)
		artifact.GET("/assets/:assetId/download", pullAuth, controller.Artifact.DownloadAsset)

		// 删除操作（统一使用 SoftRemoveByEntity，支持级联删除）
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

		// 权限管理 — 系统级 admin 鉴权
		// 使用 ArtifactAdminMiddleware 取代 Casbin permAuth：
		// 只要 Principal.Admin（zentao App）或 User.IsAdmin（DB 管理员）即可管理权限，
		// 避免「无 artifact_admin 无法授权 artifact_admin」的死锁问题。
		adminAuth := middleware.ArtifactAdminMiddleware()
		artifact.POST("/entities/:entityId/permissions", adminAuth, controller.Artifact.SetEntityPermission)
		artifact.GET("/entities/:entityId/permissions", adminAuth, controller.Artifact.GetEntityPermissions)
		artifact.POST("/permissions", adminAuth, controller.Artifact.SetPermissions)
		artifact.GET("/permissions", adminAuth, controller.Artifact.GetPermissions)
		artifact.GET("/members", pullAuth, controller.Artifact.GetMembers)
	}
}
