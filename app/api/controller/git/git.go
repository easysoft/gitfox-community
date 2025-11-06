package git

import (
	"compress/gzip"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// GitController 处理与 Git 相关的 HTTP 请求
type GitController struct {
	gitSvc git.IGit
}

// NewGitController 创建 Git 处理器实例
func NewGitController(gitSvc git.IGit) *GitController {
	return &GitController{gitSvc: gitSvc}
}

func (ctl *GitController) Clone(c *gin.Context) {
	var ctx = c.Request.Context()

	//var err error

	session, ok := req.GetAuthSession(ctx)
	if !ok {
		err := errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.BadRequest(c, err)
		return
	}

	_ = session

	// 解析仓库引用（复用现有解析逻辑，入参为 *http.Request）
	// repoRef, err := request.GetRepoRefFromPath(c.Request)
	// if err != nil {
	// 	err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
	// 	resp.BadRequest(c, err)
	// 	return
	// }

	// _ = repoRef

	// 从 Header 获取编码与 Git 协议
	contentEncoding := req.GetHeaderOrDefault(c, req.HeaderContentEncoding, "")
	gitProtocol := req.GetHeaderOrDefault(c, req.HeaderParamGitProtocol, "")

	_ = gitProtocol
	// Handle GZIP.
	dataReader := c.Request.Body
	if contentEncoding == "gzip" {
		gzipReader, err := gzip.NewReader(dataReader)
		if err != nil {
			err = errors.WithI18nKey(err, i18n.ErrorGzipFailedToCreateReader)
			resp.InternalError(c, err)
			return
		}
		defer func() {
			if cErr := gzipReader.Close(); cErr != nil {
				//log.Ctx(ctx).Warn().Err(cErr).Msg("failed to close the gzip reader")
			}
		}()
		dataReader = gzipReader
	}

	// Clients MUST NOT reuse or revalidate a cached response.
	// Servers MUST include sufficient Cache-Control headers to prevent caching of the response.
	// https://git-scm.com/docs/http-protocol
	resp.NoCache(c.Writer)
	c.Writer.Header().Set("Content-Type", fmt.Sprintf("application/x-git-%s-result", constant.GitServiceTypeUploadPack))

	// err = ctl.gitSvc.GitServicePack(ctx, session, repoRef, api.ServicePackOptions{
	// 	Service:      service,
	// 	StatelessRPC: true,
	// 	Stdout:       w,
	// 	Stdin:        dataReader,
	// 	Protocol:     gitProtocol,
	// })

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

	// if err != nil {
	// 	log.Ctx(ctx).Error().Err(err).Msg("repoCtrl.GitServicePack")
	// 	render.TranslatedUserError(ctx, w, err)
	// 	return
	// }

}

func (ctl *GitController) Push(c *gin.Context) {

}

func (ctl *GitController) InfoRefs(c *gin.Context) {

}
