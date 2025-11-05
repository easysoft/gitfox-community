package git

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/git"
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
	//var err error

	// 解析仓库引用（复用现有解析逻辑，入参为 *http.Request）
	// repoRef, err := request.GetRepoRefFromPath(c.Request)
	// if err != nil {
	// 	err = errors.WithI18nKey(err, i18n.RequestPathParameterNotFound)
	// 	resp.BadRequest(c, err)
	// 	return
	// }

	// _ = repoRef

	// 从 Header 获取编码与 Git 协议（使用我们自己的 req 工具）
	contentEncoding := req.GetHeaderOrDefault(c, req.HeaderContentEncoding, "")
	gitProtocol := req.GetHeaderOrDefault(c, req.HeaderParamGitProtocol, "")

	// TODO: 在这里调用 service 层执行 Git 智能 HTTP 操作
	_ = contentEncoding
	_ = gitProtocol
}

func (ctl *GitController) Push(c *gin.Context) {

}

func (ctl *GitController) InfoRefs(c *gin.Context) {

}
