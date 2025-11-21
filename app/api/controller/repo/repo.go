package repo

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	gitPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	urlPkg "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	repoSvc     repo.Interface
	urlProvider urlPkg.Provider
	gitPkg      gitPkg.Interface
}

// NewRepoController 创建代码库处理器实例
func NewRepoController(repoSvc repo.Interface, urlProvider urlPkg.Provider, gitPkg gitPkg.Interface) *Controller {
	return &Controller{
		repoSvc:     repoSvc,
		urlProvider: urlProvider,
		gitPkg:      gitPkg,
	}
}

// CreateRemoteRepo 创建远端仓库
// @Summary 创建远端仓库
// @Description 基于传入参数创建远端Git仓库，自动处理GitIgnore模板、License协议、README文件生成，支持派生仓库配置
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param body body repoModel.CreateRepoReq true "创建仓库请求参数"
// @Success 201 {object} resp.ObjectResponse{data=repoModel.CreateRepoResp} "创建成功（返回仓库核心信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如identifier格式非法、spaceID为空等）"
// @Error 404 {object} resp.ObjectResponse "空间不存在（传入的spaceID对应的空间未找到）"
// @Error 409 {object} resp.ObjectResponse "仓库标识冲突（同一空间内identifier已存在）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos [post]
func (ctl *Controller) Create(c *gin.Context) {
	reqData := new(repoModel.CreateRepoReq)
	if err := c.ShouldBindJSON(&reqData); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	err := reqData.Validate()
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	data, err := ctl.repoSvc.Create(logger.GetContextWithLogger(c), session, reqData)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoCreateError)
		resp.InternalError(c, err)
		return
	}

	resp.Created(c, i18n.RepoCreateSuccess, data)

}

// Get 获取代码库
// @Summary 获取仓库信息
// @Description 根据传入的仓库 ID 获取仓库信息
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Repo} "获取成功（返回仓库核心信息）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如identifier格式非法、ID为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在（传入的repoID对应的空间未找到）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{id} [get]
func (ctl *Controller) Get(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	repoID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(c, err)
		return
	}
	if repoID <= 0 {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(c, err)
		return
	}

	obj, err := ctl.repoSvc.GetByID(c, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryError)
		resp.InternalError(c, err)
		return
	}

	resp.OK(c, i18n.RepoQuerySuccess, obj)
}

// Update 修改仓库基础信息
// @Summary 修改仓库基础信息
// @Description 修改仓库基础信息
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param body body repoModel.ReqForUpdate true "更新仓库所需请求参数"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Repo} "更新成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "代码库不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id} [put]
func (c *Controller) Update(ctx *gin.Context) {

	var reqBody repoModel.ReqForUpdate
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
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
<<<<<<< HEAD

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
=======
	// TODO: 获取操作人
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
>>>>>>> 1543c33 (sync)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.repoSvc.Update(ctx, &reqBody, *repoOut, session)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoUpdateSuccess, respData)

}

func (ctl *Controller) GetContent(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.RepoId
	var err error

	_, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindUri(&params); err != nil || params.RepoID == 0 {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		resp.BadRequest(ctx, err)
		return
	}

	// gitRef 仓库引用，例如：refs/heads/main
	gitRef := req.QueryParamOrDefault(ctx, req.QueryParamGitRef, "refs/heads/main")

	// includeCommit 是否包含提交信息
	includeCommit, err := req.QueryParamAsBoolOrDefault(ctx, req.QueryParamIncludeCommit, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPullPermissionError)
		resp.BadRequest(ctx, err)
		return
	}

	// flattenDirectories 是否展平目录结构
	flattenDirectories, err := req.QueryParamAsBoolOrDefault(ctx, req.QueryParamFlattenDirectories, false)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 在像 /v1/repos/{repoRef}/contents/* 这样的路由下
	// * 表示仓库内的文件或目录的相对路径。
	// 例如请求 /v1/repos/12/contents/src/main.go ，这里的余量就是 src/main.go
	path := req.PathParamOrEmpty(ctx, req.PathParamRemainder)

	logger.Infof("origin url %s , repoID: %d,gitRef: %s, path: %s,includeCommit: %v,flattenDirectories: %v", ctx.Request.URL.String(), params.RepoID, gitRef, path, includeCommit, flattenDirectories)

	fmt.Println(ctx.Param("path"))
}

// @Summary 仓库代码blame
// @Description 根据传入的仓库 ID、文件路径、Git 引用、行号范围等信息，获取仓库代码blame信息
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param repoId path uint true "仓库 ID"
// @Param filePath path string true "文件路径" e.g. src/main.go
// @Param gitRef query string false "Git 引用" e.g. refs/heads/main
// @Param lineFrom query int false "起始行号" e.g. 1
// @Param lineTo query int false "结束行号" e.g. 100
// @Success 200 {object} resp.ObjectResponse{data=interface{}} "获取成功（返回仓库代码blame流式）"
// @Error 400 {object} resp.ObjectResponse "参数错误（如identifier格式非法、ID为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在（传入的repoID对应的空间未找到）"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{repoId}/blame/{filePath} [get]
func (ctl *Controller) Blame(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	repoID, err := req.PathParamAsPositiveUint(c, req.QueryParamRepoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(c, err)
		return
	}

	filePath, err := req.PathParamOrError(c, req.PathParamFilePath)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(c, err)
		return
	}

	var BlameFilter repoModel.BlameFilter
	if err = c.ShouldBindJSON(&BlameFilter); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryError)
		resp.BadRequest(c, err)
		return
	}

	if BlameFilter.LineTo > 0 && BlameFilter.LineFrom > BlameFilter.LineTo {
		err = errors.WithI18nKey(err, i18n.RepoLineNumberError)
		resp.BadRequest(c, err)
		return
	}

	repoInfo, err := ctl.repoSvc.GetByID(c, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryError)
		resp.InternalError(c, err)
		return
	}

	if BlameFilter.GitRef == "" {
		BlameFilter.GitRef = repoInfo.DefaultBranch
	}

	gitPkg.NewStreamReader(ctl.gitPkg.Blame(c, &gitPkg.BlameParams{
		ReadParams: gitPkg.ReadParams{RepoUID: repoInfo.GitUID},
		Path:       filePath,
		GitRef:     BlameFilter.GitRef,
		LineFrom:   BlameFilter.LineFrom,
		LineTo:     BlameFilter.LineTo,
	}))
}

// GetDiff 获取仓库指定 GitRef Diff
// @Summary 获取仓库指定 GitRef Diff
// @Description 获取仓库指定 GitRef Diff
// @Tags repos
// @Accept application/json
// @Produce text/plain
// @Param id path uint true "仓库 ID"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Repo} "获取成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "代码库不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id}/diff [post]
func (c *Controller) GetDiff(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqBody repoModel.ReqForDiff
	var err error

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SpacePathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.CommitBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	repoOut, err := c.repoSvc.GetByID(ctx, session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 设置响应头为 text/plain
	ctx.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.NoCache(ctx.Writer)

	// 调用 service 层,直接将 diff 写入响应流
	err = c.repoSvc.GetDiff(ctx.Request.Context(), reqBody, *repoOut, ctx.Writer)
	if err != nil {
		// 如果已经开始写入响应,则无法再使用标准的错误响应
		// 只能记录日志
		log.Errorf("Failed to get diff: %v", err)
		return
	}
}

// GetPaths 获取仓库指定 GitRef 路径
// @Summary 获取仓库指定 GitRef 路径
// @Description 获取仓库指定 GitRef 路径
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param gitRef query string false "Git 引用"
// @Param includeDirectories query bool false "包含目录"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.PathsOutput} "获取成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "代码库不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id}/paths [get]
func (c *Controller) GetPaths(ctx *gin.Context) {

	var reqQuery repoModel.PathsInput
	var err error

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.repoSvc.Paths(ctx, reqQuery, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoGetPathsSuccess, respData)

}

// PathsDetails 获取仓库指定 GitRef 路径详情
// @Summary 获取仓库指定 GitRef 路径详情
// @Description 获取仓库指定 GitRef 路径详情
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param body body repoModel.PathsDetailsInput true "查询路径详情的请求参数"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.PathsDetailsOutput} "获取成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 404 {object} resp.ObjectResponse "代码库不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id}/paths/details [post]
func (c *Controller) GetPathsDetails(ctx *gin.Context) {

	var reqBody repoModel.PathsDetailsInput
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
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	if len(reqBody.Paths) == 0 || len(reqBody.Paths) > 5000 {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.repoSvc.PathsDetails(ctx, reqBody, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoGetPathsDetailsSuccess, respData)

}
