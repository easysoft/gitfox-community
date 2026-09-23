// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package svn

import (
	"fmt"
	"net/http"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	svnModel "github.com/easysoft/gitfox-community/app/model/svn"
	"github.com/easysoft/gitfox-community/app/service/repo"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/easysoft/gitfox-community/pkg/logger"
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
	repo, err := c.repoSvc.GetByID(ctx, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SvnRepoNotFound)
		resp.NotFound(ctx, err)
		return
	}

	// 业务全部下沉到 service：白名单校验、SCMType 判断、Connector 解析、命令分发、Content-Type 选择
	output, contentType, err := c.repoSvc.ExecSvnCommand(ctx.Request.Context(), *repo, reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 流式原文返回：保留 svn 原生 stdout,调用方自行解析 XML / diff / 纯文本
	ctx.Header("Content-Type", contentType)
	ctx.Header("X-Content-Type-Options", "nosniff") // 防止浏览器嗅探 Content-Type
	ctx.String(http.StatusOK, output)
}

// Export 导出 SVN 仓库并以 ZIP 文件流返回。
// @Summary 导出 SVN 仓库
// @Description 导出指定 revision 的 SVN 仓库内容并直接返回 ZIP 文件流，不生成下载 URL
// @Tags svn
// @Produce application/zip
// @Param id path uint true "仓库 ID"
// @Param revision query string false "SVN revision，留空时使用 HEAD"
// @Success 200 {file} binary "SVN 仓库 ZIP 归档"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权或无会话"
// @Failure 404 {object} resp.ObjectResponse "代码库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/svn/export [get]
func (c *Controller) Export(ctx *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SvnPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("get session failed"), i18n.AuthNoSession)
		resp.Unauthorized(ctx, err)
		return
	}

	repo, err := c.repoSvc.GetByID(ctx, session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.SvnRepoNotFound)
		resp.NotFound(ctx, err)
		return
	}

	// 预设 ZIP 下载响应头。若 ExportSvn 在写响应体之前失败（streamStarted==false）,
	// 下方分支会清掉这些头，避免 JSON 错误响应被浏览器保存成 .zip 文件。
	filename := repo.Identifier + ".zip"
	ctx.Header("Content-Type", "application/zip")                                       // 设置 Content-Type 为 ZIP 归档
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename)) // 下载文件名
	ctx.Header("X-Content-Type-Options", "nosniff")                                     // 防止浏览器嗅探 Content-Type

	streamStarted, err := c.repoSvc.ExportSvn(ctx.Request.Context(), *repo, ctx.Query("revision"), ctx.Writer)
	if err != nil {
		logger.GetLogger(ctx).Errorf("failed to export SVN repository %d: %v", repoID, err)
		if !streamStarted {
			// 未向响应体写入任何字节，清掉预设的 ZIP 头，改走 500 JSON 让前端识别错误。
			h := ctx.Writer.Header()
			h.Del("Content-Type")
			h.Del("Content-Disposition")
			h.Del("X-Content-Type-Options")
			if errors.GetI18nKey(err) == i18n.ErrorDefault {
				err = errors.WithI18nKey(err, i18n.SvnExportFailed)
			}
			resp.InternalError(ctx, err)
			return
		}
		return
	}
}
