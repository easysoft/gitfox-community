// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

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
		keys.GET("/:id", controller.User.GetKey)
		keys.POST("/", controller.User.AddKey)
		keys.PUT("/:id", controller.User.UpdateKey)
		keys.DELETE("/:id", controller.User.DeleteKey)
	}
}
