package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanPlanConditionCreate 创建扫描计划级条件
// @Summary 创建扫描计划级条件
// @Description 在指定扫描计划下创建计划级条件（triggerID=0）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanConditionReqForCreate true "创建请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.ScanCondition} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions [post]
func (c *ScanController) ScanPlanConditionCreate(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanConditionReqForCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanConditionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.planConditionSvc.Create(ctx.Request.Context(), session, repoID, planID, reqData)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.Created(ctx, i18n.PlanConditionCreateSuccess, data)
}

// ScanPlanConditionList 获取扫描计划级条件列表
// @Summary 获取扫描计划级条件列表
// @Description 根据扫描计划 ID 查询计划级条件列表（triggerID=0）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanCondition,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions [get]
func (c *ScanController) ScanPlanConditionList(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	conditions, err := c.planConditionSvc.List(ctx.Request.Context(), repoID, planID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, conditions,
		resp.ListArgs{Total: int64(len(conditions)), Page: 1, PageSize: int64(len(conditions))},
		i18n.PlanConditionListSuccess, nil)
}

// ScanPlanConditionFind 获取扫描计划级条件详情
// @Summary 获取扫描计划级条件详情
// @Description 根据条件 ID 查询计划级条件详情
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param conditionID path uint true "条件 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanCondition} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions/{conditionID} [get]
func (c *ScanController) ScanPlanConditionFind(ctx *gin.Context) {
	repoID, planID, conditionID, ok := c.pathRepoPlanAndConditionID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	obj, err := c.planConditionSvc.Find(ctx.Request.Context(), repoID, planID, conditionID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanConditionQuerySuccess, obj)
}

// ScanPlanConditionUpdate 更新扫描计划级条件
// @Summary 更新扫描计划级条件
// @Description 根据条件 ID 更新计划级条件
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param conditionID path uint true "条件 ID"
// @Param body body scanModel.ScanPlanConditionReqForUpdate true "更新请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanCondition} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions/{conditionID} [patch]
func (c *ScanController) ScanPlanConditionUpdate(ctx *gin.Context) {
	repoID, planID, conditionID, ok := c.pathRepoPlanAndConditionID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	var reqBody scanModel.ScanPlanConditionReqForUpdate
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanConditionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.planConditionSvc.Update(ctx.Request.Context(), session, repoID, planID, conditionID, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanConditionUpdateSuccess, data)
}

// ScanPlanConditionDelete 删除扫描计划级条件
// @Summary 删除扫描计划级条件
// @Description 根据条件 ID 删除计划级条件
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param conditionID path uint true "条件 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions/{conditionID} [delete]
func (c *ScanController) ScanPlanConditionDelete(ctx *gin.Context) {
	repoID, planID, conditionID, ok := c.pathRepoPlanAndConditionID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	if err := c.planConditionSvc.Delete(ctx.Request.Context(), repoID, planID, conditionID); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanConditionDeleteSuccess, nil)
}

// ScanPlanConditionBatchCreate 批量创建扫描计划级条件
// @Summary 批量创建扫描计划级条件
// @Description 为指定扫描计划批量创建计划级条件（triggerID=0）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanConditionReqForBatchCreate true "批量创建请求参数"
// @Success 200 {object} resp.ObjectResponse "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions/batch/new [post]
func (c *ScanController) ScanPlanConditionBatchCreate(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanConditionReqForBatchCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanConditionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err := c.planConditionSvc.BatchCreate(ctx.Request.Context(), session, repoID, planID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanConditionBatchCreateSuccess, nil)
}

// ScanPlanConditionBatchDelete 批量删除扫描计划级条件
// @Summary 批量删除扫描计划级条件
// @Description 按条件 ID 列表批量删除计划级条件
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanConditionReqForBatchDelete true "批量删除请求参数"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/conditions/batch/delete [post]
func (c *ScanController) ScanPlanConditionBatchDelete(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanConditionPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanConditionReqForBatchDelete
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanConditionBodyParamError))
		return
	}

	if err := c.planConditionSvc.BatchDelete(ctx.Request.Context(), repoID, planID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanConditionBatchDeleteSuccess, nil)
}
