// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package repo

import (
	"fmt"
	"strings"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	artifactModel "github.com/easysoft/gitfox-community/app/model/artifact"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	artifactSvc "github.com/easysoft/gitfox-community/app/service/artifact"
	"github.com/easysoft/gitfox-community/app/service/importer"
	"github.com/easysoft/gitfox-community/app/service/pullrequest"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	gitPkg "github.com/easysoft/gitfox-community/pkg/git"
	gitAPI "github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
	urlPkg "github.com/easysoft/gitfox-community/pkg/utils/url"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	repoSvc        repo.Interface
	importerSvc    importer.Interface
	pullrequestSvc pullrequest.Interface
	urlProvider    urlPkg.Provider
	gitPkg         gitPkg.Interface
	artSvc         artifactSvc.Interface
}

// NewRepoController 创建代码库处理器实例
func NewRepoController(repoSvc repo.Interface, importerSvc importer.Interface, pullrequestSvc pullrequest.Interface, urlProvider urlPkg.Provider, gitPkg gitPkg.Interface, artSvc artifactSvc.Interface) *Controller {
	return &Controller{
		repoSvc:        repoSvc,
		importerSvc:    importerSvc,
		pullrequestSvc: pullrequestSvc,
		urlProvider:    urlProvider,
		gitPkg:         gitPkg,
		artSvc:         artSvc,
	}
}

// CreateRemoteRepo 创建远端仓库
// @Summary 创建远端仓库
// @Description 基于传入参数创建远端Git仓库，自动处理GitIgnore模板、License协议、README文件生成，支持派生仓库配置
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param body body repoModel.CreateRepoReq true "创建仓库请求参数"
// @Success 201 {object} resp.ObjectResponse{data=repoModel.Repo} "创建成功（返回仓库核心信息）"
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
		resp.BadRequest(c, err)
		return
	}

	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	if session.User.Email == "" {
		resp.BadRequest(c, errors.WithI18nKey(nil, i18n.RepoEmailRequired))
		return
	}

	data, err := ctl.repoSvc.Create(c.Request.Context(), session, reqData)
	if err != nil {
		i18nKey := errors.GetI18nKey(err)
		if i18nKey == i18n.DbDuplicate {
			err = errors.WithI18nKey(err, i18n.RepoCreateDuplicate)
			resp.BadRequest(c, err)
			return
		}
		resp.InternalError(c, err)
		return
	}

	// 创建仓库时自动创建默认 generic 制品库
	if ctl.artSvc != nil {
		_, _ = ctl.artSvc.CreateLib(c.Request.Context(), &artifactModel.CreateArtifactLibReq{
			SpaceRepoReq: artifactModel.SpaceRepoReq{SpaceID: int64(reqData.SpaceID), RepoID: int64(data.ID)},
			Name:         artifactModel.DefaultArtifactLibViewName(i18n.GetLocaleFromGin(c)),
			Code:         artifactModel.DefaultArtifactLibCode,
			Format:       string(artifactModel.ArtifactFileFormat),
		})
	}

	resp.Created(c, i18n.RepoCreateSuccess, data)

}

// List 查询仓库列表
// @Summary 查询仓库列表
// @Description 分页查询仓库列表
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param request body repoModel.ReqFilter true "查询仓库请求参数"
// @Success 200 {object} resp.ListResponse{data=[]repoModel.Repo,listArgs=resp.ListNoTotalArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/list [post]
func (c *Controller) List(ctx *gin.Context) {
	var reqQuery repoModel.ReqFilter
	var err error

	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	// 设置默认值
	req.SetDefaults(&reqQuery)

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	respData, isLast, err := c.repoSvc.List(ctx.Request.Context(), session, reqQuery)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListNoTotalArgs{
			IsLast:   isLast,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.PageSize),
		}, i18n.RepoQuerySuccess, nil)
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

	respData, err := c.repoSvc.Update(ctx, &reqBody, *repoOut, session)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoUpdateSuccess, respData)

}

// Delete 删除仓库，数据库软删除
// @Summary 删除仓库
// @Description 删除仓库,软删除
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Success 200 {object} resp.ObjectResponse "删除仓库成功"
// @Error 400 {object} resp.ObjectResponse "参数错误"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id} [delete]
func (c *Controller) Delete(ctx *gin.Context) {
	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	err = c.repoSvc.Delete(ctx.Request.Context(), session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoDeleteError)
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoDeleteSuccess, nil)
}

// GetContent 获取仓库内容
// @Summary 获取仓库文件或目录内容
// @Description 根据路径获取仓库中的文件内容、目录列表或子模块信息，支持指定 Git 引用(分支/标签/提交)。路径参数可选，不传或传空则返回根目录
// @Tags repos
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param filepath path string false "文件或目录路径 (如 src/main.go)，可选，为空则返回根目录"
// @Param gitRef query string false "Git 引用 (分支/标签/提交SHA)，默认为 refs/heads/main"
// @Param includeCommit query bool false "是否包含最新提交信息，默认 false"
// @Param flattenDirectories query bool false "是否展平单层目录结构，默认 false"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.GetContentOutput} "获取成功，返回文件/目录/符号链接/子模块内容"
// @Failure 400 {object} resp.ObjectResponse "参数错误或权限不足"
// @Failure 404 {object} resp.ObjectResponse "仓库或文件不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /repos/{id}/content [get]
// @Router /repos/{id}/content/{filepath} [get]
func (ctl *Controller) GetContent(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var repoID uint
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		resp.Unauthorized(ctx, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	repoID, err = req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
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
	// 注意：通配符参数会包含前导斜杠，需要去除（例如 "/src/main.go" -> "src/main.go"）
	path := strings.TrimPrefix(req.PathParamOrEmpty(ctx, req.PathParamFilePath), "/")
	logger.Infof("origin url %s , repoID: %d,gitRef: %s, path: %s,includeCommit: %v,flattenDirectories: %v", ctx.Request.URL.String(), repoID, gitRef, path, includeCommit, flattenDirectories)

	// todo check spcace ?

	// get repo
	repoM, err := ctl.repoSvc.GetRepoCheckAccessByID(ctx.Request.Context(), session, repoID, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoGetContentPermissionError)
		resp.Unauthorized(ctx, err)
		return
	}

	re, err := ctl.repoSvc.GetContent(ctx.Request.Context(), session, repoM, gitRef, path, includeCommit, flattenDirectories)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoGetContentError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.RepoGetContentSuccess, re)
}

// Archive 获取仓库归档文件
// @Summary 获取仓库归档文件(tar/zip/tar.gz/tgz)
// @Description 根据指定的 Git 引用和格式生成并下载仓库归档文件。通过 URL 路径传递 Git 引用和格式
// @Description URL 格式: /archive/{ref}.{format}
// @Description 示例: /archive/main.zip, /archive/refs/heads/main.tar.gz, /archive/v1.0.0.zip
// @Tags repos
// @Accept application/json
// @Produce application/tar,application/zip,application/gzip
// @Param id path uint true "仓库 ID"
// @Param gitRef path string true "Git 引用和文件格式(如 main.zip、refs/heads/main.tar.gz、v1.0.0.zip),格式由扩展名决定"
// @Param path query []string false "可选的路径列表,仅归档指定的文件/目录。不指定则归档整个仓库"
// @Param prefix query string false "归档文件中路径前缀,会添加到所有文件路径前"
// @Param attributes query string false "Git 属性(export-ignore 或 export-subst)"
// @Param time query string false "归档条目的修改时间,格式为 RFC3339(如 2006-01-02 15:04:05)"
// @Param compression query int false "压缩级别,zip 格式为 0-9,tar.gz/tgz 格式支持数字压缩级别"
// @Success 200 {file} binary "归档文件,文件名格式为 {仓库标识}_{Git引用}_{格式}"
// @Failure 400 {object} resp.ObjectResponse "参数错误(如格式不支持、Git引用无效等)"
// @Failure 401 {object} resp.ObjectResponse "未授权或无会话"
// @Failure 404 {object} resp.ObjectResponse "代码库不存在或 Git 引用不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误(如 Git 服务调用失败、数据库异常等)"
// @Router /repos/{id}/archive/{gitRef} [get]
func (c *Controller) Archive(ctx *gin.Context) {

	log := logger.GetLogger(ctx)

	// Debug: log request URL and all params
	log.Infof("Archive request - URL: %s, Path: %s", ctx.Request.URL.String(), ctx.Request.URL.Path)
	for _, param := range ctx.Params {
		log.Infof("Archive request - Param: %s = %s", param.Key, param.Value)
	}

	var repoID uint
	var err error

	repoID, err = req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), nil, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	params, filename, err := req.ParseArchiveParams(ctx)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var contentType string
	switch params.Format {
	case gitAPI.ArchiveFormatTar:
		contentType = "application/tar"
	case gitAPI.ArchiveFormatZip:
		contentType = "application/zip"
	case gitAPI.ArchiveFormatTarGz, gitAPI.ArchiveFormatTgz:
		contentType = "application/gzip"
	default:
		contentType = "application/zip"
	}

	filename = repoOut.Identifier + "_" + filename
	ctx.Writer.Header().Set("Content-Type", contentType)
	ctx.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	err = c.repoSvc.Archive(ctx.Request.Context(), params, *repoOut, ctx.Writer)
	if err != nil {
		log.Errorf("Failed to get archive: %v", err)
		return
	}

}

// @Summary 仓库代码blame
// @Description 根据传入的仓库 ID、文件路径、Git 引用、行号范围等信息，获取仓库代码blame信息
// @Tags repos
// @Accept application/json
// @Produce text/plain
// @Param repoId path uint true "仓库 ID"
// @Param filePath path string true "文件路径" e.g. src/main.go
// @Param gitRef query string false "Git 引用" e.g. refs/heads/main
// @Param lineFrom query int false "起始行号" e.g. 1
// @Param lineTo query int false "结束行号" e.g. 100
// @Success 200 {string} string "blame 文本"
// @Error 400 {object} resp.ObjectResponse "参数错误（如ID为空等）"
// @Error 404 {object} resp.ObjectResponse "代码库不存在"
// @Error 500 {object} resp.ObjectResponse "服务器内部错误（如Git服务调用失败、数据库异常等）"
// @Router /repos/{repoId}/blame/{filePath} [get]
func (ctl *Controller) Blame(c *gin.Context) {
	log := logger.GetLogger(c.Request.Context())

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

	// 获取文件路径，并去除前导斜杠
	// 通配符参数会包含前导斜杠，需要去除（例如 "/test.php" -> "test.php"）
	filePath := strings.TrimPrefix(req.PathParamOrEmpty(c, req.PathParamFilePath), "/")
	if filePath == "" {
		err = errors.WithI18nKey(errors.New("file path is required"), i18n.RepoPathParamError)
		resp.BadRequest(c, err)
		return
	}

	var BlameFilter repoModel.BlameFilter
	if err = c.ShouldBindQuery(&BlameFilter); err != nil {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	if BlameFilter.LineTo > 0 && BlameFilter.LineFrom > BlameFilter.LineTo {
		err = errors.WithI18nKey(err, i18n.RepoBodyParamError)
		resp.BadRequest(c, err)
		return
	}

	repoInfo, err := ctl.repoSvc.GetByID(c, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryError)
		resp.InternalError(c, err)
		return
	}

	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	resp.NoCache(c.Writer)

	// 调用 service 层，直接将 blame 输出写入响应流
	err = ctl.repoSvc.GetBlame(c.Request.Context(), BlameFilter, *repoInfo, filePath, c.Writer)
	if err != nil {
		// 如果已经开始写入响应,则无法再使用标准的错误响应
		// 只能记录日志
		log.Errorf("Failed to get blame: %v", err)
		return
	}
}

// GetDiff 获取仓库指定 GitRef Diff
// @Summary 获取仓库指定 GitRef Diff
// @Description 获取仓库指定 GitRef Diff
// @Tags repos
// @Accept application/json
// @Produce text/plain
// @Param id path uint true "仓库 ID"
// @Param body body repoModel.ReqForDiff true "查看 gitRef diff 所需请求参数"
// @Success 200 {string} string "diff 文本"
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
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
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
