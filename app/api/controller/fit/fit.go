// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package fit

import (
	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	"github.com/easysoft/gitfox-community/app/constant"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	fitSvc "github.com/easysoft/gitfox-community/app/service/fit"
	repoSvc "github.com/easysoft/gitfox-community/app/service/repo"
	spaceSvc "github.com/easysoft/gitfox-community/app/service/space"
	svcUtils "github.com/easysoft/gitfox-community/app/service/utils"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// ensure imports are used for swagger generation
var (
	_ *svcUtils.BranchPermissionResult
	_ *repoModel.Repo
)

// Controller 处理 fit 相关的 HTTP 请求（Basic 认证，直连外部系统）
type Controller struct {
	fitSvc   fitSvc.Interface
	repoSvc  repoSvc.Interface
	spaceSvc spaceSvc.Interface
}

// NewController 创建 fit 处理器实例
func NewController(fitSvc fitSvc.Interface, repoSvc repoSvc.Interface, spaceSvc spaceSvc.Interface) *Controller {
	return &Controller{
		fitSvc:   fitSvc,
		repoSvc:  repoSvc,
		spaceSvc: spaceSvc,
	}
}

// GetBranchPermission 获取分支权限（通过外部系统，不查库）
// @Summary 获取分支权限
// @Description 通过外部 DevOps 系统获取分支权限信息，不经过本地数据库
// @Tags fit
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param branchName query string true "分支名称"
// @Success 200 {object} resp.ObjectResponse{data=svcUtils.BranchPermissionResult} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /fit/repos/{id}/settings/advanced [get]
func (ctl *Controller) GetBranchPermission(c *gin.Context) {
	repoID, err := req.PathParamAsPositiveUint(c, "id")
	if err != nil {
		resp.BadRequest(c, errors.WithI18nKey(err, i18n.RepoPathParamError))
		return
	}

	branchName := c.Query(req.QueryParamBranchName)
	if branchName == "" {
		resp.BadRequest(c, errors.WithI18nKey(errors.New("branch name is required"), i18n.FitBranchNameRequired))
		return
	}

	result, err := ctl.fitSvc.GetBranchPermission(c.Request.Context(), repoID, branchName)
	if err != nil {
		resp.InternalError(c, errors.WithI18nKey(err, i18n.FitBranchPermissionQueryError))
		return
	}

	resp.OK(c, i18n.FitBranchPermissionQuerySuccess, result)
}

// GetRepoByName 根据空间 ID 和仓库名获取仓库信息
// @Summary 根据空间和仓库名获取仓库
// @Description 根据空间 ID 和仓库标识名获取仓库信息，带权限校验
// @Tags fit
// @Accept application/json
// @Produce application/json
// @Param code path string true "空间编码"
// @Param repoName query string true "仓库名称"
// @Success 200 {object} resp.ObjectResponse{data=repoModel.Repo} "查询成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 401 {object} resp.ObjectResponse "未授权"
// @Failure 404 {object} resp.ObjectResponse "空间或仓库不存在"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /fit/spaces/{code} [get]
func (ctl *Controller) GetRepoByName(c *gin.Context) {
	session, ok := req.GetAuthSession(c.Request.Context())
	if !ok {
		resp.Unauthorized(c, errors.WithI18nKey(nil, i18n.AuthNoSession))
		return
	}

	spaceCode, err := req.PathParamOrError(c, "code")
	if err != nil {
		resp.BadRequest(c, errors.WithI18nKey(err, i18n.SpacePathParamError))
		return
	}

	space, err := ctl.spaceSvc.GetByCode(c.Request.Context(), spaceCode)
	if err != nil {
		resp.NotFound(c, errors.WithI18nKey(err, i18n.SpaceNotFound))
		return
	}

	repoName := c.Query(req.QueryParamRepoName)
	if repoName == "" {
		resp.BadRequest(c, errors.WithI18nKey(errors.New("repo name is required"), i18n.FitRepoNameRequired))
		return
	}

	repo, err := ctl.repoSvc.GetRepoCheckAccess(c.Request.Context(), session, space.ID, repoName, []constant.Permission{constant.PermissionRepoView})
	if err != nil {
		if errors.IsNotFound(err) {
			resp.NotFound(c, err)
			return
		}
		resp.InternalError(c, errors.WithI18nKey(err, i18n.FitRepoQueryError))
		return
	}

	resp.OK(c, i18n.FitRepoQuerySuccess, repo)
}
