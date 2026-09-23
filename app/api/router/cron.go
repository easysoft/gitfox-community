// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerCronRoutes 注册定时任务相关路由
func registerCronRoutes(r *gin.RouterGroup, controller *Controller) {
	cron := r.Group("/cron/jobs")
	{
		cron.POST("/", controller.Cron.CreateOrUpdate)
		cron.DELETE("/", controller.Cron.Delete)
	}
}
