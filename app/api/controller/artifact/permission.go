package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 设置节点级权限
// @Description 设置指定实体节点的权限（私有/公开 + 用户角色规则）
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceID query int true "空间ID"
// @Param entityID path string true "实体ID（格式: kind.id，如 package.123）"
// @Param body body model.PermissionRequest true "权限配置"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities/{entityID}/permissions [post]
func (c *Controller) SetEntityPermission(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	entityID, err := req.PathParamOrError(ctx, "entityID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var permReq model.PermissionRequest
	if err := ctx.ShouldBindJSON(&permReq); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	data, err := c.artSvc.SetEntityPermission(ctx, spaceID, entityID, &permReq)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactPermissionUpdateSuccess, data)
}

// @Summary 获取节点级权限
// @Description 获取指定实体节点的权限配置（含继承信息）
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceID query int true "空间ID"
// @Param entityID path string true "实体ID（格式: kind.id，如 package.123）"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities/{entityID}/permissions [get]
func (c *Controller) GetEntityPermissions(ctx *gin.Context) {
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}

	entityID, err := req.PathParamOrError(ctx, "entityID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}

	data, err := c.artSvc.GetEntityPermissions(ctx, spaceID, entityID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactPermissionQuerySuccess, data)
}

// @Summary 设置空间级权限
// @Description 设置指定空间/仓库的制品库权限。spaceID 和 repoID 均为可选：
// 不传 spaceID（或传 0）表示设置全局权限（适用于所有 space）；
// 只传 spaceID 表示设置指定 space 权限；同时传 spaceID + repoID 表示设置指定仓库权限。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceID query int false "空间ID（可选，不传或传 0 表示全局权限）"
// @Param repoID query int false "仓库ID（可选）"
// @Param body body model.PermissionRequest true "权限配置"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/permissions [post]
func (c *Controller) SetPermissions(ctx *gin.Context) {
	// spaceID 可选：不传或 0 → 全局权限（space:0）
	var spaceID int64
	if sID, _, _ := req.QueryParamAsPositiveInt64(ctx, "spaceID"); sID > 0 {
		spaceID = sID
	}

	var permReq model.PermissionRequest
	if err := ctx.ShouldBindJSON(&permReq); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoID"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.SetSpacePermission(ctx, spaceID, repoID, &permReq)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactPermissionUpdateSuccess, data)
}

// @Summary 获取空间级权限
// @Description 获取指定空间/仓库的制品库权限配置。
// spaceID 可选：不传或传 0 表示查询全局权限。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceID query int false "空间ID（可选，不传或传 0 表示全局权限）"
// @Param repoID query int false "仓库ID（可选）"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/permissions [get]
func (c *Controller) GetPermissions(ctx *gin.Context) {
	// spaceID 可选：不传或 0 → 全局权限（space:0）
	var spaceID int64
	if sID, _, _ := req.QueryParamAsPositiveInt64(ctx, "spaceID"); sID > 0 {
		spaceID = sID
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoID"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.GetSpacePermissions(ctx, spaceID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactPermissionQuerySuccess, data)
}

// @Summary 获取空间成员列表
// @Description 获取指定空间/仓库的制品库成员列表。
// spaceID 可选：不传或传 0 表示查询全局成员。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param spaceID query int false "空间ID（可选，不传或传 0 表示全局）"
// @Param repoID query int false "仓库ID（可选）"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/members [get]
func (c *Controller) GetMembers(ctx *gin.Context) {
	// spaceID 可选：不传或 0 → 全局（space:0）
	var spaceID int64
	if sID, _, _ := req.QueryParamAsPositiveInt64(ctx, "spaceID"); sID > 0 {
		spaceID = sID
	}

	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoID"); rID > 0 {
		repoID = rID
	}

	data, err := c.artSvc.GetSpaceMembers(ctx, spaceID, repoID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactPermissionQuerySuccess, data)
}
