package tag

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	tagModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Controller 处理与标签相关的 HTTP 请求
type Controller struct {
	tagSvc tag.Interface
}

// NewController 创建标签处理器实例
func NewController(tagService tag.Interface) *Controller {
	return &Controller{
		tagSvc: tagService,
	}
}

// @Summary 创建标签
// @Description 新增一个标签
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param request body tag.ReqForCreate true "创建标签请求参数"
// @Success 200 {object} resp.ObjectResponse{data=tag.Tag} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /tags [post]
func (c *Controller) Create(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqBody tagModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.TagBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人

	// 通过仓库 ID 获取仓库
	log.Debugln(reqBody.RepoID)

	// TODO: 检查是否为镜像仓库
	// TODO: 校验用户是否有仓库的 Push 权限
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层创建标签
	respData, err := c.tagSvc.Create(ctx.Request.Context(), reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 记录审计日志
	resp.Created(ctx, i18n.TagCreateSuccess, respData)
}

// @Summary 查询标签列表
// @Description 分页查询标签列表
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param repoID query uint true "仓库 ID"
// @Param query query string false "标签名称（模糊查询）"
// @Param sort query string false "排序方式（date/name）"
// @Param order query string false "排序顺序（asc/desc）"
// @Param includeCommit query bool false "是否包含 commit 信息"
// @Param page query int false "页码（默认1）" default(1) minimum(1)
// @Param size query int false "每页条数（默认20）" default(20) minimum(1) maximum(100)
// @Success 200 {object} resp.ListResponse{data=[]tag.Tag} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /tags/list [get]
func (c *Controller) List(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	var reqQuery tagModel.ReqFilter
	var err error
	if err = ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.TagQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 通过仓库 ID 获取仓库
	log.Debugln(reqQuery.RepoID)
	// repo, err := ctl.repoSvc.Get(ctx, session, repoRef, enum.PermissionRepoView)
	// if err != nil {
	// 	return nil, err
	// }

	respData, err := c.tagSvc.List(ctx.Request.Context(), reqQuery, repoModel.Repo{})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListArgs{
			Total:    0,
			Page:     int64(1),
			PageSize: int64(1),
		}, i18n.TagListSuccess, nil)
}

// @Summary 删除标签
// @Description 删除指定标签
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param name path string true "标签名称"
// @Param repoID query uint true "仓库 ID"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /tags/{name} [delete]
func (c *Controller) Delete(ctx *gin.Context) {

	log := logger.GetLogger(ctx.Request.Context())

	name := ctx.Param("name")
	if name == "" {
		err := errors.WithI18nKey(nil, i18n.TagPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var reqQuery struct {
		RepoID uint `form:"repoID" binding:"required"`
	}
	if err := ctx.ShouldBindQuery(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.TagQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// TODO: 获取操作人

	// 通过仓库 ID 获取仓库
	log.Debugln(reqQuery.RepoID)

	// TODO: 检查是否为镜像仓库
	// TODO: 校验用户是否有仓库的 Push 权限
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层删除标签
	err := c.tagSvc.Delete(ctx.Request.Context(), name, repoModel.Repo{})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 记录审计日志
	resp.OK(ctx, i18n.TagDeleteSuccess, nil)
}
