package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanTriggerList 获取扫描计划的触发器列表
// @Summary 获取扫描计划的触发器列表
// @Description 根据扫描计划 ID 查询触发器列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanTrigger,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers [get]
func (c *ScanController) ScanTriggerList(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	triggers, err := c.triggerSvc.List(ctx.Request.Context(), repoID, planID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, triggers,
		resp.ListArgs{Total: int64(len(triggers)), Page: 1, PageSize: int64(len(triggers))},
		i18n.TriggerListSuccess, nil)
}

// ScanTriggerCreate 为扫描计划创建触发器
// @Summary 为扫描计划创建触发器
// @Description 在指定扫描计划下创建触发器配置
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanTriggerReqForCreate true "创建请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.ScanTrigger} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers [post]
func (c *ScanController) ScanTriggerCreate(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	var reqData scanModel.ScanTriggerReqForCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.triggerSvc.Create(ctx.Request.Context(), session, planID, repoID, reqData)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.Created(ctx, i18n.TriggerCreateSuccess, data)
}

// ScanTriggerGet 获取触发器详细信息
// @Summary 获取触发器详细信息
// @Description 根据触发器 ID 查询触发器详情
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanTrigger} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID} [get]
func (c *ScanController) ScanTriggerGet(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	obj, err := c.triggerSvc.GetByID(ctx.Request.Context(), repoID, planID, triggerID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.TriggerQuerySuccess, obj)
}

// ScanTriggerUpdate 更新触发器配置
// @Summary 更新触发器配置
// @Description 根据触发器 ID 更新触发器配置
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Param body body scanModel.ScanTriggerReqForUpdate true "更新请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanTrigger} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID} [patch]
func (c *ScanController) ScanTriggerUpdate(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	var reqBody scanModel.ScanTriggerReqForUpdate
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.triggerSvc.Update(ctx.Request.Context(), session, repoID, planID, triggerID, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.TriggerUpdateSuccess, data)
}

// ScanTriggerDelete 删除指定触发器
// @Summary 删除指定触发器
// @Description 根据触发器 ID 删除触发器
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID} [delete]
func (c *ScanController) ScanTriggerDelete(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	if err := c.triggerSvc.Delete(ctx.Request.Context(), repoID, planID, triggerID); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.TriggerDeleteSuccess, nil)
}

// ScanConditionList 获取触发器的条件列表
// @Summary 获取触发器的条件列表
// @Description 根据触发器 ID 查询条件列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanCondition,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID}/conditions [get]
func (c *ScanController) ScanConditionList(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	conditions, err := c.conditionSvc.List(ctx.Request.Context(), repoID, planID, triggerID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, conditions,
		resp.ListArgs{Total: int64(len(conditions)), Page: 1, PageSize: int64(len(conditions))},
		i18n.ConditionListSuccess, nil)
}

// ScanConditionBatchCreate 为触发器添加条件
// @Summary 为触发器添加条件
// @Description 为指定触发器批量创建条件
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Param body body scanModel.ScanConditionReqForBatchCreate true "批量创建请求参数"
// @Success 200 {object} resp.ObjectResponse "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID}/conditions/new [post]
func (c *ScanController) ScanConditionBatchCreate(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	var reqData scanModel.ScanConditionReqForBatchCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err := c.conditionSvc.BatchCreate(ctx.Request.Context(), session, planID, triggerID, repoID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ConditionBatchCreateSuccess, nil)
}

// ScanConditionBatchDelete 删除触发器条件
// @Summary 删除触发器条件
// @Description 按条件 ID 列表批量删除触发器条件
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID path uint true "触发器 ID"
// @Param body body scanModel.ScanConditionReqForBatchDelete true "批量删除请求参数"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/triggers/{triggerID}/conditions/delete [post]
func (c *ScanController) ScanConditionBatchDelete(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	planID, err := req.PathParamAsPositiveUint(ctx, "planID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	triggerID, err := req.PathParamAsPositiveUint(ctx, "triggerID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionPathParamError))
		return
	}

	var reqData scanModel.ScanConditionReqForBatchDelete
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ConditionBodyParamError))
		return
	}

	if err := c.conditionSvc.BatchDelete(ctx.Request.Context(), repoID, planID, triggerID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ConditionBatchDeleteSuccess, nil)
}
