// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerArtifactRoutes 注册制品库相关路由
func registerArtifactRoutes(group *gin.RouterGroup, controller *Controller) {
	artifact := group.Group("/artifacts")
	{
		// --- 制品库的相关操作 ---
		artifact.POST("/views", controller.Artifact.CreateArtifact)
		artifact.POST("/groups", controller.Artifact.CreateGroup)
		artifact.POST("/upload/:format", controller.Artifact.Upload)

		// --- 目录结构操作 ---
		artifact.POST("/nodes", controller.Artifact.ListTree)
		artifact.POST("/tree/refresh", controller.Artifact.RefreshTree)

		// --- 资产操作 ---
		artifact.GET("/assets/:assetID", controller.Artifact.GetAsset)
		artifact.GET("/assets/:assetID/download", controller.Artifact.DownloadAsset)
		artifact.POST("/assets/list", controller.Artifact.ListAssetPage)

		// --- 目录和资产相关重定向操作 ---
		artifact.POST("/entities/relocate", controller.Artifact.RelocateEntity)
		artifact.DELETE("/entities", controller.Artifact.SoftRemoveByEntity)

		// --- 回收站操作 ---
		artifact.GET("/recycle", controller.Artifact.ListRecycleBin)
		artifact.POST("/recycle/restore", controller.Artifact.Restore)
		artifact.GET("/garbage", controller.Artifact.GarbageCollect)

		// --- group 归档操作 ---
		// artifact.GET("/groups/:groupID/archive", controller.Artifact.ArchiveGroup)   // pullAuth
	}
}
