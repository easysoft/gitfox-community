package router

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	requtil "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	ctnAdapter "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	ctnHandler "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/handler"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	principal "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/user"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/cache"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"

	"github.com/gin-gonic/gin"
)

// nopTokenGetter 是一个空实现的 Getter，Find 永远返回 error。
// token 只通过 TTLCache.Set 写入；缓存 miss 意味着 token 不存在或已过期。
type nopTokenGetter struct{}

func (nopTokenGetter) Find(_ context.Context, _ string) (string, error) {
	return "", errors.New("token not found")
}

// registerContainerRoutes 挂载 OCI Distribution Spec /v2 路由。
//
// 支持三种作用域前缀：
//
//	GET /v2/                                                          — Bearer 挑战
//	GET /v2/token                                                     — 颁发 token
//	/v2/global/{viewname}/{groupname}/{imagename}                     — 全局作用域
//	/v2/space/{spaceName}/{viewname}/{groupname}/{imagename}          — 空间作用域
//	/v2/repo/{spaceName}/{repoName}/{viewname}/{groupname}/{imagename} — 仓库作用域
//
// groupname 为 container namespace（必填），imagename 为镜像名。
//
// 注意：Gin 要求 catch-all 参数（*xxx）必须是路径最后一段，因此 repopath
// 改用两个固定参数 :groupname/:imagename 来表示，不再使用 catch-all。
func registerContainerRoutes(r *gin.Engine, handler *ctnHandler.ContainerHandler, tokenStore *cache.TTLCache[string, string], cfg *config.Config) {
	handler.SetTokenStore(tokenStore)
	v2 := r.Group("/v2")

	// GET /v2/ — 探测端点，匿名时返回 401 + WWW-Authenticate Bearer
	v2.GET("/", requireContainerAccess(tokenStore, cfg), handler.HandleAPIBase)

	// GET /v2/token — 颁发 token（Basic Auth 认证）
	v2.GET("/token", handler.HandleToken)

	auth := containerTokenAuth(tokenStore)

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

// requireContainerAccess 匿名时返回 Bearer 挑战，引导 docker client 去 /v2/token 换 token。
func requireContainerAccess(tokenStore *cache.TTLCache[string, string], cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !isValidBearerToken(authHeader, tokenStore) {
			tokenURL := fmt.Sprintf("%s/v2/token", cfg.URL.Base)
			c.Header("WWW-Authenticate",
				fmt.Sprintf(`Bearer realm="%s",service="gitfox_registry",scope="*"`, tokenURL))
			renderContainerAuthErr(c, "authentication required")
			return
		}
		c.Next()
	}
}

// containerTokenAuth 内存 token 校验中间件。
// 从 Authorization: Bearer <token> 中提取 token，查内存缓存。
// 命中 → 取出 account，构造 auth.Session 注入 context 后放行；未命中 → 401。
func containerTokenAuth(tokenStore *cache.TTLCache[string, string]) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			renderContainerAuthErr(c, "missing bearer token, please docker login first")
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		account, err := tokenStore.Get(c.Request.Context(), token)
		if err != nil {
			renderContainerAuthErr(c, "token expired or invalid, please docker login again")
			return
		}

		// 构造 auth.Session 并注入 context，与项目认证中间件 Attempt 保持一致
		session := &authModel.Session{
			User: &principal.User{
				Account: account,
			},
		}
		ctx := requtil.SetAuthSession(c.Request.Context(), session)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// isValidBearerToken 检查 Authorization header 中的 Bearer token 是否在内存缓存中。
func isValidBearerToken(authHeader string, tokenStore *cache.TTLCache[string, string]) bool {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := tokenStore.Get(context.Background(), token)
	return err == nil
}

// renderContainerAuthErr 渲染 OCI 规范的 401 错误。
func renderContainerAuthErr(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"errors": []*ctnAdapter.RegistryErr{
			ctnAdapter.ErrUnauthorized.WithDetail(message),
		},
	})
}
