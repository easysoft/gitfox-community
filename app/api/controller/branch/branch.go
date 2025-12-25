package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与分支相关的 HTTP 请求
type Controller struct {
	branchSvc branch.Interface
	repoSvc   repo.Interface
}

// NewController 创建分支处理器实例
func NewController(branchService branch.Interface, repoService repo.Interface) *Controller {
	return &Controller{
		branchSvc: branchService,
		repoSvc:   repoService,
	}
}

// @Summary 创建分支
// @Description 新增一个分支
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body branchModel.ReqForCreate true "创建分支请求参数"
// @Success 200 {object} resp.ObjectResponse{data=branchModel.Branch} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branches [post]
func (c *Controller) Create(ctx *gin.Context) {
	var reqBody branchModel.ReqForCreate
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
		err = errors.WithI18nKey(err, i18n.BranchBodyParamError)
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
	if repoOut.Mirror {
		err = errors.WithI18nKey(err, i18n.RepoMirrorReadOnly)
		resp.InternalError(ctx, err)
		return
	}
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层创建分支
	if reqBody.Source == "" {
		reqBody.Source = repoOut.DefaultBranch
	}
	respData, err := c.branchSvc.Create(ctx.Request.Context(), reqBody, session, repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 记录审计日志
	resp.Created(ctx, i18n.BranchCreateSuccess, respData)
}

// @Summary 查询分支列表
// @Description 分页查询分支列表
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body branchModel.ReqFilter true "查询分支请求参数"
// @Success 200 {object} resp.ListResponse{data=[]branchModel.Branch,listArgs=resp.ListNoTotalArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/branches/list [post]
func (c *Controller) List(ctx *gin.Context) {
	var reqQuery branchModel.ReqFilter
	var err error

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchQueryParamError)
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

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, isLast, err := c.branchSvc.List(ctx.Request.Context(), reqQuery, repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListNoTotalArgs{
			IsLast:   isLast,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.PageSize),
		}, i18n.BranchListSuccess, nil)
}

// @Summary 查询分支标签
// @Description 查询分支标签
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Success 200 {object} resp.ListResponse{data=[]branchModel.Prefix} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/branches/prefixes [get]
func (c *Controller) ListPrefixes(ctx *gin.Context) {
	var err error

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

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.branchSvc.ListPrefixes(ctx.Request.Context(), *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, respData, nil, i18n.BranchListPrefixSuccess, nil)
}

// @Summary 删除分支
// @Description 删除指定分支
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body branchModel.ReqForDelete true "删除分支请求参数"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/branches [delete]
func (c *Controller) Delete(ctx *gin.Context) {

	var err error
	var reqBody branchModel.ReqForDelete

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	session, ok := req.GetAuthSession(ctx.Request.Context())
	if !ok {
		err = errors.WithI18nKey(errors.New("session not found"), i18n.AuthNoSession)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, repoID)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.RepoQueryError)
		resp.BadRequest(ctx, err)
		return
	}

	if repoOut.Mirror {
		err = errors.WithI18nKey(err, i18n.RepoMirrorReadOnly)
		resp.InternalError(ctx, err)
		return
	}

	// 调用 service 层删除分支
	err = c.branchSvc.Delete(ctx.Request.Context(), reqBody.Name, session, repoOut)
	if err != nil {
		err = errors.WithI18nKey(err, i18n.BranchDeleteError)
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.BranchDeleteSuccess, nil)
}
