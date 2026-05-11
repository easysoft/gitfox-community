package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 通用树查询
// @Description 按 artifactID 查询制品树。支持两种定位方式：1）nodeId > 0 时按节点定位当前层，不支持递归；2）否则按 format + path 查询，path 可传 /** 或 /xxx/** 递归返回目录树。level 支持 format/group/package/version/asset，表示最大可见层级。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ArtifactTreeListReq true "查询条件：nodeId 与 path 二选一时 nodeId 优先；format=all 表示查询所有格式；path 带 /** 时启用递归"
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
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody.SpaceRepoReq)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	list, err := c.artSvc.ListTree(ctx, view.ArtifactID, &reqBody.ArtifactTreeFilter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, list)
}

// @Summary 查询树节点列表
// @Description 列出树节点的子节点
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param nodeId path int true "节点ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/nodes/{nodeId}/children [get]
func (c *Controller) ListTreeNodes(ctx *gin.Context) {
	parentId, err := req.PathParamAsPositiveInt64(ctx, "nodeId")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	nodes, err := c.artSvc.ListTreeNodes(ctx, parentId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, nodes)
}

// @Summary 重建制品树
// @Description 重建所有制品树节点
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/tree/refresh [post]
func (c *Controller) RefreshTree(ctx *gin.Context) {
	if err := c.artSvc.RefreshTree(ctx); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactRefreshTreeSuccess, nil)
}
