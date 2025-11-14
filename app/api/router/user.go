package router

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"github.com/gin-gonic/gin"
)

// registerTokenRoutes 注册令牌相关路由
func registerUserRoutes(r *gin.RouterGroup, controller *Controller) {
	tokens := r.Group("user/tokens")
	{
		tokens.GET("/", controller.User.Login)
		tokens.POST("/", controller.User.Login)

		tokens.Group(fmt.Sprintf("{%s}", req.PathParamTokenIdentifier))
		{
			tokens.DELETE("/", controller.User.Login)
		}
	}

	keys := r.Group("user/keys")
	{
		keys.GET("/", controller.User.Login)
		keys.POST("/", controller.User.Login)
	}
}
