package router

import (
	"github.com/gin-gonic/gin"
)

// registerCodebaseRoutes 注册代码库相关路由
func registerCodebaseRoutes(r *gin.RouterGroup) {
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
			branchesGroup.POST("")           // 创建分支
			branchesGroup.GET("/:branch_id") // 获取分支详情
		}
	}
}
