// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      caiyan <caiyan@chandao.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerSpaceRoutes 注册空间相关路由
func registerSpaceRoutes(r *gin.RouterGroup, controller *Controller) {
	spaces := r.Group("/spaces")
	{
		spaces.POST("", controller.Space.Create) // 创建空间
		spaces.GET("", controller.Space.List)    // 列表查询
		spaceID := spaces.Group("/:id")
		{
			spaceID.GET("", controller.Space.GetByID)          // 单个查询
			spaceID.PUT("", controller.Space.Update)           // 更新空间
			spaceID.DELETE("", controller.Space.Delete)        // 删除空间
			spaceID.POST("/restore", controller.Space.Restore) // 恢复空间
		}
	}
}
