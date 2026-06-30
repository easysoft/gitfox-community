package svn

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	svnModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/svn"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与 SVN 相关的 HTTP 请求
type Controller struct {
	repoSvc repo.Interface
}

// NewController 创建 SVN 处理器实例
func NewController(repoService repo.Interface) *Controller {
	return &Controller{
		repoSvc: repoService,
	}
}

// @Summary 执行 SVN 命令
// @Description 统一调度 SVN 子命令（ls/log/blame/cat/info/diff）,原样返回 svn stdout
// @Tags svn
// @Accept application/json
// @Produce text/plain
// @Param id path uint true "仓库 ID"
// @Param request body svnModel.ReqForExec true "SVN 执行请求"
// @Success 200 {string} string "svn 原生 stdout"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 404 {object} resp.ObjectResponse "代码库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/svn/exec [post]
func (c *Controller) Exec(ctx *gin.Context) {
	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SvnPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	var reqBody svnModel.ReqForExec
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.SvnBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 鉴权会话校验（与 GetDiff 一致：取不到走 Unauthorized）
	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库详情
	repoOut, err := c.repoSvc.GetByID(ctx, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SvnRepoNotFound)
		resp.InternalError(ctx, err)
		return
	}

	// 业务全部下沉到 service：白名单校验、SCMType 判断、Connector 解析、命令分发、Content-Type 选择
	output, contentType, err := c.repoSvc.ExecSvn(ctx.Request.Context(), *repoOut, reqBody)
	if err != nil {
		// service 内已包过 i18n,按错误关联的状态码自适应
		resp.Error(ctx, err)
		return
	}

	// 流式原文返回：保留 svn 原生 stdout,调用方自行解析 XML / diff / 纯文本
	ctx.Header("Content-Type", contentType)
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.String(http.StatusOK, output)
}
