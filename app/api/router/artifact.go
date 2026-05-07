package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/authorizer"

	"github.com/gin-gonic/gin"
)

// registerArtifactRoutes 注册制品库相关路由
func registerArtifactRoutes(group *gin.RouterGroup, controller *Controller, _ authorizer.Authorizer) {
	artifact := group.Group("/artifacts")
	// 读操作 — 拉取权限
	//pullAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactPull)
	// 写操作 — 推送权限
	//pushAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactPush)
	// 创建操作
	//createAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactCreate)
	// 删除操作
	//deleteAuth := middleware.ArtifactAuthMiddleware(authz, controller.Artifact.Service(), enum.CasbinActionArtifactDelete)

	{
		// 创建制品库
		artifact.POST("/views", controller.Artifact.CreateView) // createAuth

		artifact.POST("/packages", controller.Artifact.ListPackages) // pullAuth
		artifact.POST("/formats", controller.Artifact.ListFormats)   // pullAuth
		artifact.POST("/nodes", controller.Artifact.ListTree)        // pullAuth

		// 创建 group
		artifact.POST("/groups", controller.Artifact.CreateGroup) // createAuth

		// 上传
		artifact.POST("/upload/:format", controller.Artifact.Upload) // pushAuth

		// 包版本操作（读）
		artifact.POST("/versions/list", controller.Artifact.ListVersions) // pullAuth

		// 版本级操作
		artifact.GET("/versions/:versionId/assets", controller.Artifact.ListAssets) // pullAuth

		// group 级操作（读）
		artifact.GET("/groups/:groupId/assets", controller.Artifact.ListGroupAssets) // pullAuth
		artifact.GET("/groups/:groupId/archive", controller.Artifact.ArchiveGroup)   // pullAuth

		// asset 级操作（读）
		artifact.GET("/assets/:assetId", controller.Artifact.GetAsset)               // pullAuth
		artifact.GET("/assets/:assetId/history", controller.Artifact.GetBlobHistory) // pullAuth
		artifact.GET("/assets/:assetId/download", controller.Artifact.DownloadAsset) // pullAuth

		// 删除操作（统一使用 SoftRemoveByEntity，支持级联删除）
		artifact.DELETE("/entities", controller.Artifact.SoftRemoveByEntity) // deleteAuth

		// 树节点（读）
		artifact.GET("/nodes/:nodeId/children", controller.Artifact.ListTreeNodes) // pullAuth

		// 统计
		artifact.GET("/capacity/all", controller.Artifact.StatisticCapacity)                // pullAuth
		artifact.GET("/capacity/container", controller.Artifact.StatisticContainerCapacity) // pullAuth
		artifact.GET("/statistic", controller.Artifact.StatisticSummary)                    // pullAuth

		// 管理操作
		artifact.GET("/garbage/container", controller.Artifact.GarbageCollectContainer)    // deleteAuth
		artifact.GET("/garbage/soft-remove", controller.Artifact.GarbageCollectSoftRemove) // deleteAuth
		artifact.POST("/tree/refresh", controller.Artifact.RefreshTree)                    // createAuth

		// 权限管理 — 系统级 admin 鉴权
		// 使用 ArtifactAdminMiddleware 取代 Casbin permAuth：
		// 只要 Principal.Admin（zentao App）或 User.IsAdmin（DB 管理员）即可管理权限，
		// 避免「无 artifact_admin 无法授权 artifact_admin」的死锁问题。
		// adminAuth := middleware.ArtifactAdminMiddleware()
		// artifact.POST("/entities/:entityId/permissions", adminAuth, controller.Artifact.SetEntityPermission)
		// artifact.GET("/entities/:entityId/permissions", adminAuth, controller.Artifact.GetEntityPermissions)
		// artifact.POST("/permissions", adminAuth, controller.Artifact.SetPermissions)
		// artifact.GET("/permissions", adminAuth, controller.Artifact.GetPermissions)
		// artifact.GET("/members", pullAuth, controller.Artifact.GetMembers)
	}
}
