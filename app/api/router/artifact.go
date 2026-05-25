package router

import (
	"github.com/gin-gonic/gin"
)

// registerArtifactRoutes 注册制品库相关路由
func registerArtifactRoutes(group *gin.RouterGroup, controller *Controller) {
	artifact := group.Group("/artifacts")

	{
		// 创建制品库
		artifact.POST("/views", controller.Artifact.CreateArtifact)

		// 创建 group
		artifact.POST("/groups", controller.Artifact.CreateGroup)

		// 上传
		artifact.POST("/upload/:format", controller.Artifact.Upload)

		// 下载
		artifact.GET("/assets/:assetID/download", controller.Artifact.DownloadAsset)

		// --- 操作 ---
		// 移动并/或重命名 package / asset / group（同一事务，通用入口）
		artifact.POST("/entities/relocate", controller.Artifact.RelocateEntity)
		// 刷新树信息
		artifact.POST("/tree/refresh", controller.Artifact.RefreshTree)

		// 删除操作（统一使用 SoftRemoveByEntity，支持级联删除）
		artifact.DELETE("/entities", controller.Artifact.SoftRemoveByEntity)

		// --- 查询 ---
		// 树信息
		artifact.POST("/nodes", controller.Artifact.ListTree)

		// asset 级操作（读）
		artifact.GET("/assets/:assetID", controller.Artifact.GetAsset)
		artifact.POST("/assets/list", controller.Artifact.ListAssetPage)

		// 回收站
		artifact.GET("/recycle", controller.Artifact.ListRecycleBin)
		artifact.POST("/recycle/restore", controller.Artifact.Restore)

		// group 级操作（读）
		// artifact.GET("/groups/:groupID/archive", controller.Artifact.ArchiveGroup)   // pullAuth

		// 统计
		// artifact.GET("/capacity/all", controller.Artifact.StatisticCapacity)                // pullAuth
		// artifact.GET("/capacity/container", controller.Artifact.StatisticContainerCapacity) // pullAuth
		// artifact.GET("/statistic", controller.Artifact.StatisticSummary)                    // pullAuth

		// 管理操作
		artifact.GET("/garbage", controller.Artifact.GarbageCollect)
		artifact.GET("/garbage/:entity", controller.Artifact.GarbageCollect)
	}
}
