// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      liyang <liyang@easycorp.ltd>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerStatisticsRoutes 注册统计相关路由
func registerStatisticsRoutes(r *gin.RouterGroup, controller *Controller) {
	statistics := r.Group("/statistics")
	{
		statistics.GET("/commits", controller.Statistics.GetCommitStatistics) // 遍历代码库执行git命令统计提交次数
	}
}
