package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanPlanList 获取指定仓库的扫描计划列表
// @Summary 获取指定仓库的扫描计划列表
// @Description 根据仓库 ID 和过滤条件查询扫描计划列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param body body scanModel.ScanPlanReqFilter true "查询过滤条件"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanPlan,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/list [post]
func (c *ScanController) ScanPlanList(ctx *gin.Context) {
	repoID, ok := c.pathRepoID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var filter scanModel.ScanPlanReqFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanBodyParamError))
		return
	}
	req.SetDefaults(&filter)

	result, err := c.planSvc.List(ctx.Request.Context(), repoID, filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{Total: result.Total, Page: int64(filter.Page), PageSize: int64(filter.PageSize)},
		i18n.PlanListSuccess, nil)
}

// ScanPlanCreate 仓库创建新的扫描计划
// @Summary 仓库创建新的扫描计划
// @Description 在指定仓库下创建扫描计划，并可选绑定扫描方案
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param body body scanModel.ScanPlanReqForCreate true "创建请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.ScanPlan} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans [post]
func (c *ScanController) ScanPlanCreate(ctx *gin.Context) {
	repoID, ok := c.pathRepoID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanReqForCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.planSvc.Create(ctx.Request.Context(), session, repoID, reqData)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.Created(ctx, i18n.PlanCreateSuccess, data)
}

// ScanPlanGet 获取指定仓库扫描计划
// @Summary 获取指定仓库扫描计划
// @Description 根据计划 ID 查询扫描计划详情
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanPlan} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID} [get]
func (c *ScanController) ScanPlanGet(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	obj, err := c.planSvc.GetByID(ctx.Request.Context(), repoID, planID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanQuerySuccess, obj)
}

// ScanPlanListRules 获取扫描计划展开后的有效规则列表
// @Summary 获取扫描计划展开后的有效规则列表
// @Description 根据计划 ID 查询最终生效的规则列表，可选传 triggerID 做作用域校验
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID query uint false "触发器 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.Rule,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/rules [get]
func (c *ScanController) ScanPlanListRules(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var query scanModel.ScanPlanRulesReqQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	_, rules, err := c.planSvc.ListRules(ctx.Request.Context(), repoID, planID, query.TriggerID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, rules,
		resp.ListArgs{Total: int64(len(rules)), Page: 1, PageSize: int64(len(rules))},
		i18n.PlanQuerySuccess, nil)
}

// ScanPlanRules 获取扫描计划和展开后的有效规则列表
// @Summary 获取扫描计划和展开后的有效规则列表
// @Description 根据计划 ID 查询最终生效的规则列表，可选传 triggerID 做作用域校验
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param triggerID query uint false "触发器 ID"
// @Success 200 {object} resp.ObjectResponse{data=gin.H{plan=scanModel.ScanPlan,rules=[]scanModel.Rule}} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/plan-rules [get]
func (c *ScanController) ScanPlanRules(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var query scanModel.ScanPlanRulesReqQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.TriggerPathParamError))
		return
	}

	plan, rules, err := c.planSvc.ListRules(ctx.Request.Context(), repoID, planID, query.TriggerID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanQuerySuccess, gin.H{
		"id":       plan.ID,
		"name":     plan.Name,
		"repoID":   plan.RepoID,
		"files":    plan.Files,
		"scanType": plan.ScanType,
		"rules":    rules,
	})
}

// ScanPlanUpdate 更新仓库扫描计划信息
// @Summary 更新仓库扫描计划信息
// @Description 根据计划 ID 更新扫描计划配置
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanReqForUpdate true "更新请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanPlan} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID} [patch]
func (c *ScanController) ScanPlanUpdate(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var reqBody scanModel.ScanPlanReqForUpdate
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.planSvc.Update(ctx.Request.Context(), session, repoID, planID, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanUpdateSuccess, data)
}

// ScanPlanDelete 删除指定仓库扫描计划
// @Summary 删除指定仓库扫描计划
// @Description 根据计划 ID 删除扫描计划及其绑定关系
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID} [delete]
func (c *ScanController) ScanPlanDelete(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	if err := c.planSvc.Delete(ctx.Request.Context(), repoID, planID); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanDeleteSuccess, nil)
}

// ScanPlanBindSolution 为扫描计划绑定扫描方案
// @Summary 为扫描计划绑定扫描方案
// @Description 将指定扫描方案绑定到扫描计划
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanReqForBind true "绑定请求参数"
// @Success 200 {object} resp.ObjectResponse "绑定成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/bind [post]
func (c *ScanController) ScanPlanBindSolution(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanReqForBind
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanBodyParamError))
		return
	}

	if err := c.planSvc.BindSolutions(ctx.Request.Context(), repoID, planID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanBindSuccess, nil)
}

// ScanPlanUnbindSolution 解除扫描计划与扫描方案的绑定
// @Summary 解除扫描计划与扫描方案的绑定
// @Description 解除扫描计划与指定扫描方案的关联关系
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Param body body scanModel.ScanPlanReqForBind true "解绑请求参数"
// @Success 200 {object} resp.ObjectResponse "解绑成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/unbind [post]
func (c *ScanController) ScanPlanUnbindSolution(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	var reqData scanModel.ScanPlanReqForBind
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.PlanBodyParamError))
		return
	}

	if err := c.planSvc.UnbindSolutions(ctx.Request.Context(), repoID, planID, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.PlanUnbindSuccess, nil)
}

// ScanPlanListUnbindSolution 获取扫描计划可绑定的扫描方案列表
// @Summary 获取扫描计划可绑定的扫描方案列表
// @Description 查询指定扫描计划未绑定的扫描方案列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param planID path uint true "计划 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.Solution,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/repos/{repoID}/scanplans/{planID}/unbind [get]
func (c *ScanController) ScanPlanListUnbindSolution(ctx *gin.Context) {
	repoID, planID, ok := c.pathRepoAndPlanID(ctx, i18n.PlanPathParamError)
	if !ok {
		return
	}

	solutions, err := c.planSvc.ListUnbindSolutions(ctx.Request.Context(), repoID, planID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, solutions,
		resp.ListArgs{Total: int64(len(solutions)), Page: 1, PageSize: int64(len(solutions))},
		i18n.PlanSolutionListSuccess, nil)
}
