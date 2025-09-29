package router

import (
	"gitfox.zcorp.cc/git/demo/gitfox-pms/app/api/handler/codebase"

	"github.com/gin-gonic/gin"
)

// registerCodebaseRoutes 注册代码库相关路由
func registerCodebaseRoutes(r *gin.RouterGroup, container *ServiceContainer) {
	// 1. 从容器获取代码库相关模块实例
	
	branchSvc := container.BranchService


	// 3. 初始化处理器
	branchHandler := codebase.NewBranchHandler(branchSvc)

	// 4. 注册代码库路由
	codebaseGroup := r.Group("/repos")
	{
		// 代码库仓库相关接口
		// codebaseGroup.POST("", repoHandler.Create)            // 创建代码库
		// codebaseGroup.GET("", repoHandler.List)               // 列出代码库
		// codebaseGroup.GET("/:repo_id", repoHandler.Get)       // 获取代码库详情
		// codebaseGroup.PUT("/:repo_id", repoHandler.Update)    // 更新代码库
		// codebaseGroup.DELETE("/:repo_id", repoHandler.Delete) // 删除代码库

		// 代码库分支相关接口（嵌套在代码库路由下）
		branchesGroup := codebaseGroup.Group("/:repo_id/branches")
		{
			branchesGroup.POST("", branchHandler.Create)        // 创建分支
			branchesGroup.GET("/:branch_id", branchHandler.Get) // 获取分支详情
		}
	}
}
