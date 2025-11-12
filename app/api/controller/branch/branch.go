package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/principal"
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
// @Param request body branchModel.ReqForCreate true "创建分支请求参数"
// @Success 200 {object} resp.ObjectResponse{data=branchModel.Branch} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /branches [post]
func (c *Controller) Create(ctx *gin.Context) {

	var reqBody branchModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, reqBody.RepoID)
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

	// TDOO: 记录审计日志
	resp.Created(ctx, i18n.BranchCreateSuccess, respData)
}

// @Summary 查询分支列表
// @Description 分页查询分支列表
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param request body branchModel.ReqFilter true "创建分支请求参数"
// @Success 200 {object} resp.ListResponse{data=[]branchModel.Branch} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /branches/list [post]
func (c *Controller) List(ctx *gin.Context) {

	var reqQuery branchModel.ReqFilter
	var err error
	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, reqQuery.RepoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.branchSvc.List(ctx, reqQuery, repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListArgs{
			Total:    0,
			Page:     int64(1),
			PageSize: int64(1),
		}, i18n.BranchListSuccess, nil)
}

// @Summary 查询分支标签
// @Description 查询分支标签
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param repoID query uint true "仓库 ID"
// @Success 200 {object} resp.ListResponse{data=[]branchModel.Prefix} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /branches/prefixes [get]
func (c *Controller) ListPrefixes(ctx *gin.Context) {

	var reqQuery branchModel.ReqFilter
	var err error
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人
	session := &auth.Session{
		Principal: principal.Principal{
			Email: "actor@chandao.com",
			UID:   "actor",
		},
	}
	// 通过仓库 ID 获取仓库
	repoOut, err := c.repoSvc.GetByID(ctx.Request.Context(), session, reqQuery.RepoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	respData, err := c.branchSvc.ListPrefixes(ctx, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, respData, resp.ListArgs{}, i18n.BranchListPrefixSuccess, nil)
}
