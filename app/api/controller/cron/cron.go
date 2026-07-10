package cron

import (
	"encoding/json"
	"strconv"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	jobsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/jobs"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与定时任务相关的 HTTP 请求
type Controller struct {
	scheduler *jobsSvc.Scheduler
}

// NewController 创建定时任务处理器实例
func NewController(scheduler *jobsSvc.Scheduler) *Controller {
	return &Controller{
		scheduler: scheduler,
	}
}

// CreateJobRequest 创建/更新定时任务的请求参数
type CreateJobRequest struct {
	JobName            string `json:"jobName" binding:"required"`
	JobType            string `json:"jobType" binding:"required"`
	JobData            string `json:"jobData" binding:"required"`
	CronDef            string `json:"cronDef" binding:"required"`
	MaxDurationSeconds int64  `json:"maxDurationSeconds" binding:"required,min=1"`
}

// CreateOrUpdate 创建或更新一个定时任务
// @Summary 创建或更新定时任务
// @Description 创建或更新一个 recurring cron job（通过 upsert 实现）
// @Tags cron
// @Accept json
// @Produce json
// @Param input body CreateJobRequest true "定时任务参数"
// @Success 200 {object} resp.ObjectResponse "操作成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /cron/jobs [post]
func (ctl *Controller) CreateOrUpdate(ctx *gin.Context) {
	var reqBody CreateJobRequest
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CronBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	jobData := reqBody.JobData
	if reqBody.JobType == pipelineSvc.JobTypeGitLabPipelineCron {
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

// DeleteJobRequest 删除定时任务的请求参数
type DeleteJobRequest struct {
	JobName string `json:"jobName" binding:"required"`
}

// Delete 删除一个定时任务
// @Summary 删除定时任务
// @Description 根据 jobName 删除一个 recurring cron job
// @Tags cron
// @Accept json
// @Produce json
// @Param input body DeleteJobRequest true "删除定时任务参数"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /cron/jobs [delete]
func (ctl *Controller) Delete(ctx *gin.Context) {
	var reqBody DeleteJobRequest
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
