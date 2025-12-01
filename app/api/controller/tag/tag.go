package tag

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	tagModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/tag"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理与标签相关的 HTTP 请求
type Controller struct {
	tagSvc  tag.Interface
	repoSvc repo.Interface
}

// NewController 创建标签处理器实例
func NewController(tagService tag.Interface, repoService repo.Interface) *Controller {
	return &Controller{
		tagSvc:  tagService,
		repoSvc: repoService,
	}
}

// @Summary 创建标签
// @Description 新增一个标签
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param request body tagModel.ReqForCreate true "创建标签请求参数"
// @Success 200 {object} resp.ObjectResponse{data=tagModel.Tag} "创建成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误（如名称重复）"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/tags [post]
func (c *Controller) Create(ctx *gin.Context) {

	var reqBody tagModel.ReqForCreate
	var err error

	// 解析请求体
	if err = ctx.ShouldBindJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.TagBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.TagPathParamError)
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
		resp.InternalError(ctx, err)
		return
	}

	if repoOut.Mirror {
		err = errors.WithI18nKey(err, i18n.RepoMirrorReadOnly)
		resp.InternalError(ctx, err)
		return
	}
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层创建标签
	respData, err := c.tagSvc.Create(ctx.Request.Context(), reqBody, session, repoOut)
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
// @Param id path uint true "仓库 ID"
// @Param request body tagModel.ReqFilter true "查询标签列表参数"
// @Success 200 {object} resp.ListResponse{data=[]tagModel.Tag,listArgs=resp.ListNoTotalArgs} "查询成功"
// @Failure 400 {object} resp.ListResponse "参数错误"
// @Failure 500 {object} resp.ListResponse "服务器错误"
// @Router /repos/{id}/tags/list [post]
func (c *Controller) List(ctx *gin.Context) {
	var reqQuery tagModel.ReqFilter
	var err error

	if err = ctx.ShouldBindJSON(&reqQuery); err != nil {
		err = errors.WithI18nKey(err, i18n.TagQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	// 设置默认值
	req.SetDefaults(&reqQuery)

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.TagPathParamError)
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
		resp.InternalError(ctx, err)
		return
	}

	respData, isLast, err := c.tagSvc.List(ctx.Request.Context(), reqQuery, *repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.ListSuccess(ctx, respData,
		resp.ListNoTotalArgs{
			IsLast:   isLast,
			Page:     int64(reqQuery.Page),
			PageSize: int64(reqQuery.PageSize),
		}, i18n.TagListSuccess, nil)
}

// @Summary 删除标签
// @Description 删除指定标签
// @Tags tags
// @Accept application/json
// @Produce application/json
// @Param id path uint true "仓库 ID"
// @Param name path string true "标签名称"
// @Success 200 {object} resp.ObjectResponse "删除成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /repos/{id}/tags/{name} [delete]
func (c *Controller) Delete(ctx *gin.Context) {

	var err error
	name := ctx.Param("name")
	if name == "" {
		err = errors.WithI18nKey(nil, i18n.TagPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 解析 path id
	repoID, err := req.PathParamAsPositiveUint(ctx, "id")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.TagPathParamError)
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
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 检查是否为镜像仓库
	if repoOut.Mirror {
		err = errors.WithI18nKey(err, i18n.RepoMirrorReadOnly)
		resp.InternalError(ctx, err)
		return
	}
	// TODO: 获取保护规则，如果有违规操作直接返回错误

	// 调用 service 层删除标签
	err = c.tagSvc.Delete(ctx.Request.Context(), name, session, repoOut)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// TODO: 记录审计日志
	resp.OK(ctx, i18n.TagDeleteSuccess, nil)
}
