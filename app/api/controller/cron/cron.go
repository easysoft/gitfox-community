// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package cron

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	jobModel "github.com/easysoft/gitfox-community/app/model/jobs"
	jobSvc "github.com/easysoft/gitfox-community/app/service/jobs"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与定时任务相关的 HTTP 请求
type Controller struct {
	scheduler *jobSvc.Scheduler
}

// NewController 创建定时任务处理器实例
func NewController(scheduler *jobSvc.Scheduler) *Controller {
	return &Controller{
		scheduler: scheduler,
	}
}

// CreateOrUpdate 创建或更新一个定时任务
// @Summary 创建或更新定时任务
// @Description 创建或更新一个 recurring cron job（通过 upsert 实现）
// @Tags cron
// @Accept json
// @Produce json
// @Param input body jobModel.CreateJobRequest true "定时任务参数"
// @Success 200 {object} resp.ObjectResponse "操作成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /cron/jobs [post]
func (ctl *Controller) CreateOrUpdate(ctx *gin.Context) {
	var reqBody jobModel.CreateJobRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CronBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	jobData := reqBody.JobData
	if reqBody.JobType == pipelineSvc.JobTypeGitLabPipelineCron || reqBody.JobType == pipelineSvc.JobTypeJenkinsPipelineCron {
		pipelineID, _ := strconv.ParseInt(reqBody.JobData, 10, 64)
		data, _ := json.Marshal(map[string]interface{}{
			"pipelineID": pipelineID,
		})
		jobData = string(data)
	}
	err := ctl.scheduler.AddRecurringWithData(
		ctx.Request.Context(),
		reqBody.JobName,
		reqBody.JobType,
		jobData,
		reqBody.CronDef,
		time.Duration(reqBody.MaxDurationSeconds)*time.Second,
	)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.CronAddError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.CronAddSuccess, nil)
}

// Delete 删除一个定时任务
// @Summary 删除定时任务
// @Description 根据 jobName 删除一个 recurring cron job
// @Tags cron
// @Accept json
// @Produce json
// @Param input body jobModel.DeleteJobRequest true "删除定时任务参数"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /cron/jobs [delete]
func (ctl *Controller) Delete(ctx *gin.Context) {
	var reqBody jobModel.DeleteJobRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CronBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	err := ctl.scheduler.RemoveRecurring(ctx.Request.Context(), reqBody.JobName)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.CronRemoveError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.CronRemoveSuccess, nil)
}
