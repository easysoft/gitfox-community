package commit

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	commitModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
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
// @Param request body commitModel.ReqForCreate true "创建提交请求参数"
// @Success 200 {object} resp.ObjectResponse{data=commitModel.RespForCreate} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /commits [post]
func (c *Controller) Create(ctx *gin.Context) {

	var reqBody commitModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, reqBody.RepoID)
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
	resp.Created(ctx, i18n.CommitCreateSuccess, respData)
}

// @Summary 查询提交列表
// @Description 分页查询提交列表
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param request body commitModel.ReqFilter true "查询提交请求参数"
// @Success 200 {object} resp.ListResponse{data=commitModel.RespForList} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /commits/list [post]
func (c *Controller) List(ctx *gin.Context) {

	var reqQuery commitModel.ReqFilter
	var err error
	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, reqQuery.RepoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, isLast, err := c.commitSvc.List(ctx, reqQuery, *repoOut, reqQuery.GitRef)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListNoTotalArgs{
			IsLast:   isLast,
			Page:     int64(1),
			PageSize: int64(1),
		}, i18n.CommitListSuccess, nil)
}

// @Summary 获取单个提交详情
// @Description 根据提交 SHA 获取提交详情
// @Tags commits
// @Accept application/json
// @Produce application/json
// @Param sha query string true "提交 SHA"
// @Param repoID query uint true "仓库 ID"
// @Success 200 {object} resp.ObjectResponse{data=commitModel.Commit} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "提交未找到"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /commits [get]
func (c *Controller) Get(ctx *gin.Context) {

	var req commitModel.ReqForGet
	var err error
	if err = ctx.ShouldBindQuery(&req); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, req.RepoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 调用 service 层获取提交详情
	respData, err := c.commitSvc.Get(ctx.Request.Context(), req.SHA, *repoOut)
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
// @Param request body commitModel.ReqForDivergence true "获取分歧信息请求参数"
// @Success 200 {object} resp.ObjectResponse{data=[]commitModel.CommitDivergence} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /commits/divergence [post]
func (c *Controller) GetDivergence(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqBody commitModel.ReqForDivergence
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	log.Debugln(reqBody.RepoID)
	// repo, err := ctl.repoSvc.Get(ctx, session, repoRef, enum.PermissionRepoView)
	// if err != nil {
	// 	return nil, err
	// }

	// 调用 service 层获取提交分歧信息
	respData, err := c.commitSvc.GetDivergence(ctx.Request.Context(), reqBody, repoModel.Repo{})
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
// @Param request body commitModel.ReqForGet true "获取 diff 请求参数"
// @Success 200 {string} string "diff 文本"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "提交未找到"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /commits/diff [post]
func (c *Controller) GetDiff(ctx *gin.Context) {
	log := logger.GetLogger(ctx.Request.Context())

	var reqBody commitModel.ReqForGet
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	log.Debugln("Getting diff for commit:", reqBody.SHA)

	// 通过仓库 ID 获取仓库
	// TODO: 实际项目中需要通过 repoID 获取仓库信息并检查权限
	// repo, err := c.repoSvc.GetRepoCheckAccess(ctx, session, repoRef, []constant.Permission{constant.PermissionRepoView})
	// if err != nil {
	//     resp.InternalError(ctx, err)
	//     return
	// }

	// 设置响应头为 text/plain
	ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.NoCache(ctx.Writer)

	// 调用 service 层,直接将 diff 写入响应流
	err = c.commitSvc.GetDiff(ctx.Request.Context(), reqBody.SHA, repoModel.Repo{}, ctx.Writer)
	if err != nil {
		// 如果已经开始写入响应,则无法再使用标准的错误响应
		// 只能记录日志
		log.Errorf("Failed to get commit diff: %v", err)
		return
	}
}
