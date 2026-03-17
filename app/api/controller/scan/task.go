package scan

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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

	hook, err := c.taskSvc.CreateBranchHook(ctx, repo, body.Branch, session)
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
