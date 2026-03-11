package router

import (
	"fmt"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/middleware"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/auth/authn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// NewRouter 创建路由实例并注册所有业务模块的路由
func NewRouter(mode string, logger *logger.SugaredMateLogger, controller *Controller) *gin.Engine {

	// 根据 mode 设置 Gin 运行模式
	switch mode {
	case "prod", "production", "release":
		gin.SetMode(gin.ReleaseMode)
	default: // dev, debug, local 等都使用 debug 模式
		gin.SetMode(gin.DebugMode)
	}

	// 初始化Gin引擎
	r := gin.New()

	// 注册全局中间件
	r.Use(
		middleware.RequestID(logger),                    // 生成trace_id并透传，同时注入logger到上下文
		middleware.AccessLog(logger),                    // 统一访问日志输出
		middleware.CORS(middleware.DefaultCORSConfig()), // 跨域处理中间件
		i18n.GinMiddleware(nil),                         // 国际化中间件 使用初始化配置
	)

	// debug 模式下开启 API 文档和性能分析
	if gin.Mode() != gin.ReleaseMode {
		// 注册 pprof 性能分析路由
		pprof.Register(r)
	}
	// 注册 Swagger API 文档路由（使用独立路径避免与业务路由冲突）
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 注册无需认证的公共路由
	registerPublicRoutes(r, controller)
	// 注册无需认证的特殊路由
	registerNoAuthRoutes(r, controller)

	// 创建需要 JWT 认证的路由组
	authGroup := r.Group("/api/v2")
	authGroup.Use(middleware.Attempt(authn.Get(authn.AuthKindEasy)))
	registerSpaceRoutes(authGroup, controller)    // 空间相关路由
	registerBranchRoutes(authGroup, controller)   // 分支相关路由
	registerTagRoutes(authGroup, controller)      // 标签相关路由
	registerCommitRoutes(authGroup, controller)   // 提交相关路由
	registerRepoRoutes(authGroup, controller)     // 仓库相关路由
	registerScanRoutes(authGroup, controller)     // 扫描相关路由
	registerPipelineRoutes(authGroup, controller) // 流水线相关路由
	registerRunnerRoutes(authGroup, controller)   // Runner 管理相关路由
	registerUserRoutes(authGroup, controller)     // 用户相关路由

	// 创建需要 Basic 认证的路由组（用于 Git 操作）
	basicAuthGroup := r.Group("")
	basicAuthGroup.Use(middleware.Attempt(authn.Get(authn.AuthKindBasic)))
	registerGitRoutes(basicAuthGroup, controller) // Git 相关路由

	// 注册 Git 钩子路由 内部使用无需鉴权
	internalGroup := r.Group("/api/v2/internal")
	registerGitHookRoutes(internalGroup, controller)

	// 注册 Runner client 相关路由使用 token 鉴权
	runnerClientGroup := r.Group("/api/v2/runner-client")
	registerRunnerClientRoutes(runnerClientGroup, controller)

	return r
}

// registerPublicRoutes 注册无需认证的公共路由（如健康检查、版本信息）
func registerPublicRoutes(r *gin.Engine, controller *Controller) {
	public := r.Group("/public")
	{

		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "healthy"})
		})
		public.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": "v1.0.0"})
		})

		public.GET("/test", controller.Test.Test)
	}
}

// registerNoAuthRoutes 注册无需认证的特殊路由
func registerNoAuthRoutes(r *gin.Engine, controller *Controller) {
	r.GET(fmt.Sprintf("/api/v2/na/repos/:id/archive/*%s", req.PathParamFilePath), controller.Repo.Archive)

	// 流水线日志流接口（无需鉴权，用于外部系统访问）
	r.GET(fmt.Sprintf("/api/v2/pipelines/:%s/executions/:%s/stream/:%s/:%s",
		req.PathParamPipelineID, req.PathParamExecutionID, req.PathParamStageID, req.PathParamStepID),
		controller.Execution.GetLiveLogs)
	r.POST(fmt.Sprintf("/api/v2/pipelines/:%s/executions/:%s/stream/:%s/:%s",
		req.PathParamPipelineID, req.PathParamExecutionID, req.PathParamStageID, req.PathParamStepID),
		controller.Execution.CreateLiveLogs)

	// 空间事件接口（无需鉴权，用于 SSE 推送）
	r.GET("/api/v2/spaces/:id/events", controller.Space.Events)
}
