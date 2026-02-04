package execution

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/canceler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	executionService "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与 pipeline 执行相关的 HTTP 请求
type Controller struct {
	triggerer        triggerer.Triggerer
	canceler         canceler.Canceler
	executionService executionService.Interface
}

// NewController 创建流水线执行实例
func NewController(triggerer triggerer.Triggerer, canceler canceler.Canceler, executionService executionService.Interface) *Controller {
	return &Controller{
		triggerer:        triggerer,
		canceler:         canceler,
		executionService: executionService,
	}
}

// @Summary 创建流水线执行
// @Description 创建新的流水线执行
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param request body executionModel.ReqForCreate true "创建流水线执行的请求参数"
// @Success 200 {object} resp.ObjectResponse{data=executionModel.Execution} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions [post]
func (c *Controller) Create(ctx *gin.Context) {
	var reqBody executionModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	ref, ok := reqBody.Params["gitRef"]
	if !ok {
		ref = ""
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}
	// Create manual hook for execution.
	hook := &triggerer.Hook{
		Trigger:     session.User.Account, // who/what triggered the build, different from commit author
		TriggeredBy: session.User.ID,
		Ref:         ref,
		Sender:      session.User.Account,
		Params:      reqBody.Params,
	}

	executionOutput, err := c.triggerer.Trigger(ctx, reqBody.PipelineID, hook)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionTriggerError)
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ExecutionTriggerSuccess, executionOutput)
}

// @Summary 查询流水线执行详情
// @Description 查询流水线执行的详细信息
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param id path uint true "流水线执行ID"
// @Success 200 {object} resp.ObjectResponse{data=executionModel.ExecutionWithStages} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions/{id} [get]
func (c *Controller) Get(ctx *gin.Context) {
	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	result, err := c.executionService.GetWithStages(ctx, id)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionQueryError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ExecutionQuerySuccess, result)
}

// @Summary 取消流水线执行
// @Description 取消指定的流水线执行
// @Tags executions
// @Accept application/json
// @Produce application/json
// @Param id path uint true "流水线执行ID"
// @Success 200 {object} resp.ObjectResponse "取消成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /pipeline/executions/{id}/cancel [post]
func (c *Controller) Cancel(ctx *gin.Context) {

	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	id, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	err = c.canceler.Cancel(ctx, id, session.User.Account)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ExecutionCancelError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ExecutionCancelSuccess, nil)
}
