// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package router

import (
	"github.com/gin-gonic/gin"
)

// registerCommitRoutes 注册提交相关路由
func registerCommitRoutes(r *gin.RouterGroup, controller *Controller) {
	commits := r.Group("/repos/:id/commits")
	{
		commits.POST("", controller.Commit.Create)                   // 创建提交
		commits.POST("/list", controller.Commit.List)                // 查看提交列表
		commits.GET("", controller.Commit.Get)                       // 查看单个提交信息
		commits.POST("/divergence", controller.Commit.GetDivergence) // 查看提交分歧度
		commits.POST("/diff", controller.Commit.GetDiff)             // 查看提交 diff
	}
}
