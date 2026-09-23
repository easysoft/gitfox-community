// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package git

import (
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/easysoft/gitfox-community/app/api/controller/githook"
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/app/model/auth"
	githookModel "github.com/easysoft/gitfox-community/app/model/githook"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/app/service/git"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	settingsSvc "github.com/easysoft/gitfox-community/app/service/settings"
	spaceSvc "github.com/easysoft/gitfox-community/app/service/space"
	"github.com/easysoft/gitfox-community/pkg/errors"
	gitPkg "github.com/easysoft/gitfox-community/pkg/git"
	"github.com/easysoft/gitfox-community/pkg/git/api"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GitController 处理与 Git 相关的 HTTP 请求
type GitController struct {
	gitSvc      git.Interface
	repoSvc     repoSvc.Interface
	spaceSvc    spaceSvc.Interface
	settingsSvc settingsSvc.Interface
	githookCtrl GithookController
	git         gitPkg.Interface
}

// GithookController defines the interface for githook operations
type GithookController interface {
	PreReceive(ctx context.Context, rgit githook.RestrictedGIT, in githookModel.GithookPreReceiveInput) (hook.Output, error)
	PostReceive(ctx context.Context, rgit githook.RestrictedGIT, in githookModel.GithookPostReceiveInput) (hook.Output, error)
	Update(ctx context.Context, rgit githook.RestrictedGIT, in githookModel.GithookUpdateInput) (hook.Output, error)
}

// NewGitController 创建 Git 处理器实例
func NewGitController(gitSvc git.Interface, spaceSvc spaceSvc.Interface, repoSvc repoSvc.Interface, settingsSvc settingsSvc.Interface, githookCtrl GithookController, gitPkgInterface gitPkg.Interface) *GitController {
	return &GitController{gitSvc: gitSvc, spaceSvc: spaceSvc, repoSvc: repoSvc, settingsSvc: settingsSvc, githookCtrl: githookCtrl, git: gitPkgInterface}
}

// Clone 处理 Git clone/fetch 操作 (upload-pack)
// @Summary Git Clone/Fetch 操作
// @Description 处理 Git 客户端的 clone 和 fetch 请求，通过 upload-pack 协议发送仓库对象数据
// @Tags git
// @Accept application/x-git-upload-pack-request
// @Produce application/x-git-upload-pack-result
// @Param spaceCode path string true "空间标识码"
// @Param repoRef path string true "仓库标识 (支持 .git 后缀)"
// @Param Content-Encoding header string false "内容编码 (如 gzip)"
// @Param Git-Protocol header string false "Git 协议版本"
// @Success 200 {string} string "成功返回 Git pack 数据"
// @Failure 400 {object} resp.ObjectResponse "参数错误或权限不足"
// @Failure 404 {object} resp.ObjectResponse "仓库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /git/{spaceCode}/{repoRef}/git-upload-pack [post]
func (ctl *GitController) Clone(ctx *gin.Context) {

	resp.NoCache(ctx.Writer)
	ctx.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", constant.GitServiceTypeUploadPack))

	logger := logger.GetLogger(ctx)
	var params req.SpaceRepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		resp.GitError(ctx, err, http.StatusBadRequest, false)
		return
	}
	if session.User.Email == "" {
		err = errors.WithI18nKey(errors.New("user email is empty"), i18n.RepoEmailRequired)
		resp.GitError(ctx, err, http.StatusUnauthorized, false)
		return
	}

	// 去除 .git 后缀
	repoIdentifier := strings.TrimSuffix(params.RepoRef, ".git")

	// 获取仓库信息
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceCode, repoIdentifier, constant.PermissionRepoView)
	if err != nil {
		resp.GitError(ctx, err, http.StatusBadRequest, false)
		return
	}

	dataReader := ctx.Request.Body
	if contentEncoding == "gzip" {
		gzipReader, gErr := gzip.NewReader(dataReader)
		if gErr != nil {
			gErr = errors.WithI18nKey(gErr, i18n.ErrorGzipFailedToCreateReader)
			resp.GitError(ctx, gErr, http.StatusInternalServerError, false)
			return
		}

		defer func() {
			if cErr := gzipReader.Close(); cErr != nil {
				logger.Warnf("failed to close the gzip reader : %v", cErr)
			}
		}()
		dataReader = gzipReader
	}

	// 调用 ServicePack 处理 Git 协议
	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeUploadPack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stderr:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
	}

	err = ctl.gitSvc.GitServicePack(ctx, session, repo, options)
	if err != nil {
		logger.Errorf("GitServicePack failed: %v", err)
		resp.GitError(ctx, err, http.StatusInternalServerError, false)
		return
	}
}

// Push 处理 Git push 操作 (receive-pack)
// @Summary Git Push 操作
// @Description 处理 Git 客户端的 push 请求，通过 receive-pack 协议接收仓库对象数据并更新引用
// @Tags git
// @Accept application/x-git-receive-pack-request
// @Produce application/x-git-receive-pack-result
// @Param spaceCode path string true "空间标识码"
// @Param repoRef path string true "仓库标识 (支持 .git 后缀)"
// @Param Content-Encoding header string false "内容编码 (如 gzip)"
// @Param Git-Protocol header string false "Git 协议版本"
// @Success 200 {string} string "成功接收并处理 Git pack 数据"
// @Failure 400 {object} resp.ObjectResponse "参数错误或权限不足"
// @Failure 404 {object} resp.ObjectResponse "仓库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /git/{spaceCode}/{repoRef}/git-receive-pack [post]
func (ctl *GitController) Push(ctx *gin.Context) {

	resp.NoCache(ctx.Writer)
	ctx.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", constant.GitServiceTypeReceivePack))

	logger := logger.GetLogger(ctx)
	var params req.SpaceRepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestInvalidRequestBody)
		resp.GitError(ctx, err, http.StatusBadRequest, true)
		return
	}

	if session.User.Email == "" {
		err = errors.WithI18nKey(errors.New("user email is empty"), i18n.RepoEmailRequired)
		resp.GitError(ctx, err, http.StatusUnauthorized, true)
		return
	}

	// 去除 .git 后缀
	repoIdentifier := strings.TrimSuffix(params.RepoRef, ".git")

	// 获取仓库信息
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceCode, repoIdentifier, constant.PermissionRepoPush)
	if err != nil {
		resp.GitError(ctx, err, http.StatusBadRequest, true)
		return
	}

	dataReader := ctx.Request.Body
	if contentEncoding == "gzip" {
		gzipReader, gErr := gzip.NewReader(dataReader)
		if gErr != nil {
			gErr = errors.WithI18nKey(gErr, i18n.ErrorGzipFailedToCreateReader)
			resp.GitError(ctx, gErr, http.StatusInternalServerError, true)
			return
		}

		defer func() {
			if cErr := gzipReader.Close(); cErr != nil {
				logger.Warnf("failed to close the gzip reader : %v", cErr)
			}
		}()
		dataReader = gzipReader
	}

	// need, err := ctl.settingsSvc.NeedFitClient(ctx, repo.ID, constant.KeyRequireFitClient)
	// if err != nil {
	// 	err = errors.WithI18nKey(err, i18n.DbError)
	// 	resp.GitError(ctx, err, http.StatusInternalServerError, true)
	// 	return
	// }

	// 调用 ServicePack 处理 Git 协议
	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeReceivePack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stderr:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
		IsWrite:      true,
		FitClient:    false,
	}

	err = ctl.gitSvc.GitServicePack(ctx, session, repo, options)
	if err != nil {
		logger.Errorf("GitServicePack failed: %v", err)
		resp.GitError(ctx, err, http.StatusInternalServerError, true)
		return
	}
}

// InfoRefs 处理 Git info/refs 请求
// @Summary Git Info/Refs 请求
// @Description 处理 Git 客户端的 info/refs 请求，返回仓库的引用信息(分支、标签等)，用于 Smart HTTP 协议的初始握手
// @Tags git
// @Accept */*
// @Produce application/x-git-upload-pack-advertisement,application/x-git-receive-pack-advertisement
// @Param spaceCode path string true "空间标识码"
// @Param repoRef path string true "仓库标识 (支持 .git 后缀)"
// @Param service query string true "Git 服务类型" Enums(git-upload-pack, git-receive-pack)
// @Param Git-Protocol header string false "Git 协议版本"
// @Success 200 {string} string "成功返回引用广告信息"
// @Failure 400 {object} resp.ObjectResponse "参数错误或权限不足"
// @Failure 404 {object} resp.ObjectResponse "仓库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器内部错误"
// @Router /git/{spaceCode}/{repoRef}/info/refs [get]
func (ctl *GitController) InfoRefs(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.SpaceRepoRef

	// 解析参数
	session, _, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	// 获取 service 参数
	service := ctx.Query("service")
	if service == "" {
		err = errors.WithI18nKey(errors.New("service parameter is required"), i18n.RequestInvalidRequestBody)
		resp.BadRequest(ctx, err)
		return
	}

	// Git 客户端会发送 "git-upload-pack" 或 "git-receive-pack"
	// 但 git 命令只接受 "upload-pack" 或 "receive-pack"
	// 需要去除 "git-" 前缀
	const gitPrefix = "git-"
	gitService := service
	if strings.HasPrefix(service, gitPrefix) {
		gitService = service[len(gitPrefix):]
	}

	// 去除 .git 后缀
	repoIdentifier := strings.TrimSuffix(params.RepoRef, ".git")

	logger.Infof("Parsed repoRef: %s, repoIdentifier: %s, service: %s, gitService: %s", params.RepoRef, repoIdentifier, service, gitService)

	// 获取仓库信息
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceCode, repoIdentifier, constant.PermissionRepoView)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	// 调用 git service 获取 info/refs
	data, err := ctl.gitSvc.GetInfoRefs(ctx, repo.GitUID, gitService, gitProtocol)
	if err != nil {
		logger.Errorf("GetInfoRefs failed: %v", err)
		resp.InternalError(ctx, err)
		return
	}

	// 设置响应头
	resp.NoCache(ctx.Writer)
	ctx.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-advertisement", gitService))

	// 写入响应
	if _, err := ctx.Writer.Write(data); err != nil {
		logger.Errorf("Failed to write InfoRefs response: %v", err)
	}
}

// getRepoByIdentifier 通过 spaceCode + identifier 获取仓库并检查权限
func (ctl *GitController) getRepoByIdentifier(ctx *gin.Context, session *auth.Session, spaceCode string, identifier string, permission constant.Permission) (*repoModel.Repo, error) {
	// 检查空间权限
	space, err := ctl.spaceSvc.GetSpaceCheckAccess(ctx, session, spaceCode, []constant.Permission{constant.PermissionSpaceView})
	if err != nil {
		return nil, err
	}

	// 通过 identifier 查询仓库
	repo, err := ctl.repoSvc.GetRepoCheckAccess(ctx, session, space.ID, identifier, []constant.Permission{permission})
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.RepoQueryError)
	}

	return repo, nil
}

// 提取公共解析逻辑
func (ctl *GitController) prepareRequest(ctx *gin.Context, params *req.SpaceRepoRef) (*auth.Session, string, string, error) {
	logger := logger.GetLogger(ctx)
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		return nil, "", "", err
	}

	if err = ctx.ShouldBindUri(params); err != nil {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		return nil, "", "", err
	}

	logger.Infof("origin url %s , spaceCode: %s, repoRef: %s", ctx.Request.URL.String(), params.SpaceCode, params.RepoRef)

	contentEncoding := req.GetHeaderOrDefault(ctx, req.HeaderContentEncoding, "")
	gitProtocol := req.GetHeaderOrDefault(ctx, req.HeaderParamGitProtocol, "")

	return session, contentEncoding, gitProtocol, nil
}
