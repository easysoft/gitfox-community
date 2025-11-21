package router

import (
	"github.com/gin-gonic/gin"
)

// registerTagRoutes 注册标签相关路由
func registerTagRoutes(r *gin.RouterGroup, controller *Controller) {
	tags := r.Group("/tags")
	{
		tags.POST("", controller.Tag.Create)         // 创建标签
		tags.POST("/list", controller.Tag.List)      // 查看标签列表
		tags.DELETE("/:name", controller.Tag.Delete) // 删除标签
	}
}
