package router

import (
	"github.com/gin-gonic/gin"
)

// registerTokenRoutes 注册令牌相关路由
func registerUserRoutes(r *gin.RouterGroup, controller *Controller) {
	// tokens := r.Group("user/tokens")
	// {
	// tokens.GET("/", controller.User.Login)
	// tokens.POST("/", controller.User.Login)
	// }

	keys := r.Group("user/keys")
	{
		keys.GET("/", controller.User.ListKeys)
		keys.POST("/", controller.User.AddKey)
		keys.PUT("/:id", controller.User.UpdateKey)
		keys.DELETE("/:id", controller.User.DeleteKey)
	}
}
