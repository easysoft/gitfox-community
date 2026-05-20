package webhook

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Controller 处理与 webhook 相关的 HTTP 请求
type Controller struct {
	webhookSvc webhook.Interface
}

// NewController 创建 webhook 处理器实例
func NewController(webhookSvc webhook.Interface) *Controller {
	return &Controller{
		webhookSvc: webhookSvc,
	}
}

// @Summary 创建webhook
// @Description 新增一个webhook
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body webhookModel.ReqForCreate true "创建webhook请求参数"
// @Success 200 {object} resp.ObjectResponse{data=webhookModel.Webhook} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks [post]
func (c *Controller) Create(ctx *gin.Context) {
	var reqBody webhookModel.ReqForCreate
	var err error

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.WebhookBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	// 通过 repoID + displayName 查询是否已存在同名 webhook
	createInfo := webhookModel.Webhook{
		Internal:    false,
		ParentID:    int64(repoID),
		ParentType:  string(constant.WebhookParentRepo),
		CreatedBy:   session.User.Account,
		EditedBy:    session.User.Account,
		DisplayName: reqBody.DisplayName,
		Description: reqBody.Description,
		URL:         reqBody.URL,
		Secret:      reqBody.Secret,
		Insecure:    reqBody.Insecure,
		Triggers:    reqBody.Triggers, // triggers 入库时需要在 map 函数中转换为以逗号分隔的字符串
		AuthMethod:  constant.DefaultAuthMethod,
		AuthHeader:  constant.DefaultAuthHeader,
		Enabled:     reqBody.Enabled,
	}
	// 调用 service 层创建 webhook
	respData, err := c.webhookSvc.Create(ctx.Request.Context(), createInfo)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.Created(ctx, i18n.WebhookCreateSuccess, respData)
}

// @Summary 查询webhook详情
// @Description 根据 webhook ID 查询详情
// @Tags webhooks
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Success 200 {object} resp.ObjectResponse{data=webhookModel.Webhook} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID} [get]
func (c *Controller) Get(ctx *gin.Context) {
	webhookID, err := req.PathParamAsPositiveUint(ctx, "webhookID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	respData, err := c.webhookSvc.Get(ctx.Request.Context(), webhookID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.WebhookQuerySuccess, respData)
}

// @Summary 更新 webhook
// @Description 更新webhook的基本信息
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Param request body webhookModel.ReqForUpdate true "更新webhook请求参数"
// @Success 200 {object} resp.ObjectResponse{data=webhookModel.Webhook} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID} [put]
func (c *Controller) Update(ctx *gin.Context) {
	webhookID, err := req.PathParamAsPositiveUint(ctx, "webhookID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	var reqBody webhookModel.ReqForUpdate
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookBodyParamError))
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	respData, err := c.webhookSvc.Update(ctx.Request.Context(), webhookID, reqBody, session.User.Account)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.WebhookUpdateSuccess, respData)
}

// @Summary 删除webhook
// @Description 根据 webhook ID 软删除
// @Tags webhooks
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	webhookID, err := req.PathParamAsPositiveUint(ctx, "webhookID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	if err = c.webhookSvc.Delete(ctx.Request.Context(), webhookID); err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.WebhookDeleteSuccess, nil)
}

// @Summary 查询webhook列表
// @Description 分页查询仓库下的webhook列表
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request query webhookModel.ReqForList true "分页查询参数"
// @Success 200 {object} resp.ListResponse{data=[]webhookModel.Webhook} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/webhooks [get]
func (c *Controller) List(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.RepoPathParamError))
		return
	}

	var reqQuery webhookModel.ReqForList
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookBodyParamError))
		return
	}
	req.SetDefaults(&reqQuery)

	respData, err := c.webhookSvc.List(ctx.Request.Context(), int64(repoID), reqQuery)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData.Data, resp.ListArgs{
		Total:    respData.Total,
		Page:     int64(respData.Page),
		PageSize: int64(respData.PageSize),
	}, i18n.WebhookListSuccess, nil)
}

// @Summary 查询指定 webhook 的执行记录列表
// @Description 分页查询 webhook 下的执行记录列表
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Param request query webhookModel.ReqForList true "分页查询参数"
// @Success 200 {object} resp.ListResponse{data=[]webhookModel.Execution} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID}/executions [get]
func (c *Controller) ListExecutions(ctx *gin.Context) {
	webhookID, err := req.PathParamAsPositiveUint(ctx, "webhookID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	var reqQuery webhookModel.ReqForList
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookBodyParamError))
		return
	}
	req.SetDefaults(&reqQuery)

	respData, err := c.webhookSvc.ListExecutions(ctx.Request.Context(), webhookID, reqQuery)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData.Data, resp.ListArgs{
		Total:    respData.Total,
		Page:     int64(respData.Page),
		PageSize: int64(respData.PageSize),
	}, i18n.WebhookListExecutionsSuccess, nil)
}

// @Summary 查询指定 webhook 的执行记录列表
// @Description 分页查询 webhook 下的执行记录列表
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Param executionID path uint true "执行 ID"
// @Success 200 {object} resp.ObjectResponse{data=webhookModel.Execution} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID}/executions/{executionID} [get]
func (c *Controller) GetExecution(ctx *gin.Context) {
	executionID, err := req.PathParamAsPositiveUint(ctx, "executionID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	respData, err := c.webhookSvc.GetExecution(ctx.Request.Context(), executionID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.WebhookExecutionQuerySuccess, respData)
}

// @Summary 重新触发指定 webhook 的执行
// @Description 重新触发 webhook 下的执行
// @Tags webhooks
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param webhookID path uint true "webhook ID"
// @Param executionID path uint true "执行 ID"
// @Success 200 {object} resp.ObjectResponse{data=webhookModel.Execution} "重新触发成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/webhooks/{webhookID}/executions/{executionID}/retrigger [post]
func (c *Controller) RetriggerExecution(ctx *gin.Context) {

	logger := logger.GetLoggerFromGin(ctx)

	executionID, err := req.PathParamAsPositiveUint(ctx, "executionID")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(err, i18n.WebhookPathParamError))
		return
	}

	respData, err := c.webhookSvc.RetriggerExecution(ctx.Request.Context(), executionID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	if respData.Err != nil {
		logger.Errorf("failed to retrigger webhook execution: %s", respData.Err.Error())
	}
	resp.OK(ctx, i18n.WebhookExecutionRetriggerSuccess, respData.Execution)
}
