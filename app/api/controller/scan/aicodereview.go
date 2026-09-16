package scan

import (
	"io"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

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
// @Description 上传评审任务的报告文件，解析 issues
// @Tags scan
// @Accept multipart/form-data
// @Produce application/json
// @Param ppmID path uint true "合并请求 ID"
// @Param file formData file true "报告文件"
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

	ppmID, err := req.PathParamAsPositiveUint(ctx, "ppmID")
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

	var report *scanModel.ScanIssueReport
	report, err = c.taskSvc.AnalyzeTaskUploadReport(ctx.Request.Context(), repoID, ppmID, method, content)
	if err != nil {
		resp.InternalError(ctx, errors.WithI18nKey(err, i18n.CodeReviewReportUploadError))
		return
	}

	resp.OK(ctx, i18n.CodeReviewReportUploadSuccess, report)
}
