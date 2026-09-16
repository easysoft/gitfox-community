package scan

import (
	"fmt"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanIssueList 查询扫描问题列表
// @Summary 查询扫描问题列表
// @Description 根据过滤条件查询扫描问题详情列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.ScanIssueReqFilter true "查询过滤条件"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanIssueInfo,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/list [post]
func (c *ScanController) ScanIssueList(ctx *gin.Context) {
	var filter scanModel.ScanIssueReqFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssueBodyParamError))
		return
	}
	req.SetDefaults(&filter)

	result, err := c.taskSvc.ListIssueInfos(ctx.Request.Context(), filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{Total: result.Total, Page: int64(filter.Page), PageSize: int64(filter.PageSize)},
		i18n.IssueListSuccess, nil)
}

// ScanIssueGet 获取扫描问题详情
// @Summary 获取扫描问题详情
// @Description 根据扫描问题 ID 获取扫描问题的详细信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描问题 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanIssueInfo} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/{id} [get]
func (c *ScanController) ScanIssueGet(ctx *gin.Context) {
	issueID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssuePathParamError))
		return
	}

	obj, err := c.taskSvc.GetIssueInfo(ctx.Request.Context(), issueID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.IssueQuerySuccess, obj)
}

// ScanIssueListByIds 根据 ID 列表查询扫描问题
// @Summary 根据 ID 列表查询扫描问题
// @Description 根据扫描问题 ID 列表批量查询扫描问题详情
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.ScanIssueIDsReqFilter true "ID 列表"
// @Success 200 {object} resp.ObjectResponse{data=[]scanModel.ScanIssueInfo} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/batch [post]
func (c *ScanController) ScanIssueListByIds(ctx *gin.Context) {
	var filter scanModel.ScanIssueIDsReqFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssueBodyParamError))
		return
	}
	if len(filter.IDs) > constant.MaxIssueIDsPerRequest {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("ids count exceeds limit %d", constant.MaxIssueIDsPerRequest),
			i18n.IssueBodyParamError))
		return
	}

	result, err := c.taskSvc.ListIssueInfosByIDs(ctx.Request.Context(), filter.IDs)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.IssueListSuccess, result)
}

// ScanIssueUpdateStatus 批量更新扫描问题状态
// @Summary 批量更新扫描问题状态
// @Description 批量更新扫描问题的处理状态和/或解决方案
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.ScanIssueUpdateStatusReq true "更新请求"
// @Success 200 {object} resp.ObjectResponse "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/status [put]
func (c *ScanController) ScanIssueUpdateStatus(ctx *gin.Context) {
	var in scanModel.ScanIssueUpdateStatusReq
	if err := ctx.ShouldBindJSON(&in); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssueBodyParamError))
		return
	}
	if len(in.IDs) > constant.MaxIssueIDsPerRequest {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("ids count exceeds limit %d", constant.MaxIssueIDsPerRequest),
			i18n.IssueBodyParamError))
		return
	}
	if in.Status != nil {
		if _, ok := in.Status.Sanitize(); !ok {
			resp.BadRequest(ctx, errors.WithI18nKey(
				fmt.Errorf("invalid status: %s", *in.Status),
				i18n.IssueBodyParamError))
			return
		}
	}

	if err := c.taskSvc.SetIssuesStatus(ctx.Request.Context(), in); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.IssueUpdateSuccess, nil)
}
