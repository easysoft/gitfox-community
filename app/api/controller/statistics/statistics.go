// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      liyang <liyang@easycorp.ltd>

package statistics

import (
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	statisticsModel "github.com/easysoft/gitfox-community/app/model/statistics"
	statisticsSvc "github.com/easysoft/gitfox-community/app/service/statistics"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理统计相关的 HTTP 请求
type Controller struct {
	statisticsSvc statisticsSvc.Interface
}

// NewController 创建统计处理器实例
func NewController(statisticsService statisticsSvc.Interface) *Controller {
	return &Controller{statisticsSvc: statisticsService}
}

// @Summary 查询代码提交统计
// @Description 遍历未删除的代码库执行git命令统计提交次数，支持按时间统计，支持通过 user 参数按人筛选，支持通过 byUser 参数返回每个提交者的提交次数
// @Tags statistics
// @Accept application/json
// @Produce application/json
// @Param beginTime query string true "统计开始日期，格式 yyyy-mm-dd" example(2026-09-15)
// @Param endTime query string true "统计结束日期，格式 yyyy-mm-dd" example(2026-09-16)
// @Param user query string false "提交人账号或邮箱，为空时统计全部用户" example(admin)
// @Param byUser query bool false "是否返回每个提交者的提交次数" example(false)
// @Success 200 {object} resp.ObjectResponse{data=statisticsModel.RespForCommitStatistics} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未认证"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /statistics/commits [get]
func (c *Controller) GetCommitStatistics(ctx *gin.Context) {
	var reqQuery statisticsModel.ReqForCommitStatistics
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.StatisticsQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	if err := reqQuery.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	respData, err := c.statisticsSvc.GetCommitStatistics(ctx.Request.Context(), reqQuery)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.StatisticsCommitSuccess, respData)
}
