package router

import (
	"github.com/gin-gonic/gin"
)

// registerBranchRoutes 注册分支相关路由
func registerBranchRoutes(r *gin.RouterGroup, controller *Controller) {
	branches := r.Group("/repos/:id/branches")
	{
		branches.POST("", controller.Branch.Create)               // 创建分支
		branches.POST("/list", controller.Branch.List)            // 查看分支列表
		branches.GET("/prefixes", controller.Branch.ListPrefixes) // 查看分支前缀列表
		branches.DELETE("", controller.Branch.Delete)             // 删除分支
	}

	// 分支类型路由
	branchTypes := r.Group("/repos/:id/branch-types")
	{
		branchTypes.POST("", controller.Branch.CreateBranchType)           // 创建分支类型
		branchTypes.POST("/list", controller.Branch.ListBranchTypes)       // 查询分支类型列表
		branchTypes.GET("/:typeid", controller.Branch.GetBranchType)       // 获取分支类型详情
		branchTypes.PUT("/:typeid", controller.Branch.UpdateBranchType)    // 更新分支类型
		branchTypes.DELETE("/:typeid", controller.Branch.DeleteBranchType) // 删除分支类型
	}
}
