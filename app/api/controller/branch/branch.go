package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Controller 处理与分支相关的 HTTP 请求
type Controller struct {
	branchSvc branch.Interface
}

// NewController 创建分支处理器实例
func NewController(branchService branch.Interface) *Controller {
	return &Controller{
		branchSvc: branchService,
	}
}

// @Summary 创建分支
// @Description 新增一个分支
// @Tags branches
// @Accept application/json
// @Produce application/json
// @Param request body branch.ReqForCreate true "创建分支请求参数"
// @Success 200 {object} resp.ObjectResponse{data=branch.Branch} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /branches [post]
func (c *Controller) Create(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqBody branchModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人

	// 通过仓库 ID 获取仓库
	log.Debugln(reqBody.RepoID)

	// TODO: 检查是否为镜像仓库
	// TODO: 校验用户是否有仓库的 Push 权限
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// TODO: 如果未指定目标分支，使用仓库默认分支作为新分支的基准分支
	// 调用 service 层创建分支
	respData, err := c.branchSvc.Create(ctx.Request.Context(), reqBody)
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
// @Param repoID query uint true "仓库 ID"
// @Param query query string false "分支名称（模糊查询）"
// @Param prefix query string false "分支前缀"
// @Param includeArchived query bool false "是否包含归档分支"
// @Param includeCommit query bool false "是否包含 commit 信息"
// @Param page query int false "页码（默认1）" default(1) minimum(1)
// @Param pageSize query int false "每页条数（默认20）" default(20) minimum(1) maximum(100)
// @Success 200 {object} resp.ListResponse{data=[]branch.Branch} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /branches/list [get]
func (c *Controller) List(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqQuery branchModel.ReqFilter
	var err error
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	log.Debugln(reqQuery.RepoID)
	// repo, err := ctl.repoSvc.Get(ctx, session, repoRef, enum.PermissionRepoView)
	// if err != nil {
	// 	return nil, err
	// }

	respData, err := c.branchSvc.List(ctx, reqQuery, repoModel.Repo{})
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
// @Success 200 {object} resp.ListResponse{data=[]branch.Prefix} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /branches/prefixes [get]
func (c *Controller) ListPrefixes(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqQuery branchModel.ReqFilter
	var err error
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.BranchQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	log.Debugln(reqQuery.RepoID)
	// repo, err := ctl.repoSvc.Get(ctx, session, repoRef, enum.PermissionRepoView)
	// if err != nil {
	// 	return nil, err
	// }

	respData, err := c.branchSvc.ListPrefixes(ctx, repoModel.Repo{})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.ListSuccess(ctx, respData, resp.ListArgs{}, i18n.BranchListPrefixSuccess, nil)
}
