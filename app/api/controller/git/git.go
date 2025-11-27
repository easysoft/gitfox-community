package git

import (
	"compress/gzip"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/git"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	settingsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/settings"
	spaceSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/space"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"github.com/gin-gonic/gin"
)

// GitController 处理与 Git 相关的 HTTP 请求
type GitController struct {
	gitSvc      git.Interface
	repoSvc     repoSvc.Interface
	spaceSvc    spaceSvc.Interface
	settingsSvc settingsSvc.Interface
}

// NewGitController 创建 Git 处理器实例
func NewGitController(gitSvc git.Interface, spaceSvc spaceSvc.Interface, repoSvc repoSvc.Interface) *GitController {
	return &GitController{gitSvc: gitSvc, spaceSvc: spaceSvc, repoSvc: repoSvc}
}

func (ctl *GitController) Clone(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.SpaceRepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	// 去除 .git 后缀
	repoIdentifier := strings.TrimSuffix(params.RepoRef, ".git")

	// 获取仓库信息
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceID, repoIdentifier, constant.PermissionRepoView)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	dataReader := ctx.Request.Body
	if contentEncoding == "gzip" {
		gzipReader, gErr := gzip.NewReader(dataReader)
		if gErr != nil {
			gErr = errors.WithI18nKey(gErr, i18n.ErrorGzipFailedToCreateReader)
			resp.InternalError(ctx, gErr)
			return
		}

		defer func() {
			if cErr := gzipReader.Close(); cErr != nil {
				logger.Warnf("failed to close the gzip reader : %v", cErr)
			}
		}()
		dataReader = gzipReader
	}

	resp.NoCache(ctx.Writer)
	ctx.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", constant.GitServiceTypeUploadPack))

	// 调用 ServicePack 处理 Git 协议
	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeUploadPack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
	}

	err = ctl.gitSvc.GitServicePack(ctx, repo.GitUID, options)
	if err != nil {
		logger.Errorf("GitServicePack failed: %v", err)
		resp.InternalError(ctx, err)
		return
	}
}

func (ctl *GitController) Push(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.SpaceRepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestInvalidRequestBody)
		resp.BadRequest(ctx, err)
		return
	}

	// 去除 .git 后缀
	repoIdentifier := strings.TrimSuffix(params.RepoRef, ".git")

	// 获取仓库信息
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceID, repoIdentifier, constant.PermissionRepoPush)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	dataReader := ctx.Request.Body
	if contentEncoding == "gzip" {
		gzipReader, gErr := gzip.NewReader(dataReader)
		if gErr != nil {
			gErr = errors.WithI18nKey(gErr, i18n.ErrorGzipFailedToCreateReader)
			resp.InternalError(ctx, gErr)
			return
		}

		defer func() {
			if cErr := gzipReader.Close(); cErr != nil {
				logger.Warnf("failed to close the gzip reader : %v", cErr)
			}
		}()
		dataReader = gzipReader
	}

	resp.NoCache(ctx.Writer)
	ctx.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", constant.GitServiceTypeReceivePack))

	need, err := ctl.settingsSvc.NeedFitClient(ctx, repo.ID, constant.KeyRequireFitClient)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.DbError)
		resp.InternalError(ctx, err)
		return
	}

	// 调用 ServicePack 处理 Git 协议
	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeReceivePack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
		IsWrite:      true,
		FitClient:    need,
	}

	err = ctl.gitSvc.GitServicePack(ctx, repo.GitUID, options)
	if err != nil {
		logger.Errorf("GitServicePack failed: %v", err)
		resp.InternalError(ctx, err)
		return
	}
}

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
	repo, err := ctl.getRepoByIdentifier(ctx, session, params.SpaceID, repoIdentifier, constant.PermissionRepoView)
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

// getRepoByIdentifier 通过 identifier 获取仓库并检查权限
func (ctl *GitController) getRepoByIdentifier(ctx *gin.Context, session *auth.Session, spaceID uint, identifier string, permission constant.Permission) (*repoModel.Repo, error) {
	// 检查空间权限
	_, err := ctl.spaceSvc.GetSpaceCheckAccess(ctx, session, spaceID, []constant.Permission{constant.PermissionSpaceView})
	if err != nil {
		return nil, errors.WithI18nKey(err, i18n.SpaceQueryError)
	}

	// 通过 identifier 查询仓库
	repo, err := ctl.repoSvc.GetRepoCheckAccess(ctx, session, spaceID, identifier, []constant.Permission{permission})
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

	if err = ctx.ShouldBindUri(params); err != nil || params.SpaceID == 0 {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		return nil, "", "", err
	}

	logger.Infof("origin url %s , spaceId: %d, repoRef: %s", ctx.Request.URL.String(), params.SpaceID, params.RepoRef)

	contentEncoding := req.GetHeaderOrDefault(ctx, req.HeaderContentEncoding, "")
	gitProtocol := req.GetHeaderOrDefault(ctx, req.HeaderParamGitProtocol, "")

	return session, contentEncoding, gitProtocol, nil
}
