// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

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
		branches.POST("/archive", controller.Branch.Archive)      // 批量归档分支
		branches.POST("/activate", controller.Branch.Activate)    // 批量激活归档分支
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
