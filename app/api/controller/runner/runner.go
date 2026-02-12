package runner

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	runnerModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/runner"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	runnerSvc runner.Interface
	urlConfig config.URL
}

// NewController 创建 runner 管理控制器
func NewController(runnerSvc runner.Interface, urlConfig config.URL) *Controller {
	return &Controller{
		runnerSvc: runnerSvc,
		urlConfig: urlConfig,
	}
}

// List 查询 runner 列表
// @Summary 查询 runner 列表
// @Description 分页查询 runner 列表，支持多维度过滤
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param request body runnerModel.ReqFilter true "查询 runner 请求参数"
// @Success 200 {object} resp.ListResponse{data=[]runnerModel.Runner} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /runners/list [post]
func (ctl *Controller) List(c *gin.Context) {
	var reqQuery runnerModel.ReqFilter
	if err := c.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 设置默认值
	req.SetDefaults(&reqQuery)

	result, err := ctl.runnerSvc.List(c.Request.Context(), &reqQuery)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerListError)
		resp.InternalError(c, err)
		return
	}

	resp.ListSuccess(c, result.Data,
		resp.ListArgs{
			Total:    result.Total,
			Page:     int64(result.Page),
			PageSize: int64(result.PageSize),
		}, i18n.RunnerListSuccess, nil)
}

// GetByID 根据 ID 获取 runner 详情
// @Summary 根据 ID 获取 runner 详情
// @Description 根据 ID 查询单个 runner 的详细信息
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param id path int true "Runner ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse{data=runnerModel.Runner} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "ID格式错误"
// @Failure 404 {object} resp.ObjectResponse "Runner不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runners/{id} [get]
func (ctl *Controller) GetByID(c *gin.Context) {
	runnerID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerPathParamError)
		resp.BadRequest(c, err)
		return
	}

	runner, err := ctl.runnerSvc.GetByID(c.Request.Context(), runnerID)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, i18n.RunnerQueryError)
			resp.NotFound(c, err)
			return
		}
		err = errors.WithI18nKey(err, i18n.RunnerQueryError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerQuerySuccess, runner)
}

// UpdateBasicInfo 更新 runner 基础信息
// @Summary 更新 runner 基础信息
// @Description 更新 runner 的名称和描述信息
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param id path int true "Runner ID" minimum(1)
// @Param request body runnerModel.ReqForUpdateBasicInfo true "更新基础信息请求参数"
// @Success 200 {object} resp.ObjectResponse{data=runnerModel.Runner} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "Runner不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runners/{id} [put]
func (ctl *Controller) UpdateBasicInfo(c *gin.Context) {

	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	runnerID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerPathParamError)
		resp.BadRequest(c, err)
		return
	}

	var reqData runnerModel.ReqForUpdateBasicInfo
	if err := c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 获取现有的 runner
	runner, err := ctl.runnerSvc.GetByID(c.Request.Context(), runnerID)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, i18n.RunnerQueryError)
			resp.NotFound(c, err)
			return
		}
		err = errors.WithI18nKey(err, i18n.RunnerQueryError)
		resp.InternalError(c, err)
		return
	}

	// 更新基础信息
	runner.Name = reqData.Name
	runner.Desc = reqData.Desc
	runner.EditedBy = session.User.Account

	err = ctl.runnerSvc.Update(c.Request.Context(), runner)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerUpdateError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerUpdateSuccess, runner)
}

// UpdateStatus 更新 runner 状态
// @Summary 更新 runner 状态
// @Description 启用或停用 runner
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param id path int true "Runner ID" minimum(1)
// @Param request body runnerModel.ReqForUpdateStatus true "更新状态请求参数"
// @Success 200 {object} resp.ObjectResponse{data=runnerModel.Runner} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "Runner不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runners/{id}/status [put]
func (ctl *Controller) UpdateStatus(c *gin.Context) {

	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	runnerID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerPathParamError)
		resp.BadRequest(c, err)
		return
	}

	var reqData runnerModel.ReqForUpdateStatus
	if err := c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 获取现有的 runner
	runner, err := ctl.runnerSvc.GetByID(c.Request.Context(), runnerID)
	if err != nil {
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, i18n.RunnerQueryError)
			resp.NotFound(c, err)
			return
		}
		err = errors.WithI18nKey(err, i18n.RunnerQueryError)
		resp.InternalError(c, err)
		return
	}

	// 更新状态
	runner.Status = reqData.Status
	runner.EditedBy = session.User.Account

	err = ctl.runnerSvc.Update(c.Request.Context(), runner)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerUpdateError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerUpdateSuccess, runner)
}

// Delete 删除 runner
// @Summary 删除 runner
// @Description 软删除 runner
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param id path int true "Runner ID" minimum(1)
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "ID格式错误"
// @Failure 404 {object} resp.ObjectResponse "Runner不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runners/{id} [delete]
func (ctl *Controller) Delete(c *gin.Context) {

	runnerID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerPathParamError)
		resp.BadRequest(c, err)
		return
	}

	if err = ctl.runnerSvc.Delete(c.Request.Context(), runnerID); err != nil {
		if errors.IsNotFound(err) {
			err = errors.WithI18nKey(err, i18n.RunnerQueryError)
			resp.NotFound(c, err)
			return
		}
		err = errors.WithI18nKey(err, i18n.RunnerDeleteError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerDeleteSuccess, nil)
}

// GenerateToken 为当前认证的 account 生成 runner token
// @Summary 生成 Runner Token
// @Description 为当前认证的用户生成临时 Runner 注册 Token，存入缓存
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=runnerModel.GenerateTokenResponse} "生成成功"
// @Failure 401 {object} resp.ObjectResponse "未认证"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runners/token [get]
func (ctl *Controller) GenerateToken(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	token, err := ctl.runnerSvc.GenerateRunnerToken(c.Request.Context(), session.User.Account)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerGenerateTokenError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerGenerateTokenSuccess, runnerModel.GenerateTokenResponse{
		Token: token,
		URL:   ctl.urlConfig.Base,
	})
}

// Register runner 信息注册
// @Summary 注册 Runner
// @Description 使用临时 Token 注册 Runner 信息，注册成功后返回持久化的 Runner Token
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Param request body runnerModel.RegisterRunnerReq true "注册 runner 请求参数"
// @Success 201 {object} resp.ObjectResponse{data=runnerModel.Runner} "注册成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "Token无效或已过期"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runner-client/register [post]
func (ctl *Controller) Register(c *gin.Context) {

	// 解析 X-Runner-Token 头部
	runnerToken := c.GetHeader("X-Runner-Token")
	if runnerToken == "" {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.RunnerInvalidToken))
		return
	}
	// 验证临时 token，获取对应的 account 信息
	account, err := ctl.runnerSvc.ValidateTempRunnerToken(c.Request.Context(), runnerToken)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerInvalidToken)
		resp.Unauthorized(c, err)
		return
	}
	// 获取请求来源IP
	clientIP := c.ClientIP()

	var reqData runnerModel.RegisterRunnerReq
	if err := c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	// 构建 runner 模型
	runner := &runnerModel.Runner{
		Name:      reqData.Name,
		Desc:      reqData.Desc,
		Version:   reqData.Version,
		Hostname:  reqData.Hostname,
		IP:        reqData.IP,
		OS:        reqData.OS,
		Arch:      reqData.Arch,
		Labels:    reqData.Labels,
		CreatedBy: account,
		EditedBy:  account,
	}

	// 创建 runner
	err = ctl.runnerSvc.Create(c.Request.Context(), runner, runnerToken, clientIP)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerRegisterError)
		resp.InternalError(c, err)
		return
	}

	resp.Created(c, i18n.RunnerRegisterSuccess, runner)
}

// UpdateHeartBeat 更新 runner 心跳
// @Summary 更新 runner 心跳
// @Description 更新 runner 的心跳时间戳
// @Tags runners
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse{data=runnerModel.Runner} "更新成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "Runner不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /runner-client/keep-alive [post]
func (ctl *Controller) UpdateHeartBeat(c *gin.Context) {

	// 解析 X-Runner-Token 头部
	runnerToken := c.GetHeader("X-Runner-Token")
	if runnerToken == "" {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.RunnerInvalidToken))
		return
	}

	// 验证 runner token 并获取 runner ID
	runner, err := ctl.runnerSvc.ValidateRunnerToken(c.Request.Context(), runnerToken)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerInvalidToken)
		resp.Unauthorized(c, err)
		return
	}

	// 更新心跳
	err = ctl.runnerSvc.UpdateRunnerHeartbeat(c.Request.Context(), runner.ID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RunnerUpdateHeartBeatError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RunnerUpdateHeartBeatSuccess, runner)
}
