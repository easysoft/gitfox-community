package pipeline

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// CreateTrigger 创建流水线触发
// @Summary 创建流水线触发
// @Description 创建流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param input body pipelineModel.Trigger true "创建触发参数"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/triggers [post]
func (c *Controller) CreateTrigger(ctx *gin.Context) {
	var params pipelineModel.Trigger
	if err := ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.pipelineSvc.CreateTrigger(ctx.Request.Context(), &params); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerCreateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerCreateSuccess, params)
}

// ListTriggers 查询流水线触发列表
// @Summary 查询流水线触发列表
// @Description 查询流水线触发配置列表
// @Tags pipelines
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param pipelineID query int false "流水线ID"
// @Success 200 {object} resp.ListResponse{data=[]pipelineModel.Trigger,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /pipeline/triggers [get]
func (c *Controller) ListTriggers(ctx *gin.Context) {
	var filter pipelineModel.TriggerFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	result, err := c.pipelineSvc.ListTriggers(ctx.Request.Context(), &filter)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, result.Data, resp.ListArgs{
		Total:    result.Total,
		Page:     int64(result.Page),
		PageSize: int64(result.PageSize),
	}, i18n.PipelineTriggerListSuccess, nil)
}

// GetTriggerByID 查询流水线触发配置
// @Summary 根据ID获取触发配置
// @Description 根据触发ID获取流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path int true "触发ID"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/triggers/{id} [get]
func (c *Controller) GetTriggerByID(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveInt64(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	trigger, err := c.pipelineSvc.GetTriggerByID(ctx.Request.Context(), id)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerGetError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerGetSuccess, trigger)
}

// GetTriggerByPipelineID 查询流水线触发配置
// @Summary 根据流水线ID获取触发配置
// @Description 根据流水线ID获取触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/triggers/by-pipeline/{pipelineID} [get]
func (c *Controller) GetTriggerByPipelineID(ctx *gin.Context) {
	pipelineID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	trigger, err := c.pipelineSvc.GetTriggerByPipelineID(ctx.Request.Context(), pipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerGetError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerGetSuccess, trigger)
}

// DeleteTrigger 删除流水线触发配置
// @Summary 删除触发配置
// @Description 根据触发ID删除流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param id path int true "触发ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/triggers/{id} [delete]
func (c *Controller) DeleteTrigger(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveInt64(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.pipelineSvc.DeleteTrigger(ctx.Request.Context(), id); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerDeleteError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerDeleteSuccess, nil)
}
