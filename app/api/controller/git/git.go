package git

import (
	"compress/gzip"
	"fmt"
	"net/url"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/git"
	repoSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	settingsSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/settings"
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
	settingsSvc settingsSvc.Interface
}

// NewGitController 创建 Git 处理器实例
func NewGitController(gitSvc git.Interface, repoSvc repoSvc.Interface) *GitController {
	return &GitController{gitSvc: gitSvc, repoSvc: repoSvc}
}

func (ctl *GitController) Clone(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.RepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
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

	// clone 是写操作 只要可读即可
	// isWriteOperation := false
	permission := constant.PermissionRepoView
	repo, err := ctl.repoSvc.GetRepoCheckAccess(ctx, session, params.RepoRef, []constant.Permission{permission})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPullPermissionError)
		resp.BadRequest(ctx, err)
		return
	}

	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeUploadPack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
	}
	_ = options
	_ = repo
	// if errors.Is(err, apiauth.ErrNotAuthorized) && auth.IsAnonymousSession(session) {
	// 	renderBasicAuth(ctx, w, urlProvider)
	// 	return
	// }

	// if errors.Is(err, store.ErrRequireFitClient) {
	// 	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	// 	w.Header().Set("X-Content-Type-Options", "nosniff")
	// 	w.WriteHeader(http.StatusForbidden)
	// 	_, _ = w.Write([]byte(err.Error()))
	// 	return
	// }

	if err != nil {
		// todo
		logger.Errorf("repoCtrl.GitServicePack failed: %v", err)
		resp.NotFound(ctx, err)
		return
	}
}

func (ctl *GitController) Push(ctx *gin.Context) {
	logger := logger.GetLogger(ctx)
	var params req.RepoRef
	var err error

	// 公共解析逻辑
	session, contentEncoding, gitProtocol, err := ctl.prepareRequest(ctx, &params)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestInvalidRequestBody)
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

	// clone 是写操作 只要可读即可
	// isWriteOperation := true
	permission := constant.PermissionRepoPush
	repo, err := ctl.repoSvc.GetRepoCheckAccess(ctx, session, params.RepoRef, []constant.Permission{permission})
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPushPermissionError)
		resp.BadRequest(ctx, err)
		return
	}

	need, err := ctl.settingsSvc.NeedFitClient(ctx, repo.ID, constant.KeyRequireFitClient)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.DbError)
		resp.InternalError(ctx, err)
		return
	}

	if need {
		// todo
	}

	options := api.ServicePackOptions{
		Service:      constant.GitServiceTypeUploadPack,
		StatelessRPC: true,
		Stdout:       ctx.Writer,
		Stdin:        dataReader,
		Protocol:     gitProtocol,
		IsWrite:      true,
	}

	_ = options
	_ = repo
	// if errors.Is(err, apiauth.ErrNotAuthorized) && auth.IsAnonymousSession(session) {
	// 	renderBasicAuth(ctx, w, urlProvider)
	// 	return
	// }

	// if errors.Is(err, store.ErrRequireFitClient) {
	// 	w.Header().Set("Content-Type", "text/plain;charset=utf-8")
	// 	w.Header().Set("X-Content-Type-Options", "nosniff")
	// 	w.WriteHeader(http.StatusForbidden)
	// 	_, _ = w.Write([]byte(err.Error()))
	// 	return
	// }

	if err != nil {
		// todo
		logger.Errorf("repoCtrl.GitServicePack failed: %v", err)
		resp.NotFound(ctx, err)
		return
	}
}

func (ctl *GitController) InfoRefs(c *gin.Context) {

}

// 提取公共解析逻辑
func (ctl *GitController) prepareRequest(ctx *gin.Context, params *req.RepoRef) (*auth.Session, string, string, error) {
	logger := logger.GetLogger(ctx)
	var err error

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		return nil, "", "", err
	}

	if err = ctx.ShouldBindUri(params); err != nil || params.RepoRef == "" {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		return nil, "", "", err
	}

	val, err := url.PathUnescape(params.RepoRef)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
		return nil, "", "", err
	}

	logger.Infof("origin url %s , repoRef: %s", ctx.Request.URL.String(), val)

	contentEncoding := req.GetHeaderOrDefault(ctx, req.HeaderContentEncoding, "")
	gitProtocol := req.GetHeaderOrDefault(ctx, req.HeaderParamGitProtocol, "")

	return session, contentEncoding, gitProtocol, nil
}
