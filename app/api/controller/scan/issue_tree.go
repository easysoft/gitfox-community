package scan

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanIssueFileTree 获取 issue 文件树
// @Summary 获取 issue 文件树
// @Description 按文件路径聚合扫描问题，返回树形结构
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID query int64 false "仓库 ID"
// @Param taskID query int64 false "任务 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.TreeNode} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/file-tree [get]
func (c *ScanController) ScanIssueFileTree(ctx *gin.Context) {
	var filter scanModel.IssueTreeFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssueBodyParamError))
		return
	}
	if filter.RepoID == 0 && filter.TaskID == 0 {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("repoID or taskID is required"),
			i18n.IssueBodyParamError))
		return
	}

	data, err := c.taskSvc.GetIssueFileTree(ctx.Request.Context(), filter.RepoID, filter.TaskID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.IssueTreeQuerySuccess, data)
}

// ScanIssueRuleTree 获取 issue 规则树
// @Summary 获取 issue 规则树
// @Description 按规则维度聚合扫描问题，返回树形结构
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param repoID query int64 false "仓库 ID"
// @Param taskID query int64 false "任务 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.TreeNode} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/issues/rule-tree [get]
func (c *ScanController) ScanIssueRuleTree(ctx *gin.Context) {
	var filter scanModel.IssueTreeFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.IssueBodyParamError))
		return
	}
	if filter.RepoID == 0 && filter.TaskID == 0 {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("repoID or taskID is required"),
			i18n.IssueBodyParamError))
		return
	}

	data, err := c.taskSvc.GetIssueRuleTree(ctx.Request.Context(), filter.RepoID, filter.TaskID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.IssueTreeQuerySuccess, data)
}
