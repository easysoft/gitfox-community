package router

import (
	ctnHandler "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/handler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"

	"github.com/gin-gonic/gin"
)

// registerContainerRoutes 挂载 OCI Distribution Spec /v2 路由。
//
// 路径形式：
//
//	/v2/                                           — Basic 挑战
//	/v2/:domainid/:artifactname/:imagename/...
//
// :domainid 支持三种取值，由 handler 解析：
//
//	"global"           → 全局作用域
//	"space{数字ID}"     → 空间作用域（如 space12）
//	"repo{数字ID}"      → 仓库作用域（如 repo33）
func registerContainerRoutes(r *gin.Engine, handler *ctnHandler.ContainerHandler) {
	v2 := r.Group("/v2")

	// /v2/ 探测端点（Basic 挑战）
	v2.GET("/", authn.ContainerRegistryBasicAuth("gitfox_registry"), handler.HandleAPIBase)

	auth := authn.ContainerRegistryBasicAuth("gitfox_registry")

	// 统一路由组：所有作用域共用一条路径模板，domainid 由 handler 解析。
	g := v2.Group("/:domainid/:artifactname/:imagename", auth)
	registerContainerOps(g, handler)
}

// registerContainerOps 将 OCI Distribution Spec 定义的标准操作路由注册到指定分组。
// 所有作用域共用同一个路由组，这些 OCI 操作挂载一次即可。
func registerContainerOps(g *gin.RouterGroup, h *ctnHandler.ContainerHandler) {
	// 读操作
	g.HEAD("/blobs/:digest", h.HandleBlobHead)
	g.GET("/blobs/:digest", h.HandleBlobGet)
	g.GET("/blobs/uploads/:uuid", h.HandleBlobUploadStatus)
	g.HEAD("/manifests/:reference", h.HandleManifestHead)
	g.GET("/manifests/:reference", h.HandleManifestGet)
	g.GET("/tags/list", h.HandleListTag)
	// 写操作
	g.DELETE("/blobs/:digest", h.HandleBlobDelete)
	g.POST("/blobs/uploads/", h.HandleBlobUploadStart)
	g.PATCH("/blobs/uploads/:uuid", h.HandleBlobUploadPatch)
	g.PUT("/blobs/uploads/:uuid", h.HandleBlobUploadFinish)
	g.DELETE("/blobs/uploads/:uuid", h.HandleBlobUploadCancel)
	g.PUT("/manifests/:reference", h.HandleManifestPut)
	g.DELETE("/manifests/:reference", h.HandleManifestDelete)
}
