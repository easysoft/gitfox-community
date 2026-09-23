// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package commit

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	commitModel "github.com/easysoft/gitfox-community/app/model/commit"
	"github.com/easysoft/gitfox-community/app/service/commit"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Controller 处理与提交相关的 HTTP 请求
type Controller struct {
	commitSvc commit.Interface
	repoSvc   repo.Interface
}

// NewController 创建提交处理器实例
func NewController(commitService commit.Interface, repoService repo.Interface) *Controller {
	return &Controller{
		commitSvc: commitService,
		repoSvc:   repoService,
	}
}

// @Summary 创建提交
// @Description 新增一个提交
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body commitModel.ReqForCreate true "创建提交请求参数"
// @Success 200 {object} resp.ObjectResponse{data=commitModel.RespForCreate} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/commits [post]
func (c *Controller) Create(ctx *gin.Context) {
	var reqBody commitModel.ReqForCreate
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 检查是否为镜像仓库
	if repoOut.Mirror {
		err = errors.WithI18nKey(err, i18n.RepoMirrorReadOnly)
		resp.InternalError(ctx, err)
		return
	}
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层创建提交
	respData, err := c.commitSvc.Create(ctx.Request.Context(), reqBody, session, repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 记录审计日志
	// TODO: 记录审计日志
	resp.Created(ctx, i18n.CommitCreateSuccess, respData)
}

// @Summary 查询提交列表
// @Description 分页查询提交列表
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body commitModel.ReqFilter true "查询提交请求参数"
// @Success 200 {object} resp.ListResponse{data=commitModel.RespForList,listArgs=resp.ListNoTotalArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/commits/list [post]
func (c *Controller) List(ctx *gin.Context) {
	var reqQuery commitModel.ReqFilter
	var err error
	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	// 设置默认值
	req.SetDefaults(&reqQuery)

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.commitSvc.List(ctx.Request.Context(), reqQuery, *repoOut, reqQuery.GitRef)
	if err != nil {
		resp.InternalError(ctx, err, err.Error())
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListArgs{
			Total:    int64(respData.TotalCommits),
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.PageSize),
		}, i18n.CommitListSuccess, nil)
}

// @Summary 获取单个提交详情
// @Description 根据提交 SHA 获取提交详情
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param sha query string true "提交 SHA"
// @Success 200 {object} resp.ObjectResponse{data=commitModel.Commit} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "提交未找到"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/commits [get]
func (c *Controller) Get(ctx *gin.Context) {
	var params commitModel.ReqForGet
	var err error
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindQuery(&params); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 调用 service 层获取提交详情
	respData, err := c.commitSvc.Get(ctx.Request.Context(), params.SHA, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.CommitGetSuccess, respData)
}

// @Summary 获取提交分歧信息
// @Description 获取多个 git 引用之间的提交分歧（ahead/behind）信息
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body commitModel.ReqForDivergence true "获取分歧信息请求参数"
// @Success 200 {object} resp.ObjectResponse{data=[]commitModel.CommitDivergence} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/commits/divergence [post]
func (c *Controller) GetDivergence(ctx *gin.Context) {
	var reqBody commitModel.ReqForDivergence
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 调用 service 层获取提交分歧信息
	respData, err := c.commitSvc.GetDivergence(ctx.Request.Context(), reqBody, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.CommitGetDivergenceSuccess, respData)
}

// @Summary 获取提交 Diff
// @Description 获取指定提交的 diff 内容(text/plain 格式)
// @Tags commits
// @Accept application/json
// @Produce text/plain
// @Param id path uint true "仓库 ID"
// @Param request body commitModel.ReqForGet true "获取 diff 请求参数"
// @Success 200 {string} string "diff 文本"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "提交未找到"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/commits/diff [post]
func (c *Controller) GetDiff(ctx *gin.Context) {
	log := logger.GetLogger(ctx.Request.Context())
	var reqBody commitModel.ReqForGet
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	// 设置响应头为 text/plain
	ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.NoCache(ctx.Writer)

	// 调用 service 层,直接将 diff 写入响应流
	err = c.commitSvc.GetDiff(ctx.Request.Context(), reqBody.SHA, *repoOut, ctx.Writer)
	if err != nil {
		// 如果已经开始写入响应,则无法再使用标准的错误响应
		// 只能记录日志
		log.Errorf("Failed to get commit diff: %v", err)
		return
	}
}
