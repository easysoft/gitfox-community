// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// @Summary 迁移规则
// @Description 迁移扫描规则，包括规则集、解决方法和规则
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=string} "迁移成功"
// @Failure 401 {object} resp.ObjectResponse "未授权"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /scan/rules/migrate [post]
func (s *ScanController) MigrateRules(ctx *gin.Context) {
	var err error

	err = s.ruleSvc.Migrate(ctx.Request.Context())
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleCreateError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleCreateSuccess, nil)
}

// @Summary 查询规则列表
// @Description 根据传入的查询参数分页查询扫描规则列表，支持按规则集、语言、插件等条件过滤
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param request body scan.RuleListReqFilter true "查询规则请求参数"
// @Success 200 {object} resp.ListResponse{data=[]scan.Rule,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
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

	result, err := s.ruleSvc.List(ctx.Request.Context(), session, params)
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

// @Summary 获取规则详情
// @Description 根据规则 ID 获取规则详细信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则 ID"
// @Success 200 {object} resp.ObjectResponse{data=scan.Rule} "获取成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权"
// @Failure 404 {object} resp.ObjectResponse "规则不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
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

	rule, err := s.ruleSvc.GetByID(ctx.Request.Context(), session, ruleID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleQueryError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleQuerySuccess, rule)
}

// @Summary 切换规则状态
// @Description 根据规则 ID 切换规则的启用/禁用状态（enabled ↔ disabled）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "规则 ID"
// @Success 200 {object} resp.ObjectResponse "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权"
// @Failure 404 {object} resp.ObjectResponse "规则不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
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

	err = s.ruleSvc.UpdateStatus(ctx.Request.Context(), session, ruleID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleUpdateError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleUpdateSuccess, nil)
}

// @Summary 查询所有扫描类型
// @Description 获取所有可用的扫描类型列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=[]scan.ScanBaseResp} "查询成功"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（数据库查询失败等）"
// @Router /scan/types [get]
func (s *ScanController) AllTypes(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	result, err := s.ruleSvc.AllScanBaseByField(ctx.Request.Context(), session, "type")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleGetAllTypesError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleGetAllTypesSuccess, result)
}

// @Summary 查询所有扫描插件
// @Description 获取所有可用的扫描插件列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=[]scan.ScanBaseResp} "查询成功"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（数据库查询失败等）"
// @Router /scan/plugins [get]
func (s *ScanController) AllPlugins(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	result, err := s.ruleSvc.AllScanBaseByField(ctx.Request.Context(), session, "plugin")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleGetAllPluginsError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleGetAllPluginsSuccess, result)
}

// @Summary 查询所有扫描语言
// @Description 获取所有可用的扫描语言列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=[]scan.ScanBaseResp} "查询成功"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（数据库查询失败等）"
// @Router /scan/langs [get]
func (s *ScanController) AllLangs(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	result, err := s.ruleSvc.AllScanBaseByField(ctx.Request.Context(), session, "lang")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleGetAllLangsError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleGetAllLangsSuccess, result)
}

// @Summary 查询所有扫描标签
// @Description 获取所有可用的扫描标签列表
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=[]scan.ScanBaseResp} "查询成功"
// @Error 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（数据库查询失败等）"
// @Router /scan/tags [get]
func (s *ScanController) AllTags(ctx *gin.Context) {
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	result, err := s.ruleSvc.AllScanBaseByField(ctx.Request.Context(), session, "tag")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RuleGetAllTagsError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RuleGetAllTagsSuccess, result)
}
