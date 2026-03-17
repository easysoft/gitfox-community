package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// SolutionCreate 创建扫描方案
// @Summary 创建扫描方案
// @Description 基于传入参数创建扫描方案
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.SolutionReqForCreate true "创建扫描方案请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.Solution} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions [post]
func (c *ScanController) SolutionCreate(ctx *gin.Context) {
	var reqData scanModel.SolutionReqForCreate
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.solutionSvc.Create(ctx.Request.Context(), session, reqData)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.Created(ctx, i18n.SolutionCreateSuccess, data)
}

// SolutionList 查询扫描方案列表
// @Summary 查询扫描方案列表
// @Description 根据过滤条件查询扫描方案列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.SolutionReqFilter true "查询过滤条件"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.Solution,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/list [post]
func (c *ScanController) SolutionList(ctx *gin.Context) {
	var reqQuery scanModel.SolutionReqFilter
	if err := ctx.ShouldBindJSON(&reqQuery); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionBodyParamError))
		return
	}
	req.SetDefaults(&reqQuery)

	result, err := c.solutionSvc.List(ctx.Request.Context(), reqQuery)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{Total: result.Total, Page: int64(reqQuery.Page), PageSize: int64(reqQuery.PageSize)},
		i18n.SolutionListSuccess, nil)
}

// SolutionGet 获取扫描方案详情
// @Summary 获取扫描方案详情
// @Description 根据传入的扫描方案 ID 获取扫描方案详细信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.Solution} "获取成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "扫描方案不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id} [get]
func (c *ScanController) SolutionGet(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	obj, err := c.solutionSvc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionQuerySuccess, obj)
}

// SolutionUpdate 更新扫描方案
// @Summary 更新扫描方案
// @Description 根据扫描方案 ID 更新扫描方案信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Param body body scanModel.SolutionReqForUpdate true "更新请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.Solution} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 404 {object} resp.ObjectResponse "扫描方案不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id} [put]
func (c *ScanController) SolutionUpdate(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	var reqBody scanModel.SolutionReqForUpdate
	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.solutionSvc.Update(ctx.Request.Context(), session, id, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionUpdateSuccess, data)
}

// SolutionDelete 删除扫描方案
// @Summary 删除扫描方案
// @Description 根据扫描方案 ID 删除扫描方案及其关联的规则集绑定关系
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id} [delete]
func (c *ScanController) SolutionDelete(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	if err := c.solutionSvc.Delete(ctx.Request.Context(), id); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionDeleteSuccess, nil)
}

// SolutionUpdateStatus 切换扫描方案状态（enabled ↔ disabled）
// @Summary 切换扫描方案状态
// @Description 切换扫描方案的启用/禁用状态，enabled ↔ disabled 互相切换
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Success 200 {object} resp.ObjectResponse "切换成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id}/status [patch]
func (c *ScanController) SolutionUpdateStatus(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err := c.solutionSvc.UpdateStatus(ctx.Request.Context(), session, id); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionStatusUpdateSuccess, nil)
}

// SolutionListRulesets 获取扫描方案已绑定的规则集列表
// @Summary 获取扫描方案已绑定的规则集列表
// @Description 根据扫描方案 ID 查询已绑定的规则集列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.RuleSet,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id}/rulesets [get]
func (c *ScanController) SolutionListRulesets(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	rulesets, err := c.solutionSvc.ListRulesets(ctx.Request.Context(), id)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, rulesets,
		resp.ListArgs{Total: int64(len(rulesets)), Page: 1, PageSize: int64(len(rulesets))},
		i18n.SolutionRulesetListSuccess, nil)
}

// SolutionListUnbindRulesets 获取扫描方案未绑定的规则集列表
// @Summary 获取扫描方案未绑定的规则集列表
// @Description 根据扫描方案 ID 查询未绑定的规则集列表（可用于绑定）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.RuleSet,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id}/rulesets/unbind [get]
func (c *ScanController) SolutionListUnbindRulesets(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	rulesets, err := c.solutionSvc.ListUnbindRulesets(ctx.Request.Context(), id)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, rulesets,
		resp.ListArgs{Total: int64(len(rulesets)), Page: 1, PageSize: int64(len(rulesets))},
		i18n.SolutionRulesetListSuccess, nil)
}

// SolutionBindRulesets 为扫描方案绑定规则集
// @Summary 为扫描方案绑定规则集
// @Description 将指定规则集绑定到扫描方案，建立关联关系
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Param body body scanModel.SolutionReqForBind true "绑定请求参数"
// @Success 200 {object} resp.ObjectResponse "绑定成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id}/rulesets/bind [post]
func (c *ScanController) SolutionBindRulesets(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	var reqData scanModel.SolutionReqForBind
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionBodyParamError))
		return
	}

	if err := c.solutionSvc.BindRulesets(ctx.Request.Context(), id, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionBindSuccess, nil)
}

// SolutionUnbindRulesets 解除扫描方案与规则集的绑定
// @Summary 解除扫描方案与规则集的绑定
// @Description 解除扫描方案与指定规则集的关联关系
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描方案 ID"
// @Param body body scanModel.SolutionReqForBind true "解绑请求参数"
// @Success 200 {object} resp.ObjectResponse "解绑成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/solutions/{id}/rulesets/unbind [post]
func (c *ScanController) SolutionUnbindRulesets(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionPathParamError))
		return
	}

	var reqData scanModel.SolutionReqForBind
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.SolutionBodyParamError))
		return
	}

	if err := c.solutionSvc.UnbindRulesets(ctx.Request.Context(), id, reqData); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.SolutionUnbindSuccess, nil)
}
