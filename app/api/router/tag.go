// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerTagRoutes 注册标签相关路由
func registerTagRoutes(r *gin.RouterGroup, controller *Controller) {
	tags := r.Group("/repos/:id/tags")
	{
		tags.POST("", controller.Tag.Create)         // 创建标签
		tags.POST("/list", controller.Tag.List)      // 查看标签列表
		tags.DELETE("/:name", controller.Tag.Delete) // 删除标签
	}
}
