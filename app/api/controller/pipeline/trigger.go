package pipeline

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// CreateTrigger 创建流水线触发
// @Summary 创建流水线触发
// @Description 创建流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Param input body pipelineModel.Trigger true "创建触发参数"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/triggers [post]
func (c *Controller) CreateTrigger(ctx *gin.Context) {
	pipelineID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var params pipelineModel.Trigger
	if err := ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 设置 pipelineID
	params.PipelineID = pipelineID

	// 校验参数
	if err := params.Validate(); err != nil {
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
// @Description 查询指定流水线的触发配置列表
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Param repoID query string false "仓库ID"
// @Success 200 {object} resp.ListResponse{data=[]pipelineModel.Trigger,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /pipelines/{pipelineID}/triggers [get]
func (c *Controller) ListTriggers(ctx *gin.Context) {
	pipelineID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var filter pipelineModel.TriggerFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	req.SetDefaults(&filter)
	filter.PipelineID = pipelineID

	triggers, err := c.pipelineSvc.GetTriggers(ctx.Request.Context(), &filter)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerListSuccess, triggers)
}

// GetTriggerByID 查询流水线触发配置
// @Summary 根据ID获取触发配置
// @Description 根据触发ID获取流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Param triggerID path int true "触发ID"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/triggers/{triggerID} [get]
func (c *Controller) GetTriggerByID(ctx *gin.Context) {
	_, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	triggerID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamTriggerID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	trigger, err := c.pipelineSvc.GetTriggerByID(ctx.Request.Context(), triggerID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerGetError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerGetSuccess, trigger)
}

// UpdateTrigger 更新流水线触发配置
// @Summary 更新触发配置
// @Description 根据触发ID更新流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Param triggerID path int true "触发ID"
// @Param input body pipelineModel.Trigger true "更新触发参数"
// @Success 200 {object} resp.ObjectResponse{data=pipelineModel.Trigger} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/triggers/{triggerID} [put]
func (c *Controller) UpdateTrigger(ctx *gin.Context) {
	pipelineID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	triggerID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamTriggerID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var params pipelineModel.Trigger
	if err := ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 校验参数
	if err := params.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	// 设置 ID 和 pipelineID
	params.ID = triggerID
	params.PipelineID = pipelineID

	if err := c.pipelineSvc.UpdateTrigger(ctx.Request.Context(), &params); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerUpdateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerUpdateSuccess, params)
}

// DeleteTrigger 删除流水线触发配置
// @Summary 删除触发配置
// @Description 根据触发ID删除流水线触发配置
// @Tags pipelines
// @Accept json
// @Produce json
// @Param pipelineID path int true "流水线ID"
// @Param triggerID path int true "触发ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipelines/{pipelineID}/triggers/{triggerID} [delete]
func (c *Controller) DeleteTrigger(ctx *gin.Context) {
	_, err := req.PathParamAsPositiveInt64(ctx, req.PathParamPipelineID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	triggerID, err := req.PathParamAsPositiveInt64(ctx, req.PathParamTriggerID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.PipelinePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.pipelineSvc.DeleteTrigger(ctx.Request.Context(), triggerID); err != nil {
		err = errors.WithI18nKey(err, i18n.PipelineTriggerDeleteError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.PipelineTriggerDeleteSuccess, nil)
}
