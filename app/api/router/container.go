package router

import (
	ctnHandler "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/handler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"

	"github.com/gin-gonic/gin"
)

// registerContainerRoutes 挂载 OCI Distribution Spec /v2 路由。
//
// 支持三种作用域前缀：
//
//	GET /v2/                                                          — Basic 挑战
//	/v2/global/{viewname}/{groupname}/{imagename}                     — 全局作用域
//	/v2/space/{spaceName}/{viewname}/{groupname}/{imagename}          — 空间作用域
//	/v2/repo/{spaceName}/{repoName}/{viewname}/{groupname}/{imagename} — 仓库作用域
//
// groupname 为 container namespace（必填），imagename 为镜像名。
//
// 注意：Gin 要求 catch-all 参数（*xxx）必须是路径最后一段，因此 repopath
// 改用两个固定参数 :groupname/:imagename 来表示，不再使用 catch-all。
func registerContainerRoutes(r *gin.Engine, handler *ctnHandler.ContainerHandler) {
	v2 := r.Group("/v2")

	// GET /v2/ — 探测端点，匿名时返回 401 + WWW-Authenticate Basic
	v2.GET("/", authn.ContainerRegistryBasicAuth("gitfox_registry"), handler.HandleAPIBase)

	auth := authn.ContainerRegistryBasicAuth("gitfox_registry")

	// old:
	// global/{viewname}/{groupname}/{imagename} - 全局作用域 eg: /v2/global/demoview/demogroup/pipeline-helper
	// space/{spaceName}/{viewname}/{groupname}/{imagename} - 空间作用域 eg: /v2/space/demospace/demoview/demogroup/pipeline-helper
	// repo/{spaceName}/{repoName}/{viewname}/{groupname}/{imagename} - 仓库作用域 eg: /v2/repo/demospace/demorepo/demoview/demogroup/pipeline-helper
	// new:
	// global/{viewname}/{imagename} - 全局作用域 eg: /v2/global/demoview/pipeline-helper
	// space+{spaceID}/{viewname}/{imagename} - 空间作用域 eg: /v2/space12/demoview/pipeline-helper
	// repo+{spaceID}/{viewname}/{imagename} - 仓库作用域 eg: /v2/repo33/demoview/pipeline-helper

	// 全局作用域：/v2/global/{viewname}/{groupname}/{imagename}
	global := v2.Group("/global/:viewname/:groupname/:imagename", auth)
	// 空间作用域：/v2/space/{spaceName}/{viewname}/{groupname}/{imagename}
	space := v2.Group("/space/:spaceName/:viewname/:groupname/:imagename", auth)
	// 仓库作用域：/v2/repo/{spaceName}/{repoName}/{viewname}/{groupname}/{imagename}
	repo := v2.Group("/repo/:spaceName/:repoName/:viewname/:groupname/:imagename", auth)

	for _, g := range []*gin.RouterGroup{global, space, repo} {
		registerContainerOps(g, handler)
	}
}

// registerContainerOps 将 OCI Distribution Spec 定义的标准操作路由注册到指定分组。
// 供三个作用域（global/space/repo）共用，避免重复注册。
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
