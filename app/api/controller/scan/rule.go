package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

func (s *ScanController) MigrateRules(ctx *gin.Context) {
	// 实现规则迁移逻辑
}

// ListRules 获取规则列表
// @Summary 获取规则列表
// @Description 根据传入的查询参数获取规则列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param filter query scan.RuleListReqFilter true "查询参数"
// @Success 200 {object} resp.ObjectResponse{data=database.PageResult[scanModel.Rule]} "获取成功"
// @Router /scan/rules/list [post]
func (s *ScanController) ListRules(ctx *gin.Context) {
	var err error
	var params scan.RuleListReqFilter
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if err = ctx.ShouldBindJSON(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err = params.Validater(); err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	result, err := s.ruleSvc.List(ctx, session, params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleListError)
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{
			Total:    result.Total,
			Page:     int64(params.Page),
			PageSize: int64(params.PageSize),
		}, i18n.RuleListSuccess, nil)
}

// GetRule 获取规则
// @Summary 获取规则信息
// @Description 根据传入的规则 ID 获取规则信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.Rule} "获取成功"
// @Router /scan/rules/{id} [get]
func (s *ScanController) GetRule(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	ruleID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	rule, err := s.ruleSvc.GetByID(ctx, session, ruleID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleQuerySuccess, rule)
}

// UpdateRuleStatus 更新规则状态
// @Summary 更新规则状态
// @Description 根据传入的规则 ID 更新规则状态
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.Rule} "更新成功"
// @Router /scan/rules/{id}/status [put]
func (s *ScanController) UpdateRuleStatus(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	ruleID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RulePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	err = s.ruleSvc.UpdateStatus(ctx, session, ruleID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleUpdateError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleUpdateSuccess, nil)
}
