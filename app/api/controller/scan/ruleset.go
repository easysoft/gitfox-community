package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Create 创建规则集
// @Summary 创建规则集
// @Description 基于传入参数创建规则集
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.RuleSetReqForCreate true "创建规则集请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.RuleSet} "创建成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets [post]
func (c *ScanController) Create(ctx *gin.Context) {
	reqData := new(scanModel.RuleSetReqForCreate)
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := c.ruleSetSvc.Create(ctx.Request.Context(), session, *reqData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetCreateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.Created(ctx, i18n.RuleSetCreateSuccess, data)
}

// List 查询规则集列表
// @Summary 查询规则集列表
// @Description 根据过滤条件查询规则集列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/js
// @Param body body scanModel.RuleSetReqFilter true "查询规则集请求参数"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.RuleSet,listArgs=resp.ListArgs} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/list [post]
func (c *ScanController) List(ctx *gin.Context) {
	var reqQuery scanModel.RuleSetReqFilter
	if err := ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 设置默认值
	req.SetDefaults(&reqQuery)

	result, err := c.ruleSetSvc.List(ctx.Request.Context(), reqQuery)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{
			Total:    result.Total,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.PageSize),
		}, i18n.RuleSetListSuccess, nil)
}

// Get 获取规则集
// @Summary 获取规则集信息
// @Description 根据传入的规则集 ID 获取规则集信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.RuleSet} "获取成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "规则集不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id} [get]
func (c *ScanController) Get(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	obj, err := c.ruleSetSvc.GetByID(ctx.Request.Context(), ruleSetID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetQueryError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetQuerySuccess, obj)
}

// Update 修改规则集信息
// @Summary 修改规则集信息
// @Description 修改规则集信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集 ID"
// @Param body body scanModel.RuleSetReqForUpdate true "更新规则集所需请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.RuleSet} "更新成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "规则集不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id} [put]
func (c *ScanController) Update(ctx *gin.Context) {
	var reqBody scanModel.RuleSetReqForUpdate
	var err error

	// 解析 path id
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	respData, err := c.ruleSetSvc.Update(ctx.Request.Context(), session, ruleSetID, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetUpdateSuccess, respData)
}

// Delete 删除规则集
// @Summary 删除规则集
// @Description 根据规则集 ID 删除规则集及其关联的规则
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id} [delete]
func (c *ScanController) Delete(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.ruleSetSvc.Delete(ctx.Request.Context(), ruleSetID); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetDeleteError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetDeleteSuccess, nil)
}

// GetBindRule 获取规则集已绑定的规则列表
// @Summary 获取规则集已绑定的规则列表
// @Description 根据规则集ID查询已绑定的规则列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(30)
// @Success 200 {object} resp.ListResponse{data=[]scanModel.Rule} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id}/bind [get]
func (c *ScanController) GetBindRule(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RulePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var reqQuery scanModel.RuleReqFilter
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 设置默认值
	req.SetDefaults(&reqQuery)

	result, err := c.ruleSetSvc.GetBindRule(ctx.Request.Context(), ruleSetID, reqQuery)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{
			Total:    result.Total,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.Size),
		}, i18n.RuleListSuccess, nil)
}

// GetUnBindRule 获取规则集未绑定的规则列表
// @Summary 获取规则集未绑定的规则列表
// @Description 根据规则集ID查询未绑定的规则列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(30)
// @Success 200 {object} resp.ListResponse{data=[]scanModel.Rule} "查询成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id}/unbind [get]
func (c *ScanController) GetUnBindRule(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RulePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var reqQuery scanModel.RuleReqFilter
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 设置默认值
	req.SetDefaults(&reqQuery)

	result, err := c.ruleSetSvc.GetUnBindRule(ctx.Request.Context(), ruleSetID, reqQuery)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{
			Total:    result.Total,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.Size),
		}, i18n.RuleListSuccess, nil)
}

// BindRule 绑定规则到规则集
// @Summary 绑定规则到规则集
// @Description 将指定的规则绑定到规则集
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集ID"
// @Param body body scanModel.RuleSetReqForBind true "绑定规则请求参数"
// @Success 200 {object} resp.ObjectResponse "绑定成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id}/bind [post]
func (c *ScanController) BindRule(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	reqData := new(scanModel.RuleSetReqForBind)
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.ruleSetSvc.BindRule(ctx.Request.Context(), ruleSetID, *reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetBindError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetBindSuccess, nil)
}

// UnBindRule 解绑规则与规则集
// @Summary 解绑规则与规则集
// @Description 将指定的规则从规则集中解绑
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则集ID"
// @Param body body scanModel.RuleSetReqForBind true "解绑规则请求参数"
// @Success 200 {object} resp.ObjectResponse "解绑成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id}/unbind [post]
func (c *ScanController) UnBindRule(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	reqData := new(scanModel.RuleSetReqForBind)
	if err := ctx.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.ruleSetSvc.UnBindRule(ctx.Request.Context(), ruleSetID, *reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetUnbindError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetUnbindSuccess, nil)
}
