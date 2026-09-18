package scan

import (
	"io"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/app/service/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ScanTaskCreate 创建扫描任务
// @Summary 创建扫描任务
// @Description 创建一个扫描任务，触发流水线执行
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param taskInfo body scanModel.ScanTaskReqForCreate true "扫描任务信息"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanTask} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks [post]
func (c *ScanController) ScanTaskCreate(ctx *gin.Context) {

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	var body scanModel.ScanTaskReqForCreate
	if err := ctx.ShouldBindJSON(&body); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskBodyParamError))
		return
	}

	repo, err := c.repoSvc.GetByID(ctx, session, body.RepoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	plan, err := c.planSvc.GetByID(ctx.Request.Context(), repo.ID, body.PlanID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 按扫描计划的 scanType 解析提交范围：增量取最新两笔、全量取同一笔
	commitBefore, commitAfter, err := c.taskSvc.ResolveCommitRange(ctx.Request.Context(), repo.GitUID, constant.ScanType(plan.ScanType), body.Branch)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// CreateBranchHook 的形参为 sha=最新、beforeSha=次新，与 ResolveCommitRange 的
	// 返回顺序（before 次新、after 最新）相反，故交换传入（同 Retry）
	hook, err := c.taskSvc.CreateBranchHook(ctx.Request.Context(), repo, body.Branch, commitAfter, commitBefore, session)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	data, err := c.taskSvc.Create(ctx.Request.Context(), &scan.ScanTaskInput{
		Plan:    plan,
		Repo:    repo,
		Trigger: nil,
		Hook:    hook,
	})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ScanTaskCreateSuccess, data)
}

// ScanTaskCreateByRunner 流水线 runner 回调创建扫描任务
// @Summary 创建扫描任务（流水线 runner 回调）
// @Description codescan 插件在非内嵌流水线（INNER_PIPELINE 非 "true"）中扫描前回调本接口，创建扫描任务记录并返回上报报告所需的 taskID（扫描范围由插件的 SCAN_METHOD + SCAN_REF_RANGE 声明，不由本接口解析）
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param taskInfo body scanModel.ScanTaskRunnerCreateReq true "扫描任务信息"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanTaskRunnerCreateResp} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/create-by-runner [post]
func (c *ScanController) ScanTaskCreateByRunner(ctx *gin.Context) {
	var body scanModel.ScanTaskRunnerCreateReq
	if err := ctx.ShouldBindJSON(&body); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskBodyParamError))
		return
	}

	res, err := c.taskSvc.CreateRunnerTask(ctx.Request.Context(), &body)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ScanTaskCreateSuccess, res)
}

// ScanTaskList 查询扫描任务列表
// @Summary 查询扫描任务列表
// @Description 根据过滤条件查询扫描任务详情列表，支持分页
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param body body scanModel.ScanTaskInfoReqFilter true "查询过滤条件"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanTaskInfo,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/list [post]
func (c *ScanController) ScanTaskList(ctx *gin.Context) {
	var filter scanModel.ScanTaskInfoReqFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskBodyParamError))
		return
	}
	req.SetDefaults(&filter)

	result, err := c.taskSvc.ListTaskInfos(ctx.Request.Context(), filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{Total: result.Total, Page: int64(filter.Page), PageSize: int64(filter.PageSize)},
		i18n.ScanTaskListSuccess, nil)
}

// ScanTaskGet 获取扫描任务详情
// @Summary 获取扫描任务详情
// @Description 根据扫描任务 ID 获取扫描任务的详细信息
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param taskID path uint true "扫描任务 ID"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanTaskInfo} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/{taskID} [get]
func (c *ScanController) ScanTaskGet(ctx *gin.Context) {
	taskID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskPathParamError))
		return
	}

	obj, err := c.taskSvc.GetTaskInfo(ctx.Request.Context(), taskID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ScanTaskQuerySuccess, obj)
}

// ScanTaskUploadReport 上传扫描任务报告
// @Summary 上传扫描任务报告
// @Description 上传扫描任务的报告文件，解析 issues 并进行质量门禁判定
// @Tags scan
// @Accept multipart/form-data
// @Produce application/json
// @Param id path uint true "扫描任务 ID"
// @Param file formData file true "报告文件"
// @Param method formData string true "扫描方法（check/smell）"
// @Success 200 {object} resp.ObjectResponse{data=scanModel.ScanIssueReport} "上传成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/{id}/report [post]
func (c *ScanController) ScanTaskUploadReport(ctx *gin.Context) {
	taskID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskPathParamError))
		return
	}

	file, _, err := ctx.Request.FormFile("file")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskBodyParamError))
		return
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		resp.InternalError(ctx, errors.WithI18nKey(err, i18n.ScanTaskReportUploadError))
		return
	}

	method := ctx.Request.FormValue("method")
	if method == "" {
		resp.BadRequest(ctx, errors.WithI18nKey(nil, i18n.ScanTaskBodyParamError))
		return
	}

	report, err := c.taskSvc.AnalyzeTaskUploadReport(ctx.Request.Context(), 0, taskID, 0, method, content)
	if err != nil {
		resp.InternalError(ctx, errors.WithI18nKey(err, i18n.ScanTaskReportUploadError))
		return
	}

	resp.OK(ctx, i18n.ScanTaskReportUploadSuccess, report)
}

// ScanTaskIssueList 查询扫描任务关联的问题列表
// @Summary 查询扫描任务关联的问题列表
// @Description 根据任务 ID 查询关联的扫描问题列表，支持分页和过滤
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描任务 ID"
// @Param body body scanModel.ScanTaskIssueReqFilter true "查询过滤条件"
// @Success 200 {object} resp.ListResponse{data=[]scanModel.ScanIssueInfo,listArgs=resp.ListArgs} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/{id}/issues [post]
func (c *ScanController) ScanTaskIssueList(ctx *gin.Context) {
	taskID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskPathParamError))
		return
	}

	var filter scanModel.ScanTaskIssueReqFilter
	if err := ctx.ShouldBindJSON(&filter); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskBodyParamError))
		return
	}
	filter.TaskID = int64(taskID)
	req.SetDefaults(&filter)

	result, err := c.taskSvc.ListTaskIssues(ctx.Request.Context(), taskID, filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, result.Data,
		resp.ListArgs{Total: result.Total, Page: int64(filter.Page), PageSize: int64(filter.PageSize)},
		i18n.ScanTaskIssueListSuccess, nil)
}

// ScanTaskRetry 重试扫描任务
// @Summary 重试扫描任务
// @Description 重试一个失败的扫描任务，重新触发流水线执行
// @Tags scan
// @Accept application/json
// @Produce application/json
// @Param id path uint true "扫描任务 ID"
// @Success 201 {object} resp.ObjectResponse{data=scanModel.ScanTask} "重试成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权（会话不存在或已过期）"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /scan/tasks/{id}/retry [post]
func (c *ScanController) ScanTaskRetry(ctx *gin.Context) {
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	taskID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.ScanTaskPathParamError))
		return
	}

	taskInfo, err := c.taskSvc.GetTaskInfo(ctx.Request.Context(), taskID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	repo, err := c.repoSvc.GetByID(ctx, session, uint(taskInfo.RepoID))
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	task, err := c.taskSvc.Retry(ctx.Request.Context(), session, taskID, repo)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.Created(ctx, i18n.ScanTaskRetrySuccess, task)
}
