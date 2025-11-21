package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

type RuleSetController struct {
	ruleSetSvc scanService.IRuleSet
}

// NewRuleSetController 创建规则集处理器实例
func NewRuleSetController(ruleSetSvc scanService.IRuleSet) *RuleSetController {
	return &RuleSetController{
		ruleSetSvc: ruleSetSvc,
	}
}

// Create 创建规则集
// @Summary 创建规则集
// @Description 基于传入参数创建规则集
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.RuleSetCreateReq true "创建规则集请求参数"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.RuleSet} "创建成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets [post]
func (c *RuleSetController) Create(ctx *gin.Context) {
	reqData := new(scanModel.RuleSetCreateReq)
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

	data, err := c.ruleSetSvc.Create(logger.GetContextWithLogger(ctx), session, *reqData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetCreateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.Created(ctx, i18n.RuleSetCreateSuccess, data)
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
func (c *RuleSetController) Get(ctx *gin.Context) {
	ruleSetID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleSetPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	obj, err := c.ruleSetSvc.GetByID(logger.GetContextWithLogger(ctx), ruleSetID)
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
// @Param body body scanModel.RuleSetUpdateReq true "更新规则集所需请求参数"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.RuleSet} "更新成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "规则集不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/rulesets/{id} [put]
func (c *RuleSetController) Update(ctx *gin.Context) {
	var reqBody scanModel.RuleSetUpdateReq
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

	respData, err := c.ruleSetSvc.Update(logger.GetContextWithLogger(ctx), session, ruleSetID, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleSetUpdateSuccess, respData)
}
