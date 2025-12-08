package router

import (
	"github.com/gin-gonic/gin"
)

// registerSpaceRoutes 注册空间相关路由
func registerSpaceRoutes(r *gin.RouterGroup, controller *Controller) {
	spaces := r.Group("/spaces")
	{
		spaces.POST("", controller.Space.Create)            // 创建空间
		spaces.GET("", controller.Space.List)               // 列表查询
		spaces.GET("/:id", controller.Space.GetByID)        // 单个查询
		spaces.PUT("/:id", controller.Space.Update)         // 更新空间
		spaces.DELETE("/:id", controller.Space.Delete)      // 删除空间
		spaces.POST("/:id/restore", controller.Space.Undelete) // 恢复空间
	}
}
