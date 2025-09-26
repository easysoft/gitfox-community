package router

import (
	"github.com/easysoft/dev-gitfox/internal/api/middleware"
	"github.com/easysoft/dev-gitfox/internal/di"

	"github.com/gin-gonic/gin"
)

// NewRouter 创建路由实例并注册所有业务模块的路由
func NewRouter(container *di.Container) *gin.Engine {
	// 初始化Gin引擎
	r := gin.New()

	// 注册全局中间件
	r.Use(
	// middleware.Logger(),   // 日志中间件
	// middleware.Recovery(), // 异常恢复中间件
	// middleware.CORS(),     // 跨域处理中间件
	)

	// 注册无需认证的公共路由
	registerPublicRoutes(r)

	// 创建需要认证的路由组（应用鉴权中间件）
	authGroup := r.Group("/api")
	authGroup.Use(middleware.Auth()) // 所有/api/*路由需要先通过认证

	// 按业务模块注册路由（从容器获取模块实例）
	registerCodebaseRoutes(authGroup, container) // 代码库相关路由
	// registerPipelineRoutes(authGroup, container) // 流水线相关路由
	// registerArtifactRoutes(authGroup, container) // 制品库相关路由
	// registerSecurityRoutes(authGroup, container) // 安全扫描相关路由

	return r
}

// registerPublicRoutes 注册无需认证的公共路由（如健康检查、版本信息）
func registerPublicRoutes(r *gin.Engine) {
	public := r.Group("/public")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})
		public.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": "v1.0.0"})
		})
	}
}
