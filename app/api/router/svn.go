// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerSvnRoutes 注册 SVN 相关路由
func registerSvnRoutes(r *gin.RouterGroup, controller *Controller) {
	svn := r.Group("/repos/:id/svn")
	{
		svn.POST("/exec", controller.Svn.Exec) // 统一执行入口
		svn.GET("/export", controller.Svn.Export)
	}
}
