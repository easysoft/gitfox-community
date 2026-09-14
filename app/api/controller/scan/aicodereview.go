package scan

import (
	"io"
	"strconv"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// GetAIReviewConfig 获取 AI 评审（OCR LLM）配置
// @Summary 获取 AI 评审配置
// @Description aireview 插件在非内嵌流水线（INNER_PIPELINE 非 "true"）中拉取 AI 评审所需的 LLM 配置（供应商/地址/协议/密钥/模型/语言）
// @Tags scan
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=scanModel.AIReviewConfig} "查询成功"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/aicodereview/ai-config [get]
func (c *ScanController) GetAIReviewConfig(ctx *gin.Context) {
	config, err := c.aicodereviewSvc.GetAIReviewConfig(ctx.Request.Context())
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.CodeReviewConfigSuccess, config)
}

// GetAIReviewRange 获取 AI 评审范围（提交区间）
// @Summary 获取 AI 评审范围
// @Description aireview 插件在非内嵌流水线（INNER_PIPELINE 非 "true"）中拉取指定分支的评审范围：取该分支最新两笔提交，commitBefore 为次新、commitAfter 为最新（即评审最新一笔的改动）；分支只有一笔提交时两者相同，插件侧会得到 skipped 空报告。插件据此写入 GITFOX_COMMIT_BEFORE/GITFOX_COMMIT_AFTER
// @Tags scan
// @Produce application/json
// @Param repoID query uint true "仓库 ID"
// @Param branch query string true "评审分支（如 main）"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.AIReviewRange} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/aicodereview/review-range [get]
func (c *ScanController) GetAIReviewRange(ctx *gin.Context) {
	repoID, err := req.QueryParamAsPositiveInt64OrError(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.CodeReviewQueryParamError))
		return
	}
	branch, err := req.QueryParamOrError(ctx, "branch")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.CodeReviewQueryParamError))
		return
	}

	result, err := c.aicodereviewSvc.GetReviewRange(ctx.Request.Context(), uint(repoID), branch)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.CodeReviewRangeSuccess, result)
}

func (c *ScanController) GetAIReviewRules(ctx *gin.Context) {
	result, err := c.aicodereviewSvc.GetAIReviewRules(ctx.Request.Context())
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result, nil, i18n.CodeReviewRuleListSuccess)
}

// ReviewTaskUploadReport 上传评审任务报告
// @Summary 上传评审任务报告
// @Description 上传评审任务的报告文件，解析 issues。ppmID 可为 0（无 MR，推送/内嵌流水线触发），此时报告按表单 executionID 归属到对应执行
// @Tags scan
// @Accept multipart/form-data
// @Produce application/json
// @Param repoID path uint true "仓库 ID"
// @Param ppmID path uint true "合并请求 ID，可为 0（无 MR）"
// @Param file formData file true "报告文件"
// @Param method formData string true "扫描方法（ai）"
// @Param executionID formData int true "流水线执行 ID（ppmID 为 0 时必填）"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanIssueReport} "上传成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/aicodereview/{repoID}/{ppmID}/report [post]
func (c *ScanController) ReviewTaskUploadReport(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "repoID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.CodeReviewPathParamError))
		return
	}

	// ppmID 允许为 0：语义为无 MR（推送/内嵌流水线触发），报告归属由 executionID 决定
	ppmID, err := req.PathParamAsNaturalNum(ctx, "ppmID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.CodeReviewPathParamError))
		return
	}

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.CodeReviewBodyParamError))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		resp.InternalError(ctx, errors.WithI18nKey(err, i18n.CodeReviewReportUploadError))
		return
	}

	method := ctx.Request.FormValue("method")
	if method == "" {
		resp.BadRequest(ctx, errors.WithI18nKey(nil, i18n.CodeReviewBodyParamError))
		return
	}

	// executionID（GITFOX_EXECUTION_ID）为数字字符串；ppmID=0（无 MR）时必须携带非 0 值
	var executionID int64
	if method == string(constant.ScanMethodAI) {
		if executionID, err = strconv.ParseInt(ctx.Request.FormValue("executionID"), 10, 64); err != nil {
			executionID = 0
		}
		if ppmID == 0 && executionID <= 0 {
			resp.BadRequest(ctx, errors.WithI18nKey(nil, i18n.CodeReviewBodyParamError))
			return
		}
	}

	var report *scanModel.ScanIssueReport
	report, err = c.taskSvc.AnalyzeTaskUploadReport(ctx.Request.Context(), repoID, ppmID, executionID, method, content)
	if err != nil {
		resp.InternalError(ctx, errors.WithI18nKey(err, i18n.CodeReviewReportUploadError))
		return
	}

	resp.OK(ctx, i18n.CodeReviewReportUploadSuccess, report)
}
