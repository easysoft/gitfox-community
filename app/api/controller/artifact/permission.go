package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 设置节点级权限
// @Description 设置指定实体节点的权限（私有/公开 + 用户角色规则）
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceId query int true "空间ID"
// @Param entityId path string true "实体ID（格式: kind.id，如 package.123）"
// @Param body body model.PermissionRequest true "权限配置"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities/{entityId}/permissions [post]
func (c *Controller) SetEntityPermission(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	entityId, err := req.PathParamOrError(ctx, "entityId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var permReq model.PermissionRequest
	if err := ctx.ShouldBindJSON(&permReq); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	data, err := c.artSvc.SetEntityPermission(ctx, spaceID, entityId, &permReq)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.RepoDeleteSuccess, data)
}

// @Summary 获取节点级权限
// @Description 获取指定实体节点的权限配置（含继承信息）
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceId query int true "空间ID"
// @Param entityId path string true "实体ID（格式: kind.id，如 package.123）"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities/{entityId}/permissions [get]
func (c *Controller) GetEntityPermissions(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	entityId, err := req.PathParamOrError(ctx, "entityId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	data, err := c.artSvc.GetEntityPermissions(ctx, spaceID, entityId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.RepoDeleteSuccess, data)
}

// @Summary 设置空间级权限
// @Description 设置指定空间的制品库权限
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceId query int true "空间ID"
// @Param body body model.PermissionRequest true "权限配置"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/permissions [post]
func (c *Controller) SetSpacePermission(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var permReq model.PermissionRequest
	if err := ctx.ShouldBindJSON(&permReq); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoId"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.SetSpacePermission(ctx, spaceID, repoID, &permReq)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.RepoDeleteSuccess, data)
}

// @Summary 获取空间级权限
// @Description 获取指定空间的制品库权限配置
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceId query int true "空间ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/permissions [get]
func (c *Controller) GetSpacePermissions(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoId"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.GetSpacePermissions(ctx, spaceID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.RepoDeleteSuccess, data)
}

// @Summary 获取空间成员列表
// @Description 获取指定空间的制品库成员列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceId query int true "空间ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/members [get]
func (c *Controller) GetSpaceMembers(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoId"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.GetSpaceMembers(ctx, spaceID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.RepoDeleteSuccess, data)
}
