package router

import (
	"github.com/gin-gonic/gin"
)

// registerSpaceRoutes 注册空间相关路由
func registerSpaceRoutes(r *gin.RouterGroup, controller *Controller) {
	spaceGroup := r.Group("/spaces")
	{

		spaceGroup.POST("/", controller.SpaceController.Create)
	}
}
