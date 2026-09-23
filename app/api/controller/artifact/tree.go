// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 通用树查询
// @Description 按 artifactID 查询制品树。支持两种定位方式：1）nodeID > 0 时按节点定位当前层，不支持递归；2）否则按 format + path 查询，path 可传 /** 或 /xxx/** 递归返回目录树。level 支持 format/group/package/version/asset，表示最大可见层级。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ArtifactTreeListReq true "查询条件：nodeID 与 path 二选一时 nodeID 优先；format=all 表示查询所有格式；path 带 /** 时启用递归"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/nodes [post]
func (c *Controller) ListTree(ctx *gin.Context) {
	var reqBody model.ArtifactTreeListReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	lib, err := c.artSvc.ParseSpaceLib(ctx.Request.Context(), &reqBody.SpaceRepoReq)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	list, err := c.artSvc.ListTree(ctx, lib.ArtifactID, &reqBody.ArtifactTreeFilter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, list)
}

// @Summary 重建制品树
// @Description 按 format 重建制品树节点。format 必传，支持 file/raw/maven/container/helm/pypi/npm，或 all 表示全量刷新。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.RefreshTreeReq true "刷新参数（format 必填）"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/tree/refresh [post]
func (c *Controller) RefreshTree(ctx *gin.Context) {
	var reqBody model.RefreshTreeReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		reqBody.Format = ctx.Query("format")
	}
	if err := reqBody.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	if err := c.artSvc.RefreshTree(ctx.Request.Context(), reqBody.Format); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactRefreshTreeSuccess, nil)
}
