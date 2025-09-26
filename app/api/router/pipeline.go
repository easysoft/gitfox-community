package router

// 	"gitfox.zcorp.cc/git/demo/gitfox-pms/internal/api/handler/pipeline"
// 	"gitfox.zcorp.cc/git/demo/gitfox-pms/internal/di"
// 	"gitfox.zcorp.cc/git/demo/gitfox-pms/internal/service/pipeline"

// "github.com/gin-gonic/gin"

// registerPipelineRoutes 注册流水线相关路由
// func registerPipelineRoutes(r *gin.RouterGroup, container *di.Container) {
// 	// 从容器获取流水线模块实例（与代码库模块注册逻辑一致）
// 	pipelineSvc, err := container.Get("pipeline")
// 	if err != nil {
// 		panic("流水线模块初始化失败: " + err.Error())
// 	}

// 	// 类型断言
// 	pipelineService := pipelineSvc.(pipeline.Service)

// 	// 初始化处理器
// 	pipelineHandler := pipeline.NewPipelineHandler(pipelineService)

// 	// 注册流水线路由
// 	pipelineGroup := r.Group("/pipelines")
// 	{
// 		pipelineGroup.POST("", pipelineHandler.Create)                // 创建流水线
// 		pipelineGroup.GET("", pipelineHandler.List)                   // 列出流水线
// 		pipelineGroup.GET("/:pipeline_id", pipelineHandler.Get)       // 获取流水线详情
// 		pipelineGroup.POST("/:pipeline_id/run", pipelineHandler.Run)  // 运行流水线
// 		pipelineGroup.DELETE("/:pipeline_id", pipelineHandler.Delete) // 删除流水线
// 	}
// }
