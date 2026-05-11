package artifact

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 查询制品包列表
// @Description 列出指定作用域下的制品包。artifactID > 0 时优先按指定视图查询，否则按默认 view 逻辑查询
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SpaceRepoReq true "查询范围"
// @Param format query string false "制品格式"
// @Param query query string false "查询关键字"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/packages [post]
func (c *Controller) ListPackages(ctx *gin.Context) {
	var reqBody model.SpaceRepoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	filter := &model.ArtifactFilter{
		Format: req.QueryParamOrDefault(ctx, "format", ""),
		Query:  req.QueryParamOrDefault(ctx, "query", ""),
	}
	list, err := c.artSvc.ListPackages(ctx, view.ArtifactID, filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, list)
}

// @Summary 查询格式列表
// @Description 列出指定作用域下的已用格式。artifactID > 0 时优先按指定视图查询，否则按默认 view 逻辑查询
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SpaceRepoReq true "查询范围"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/formats [post]
func (c *Controller) ListFormats(ctx *gin.Context) {
	var reqBody model.SpaceRepoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	formats, err := c.artSvc.ListFormats(ctx, view.ArtifactID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, formats)
}

// @Summary 查询节点详情
// @Description 兼容参考项目 `/artifacts/nodeInfo`，按 entity_ids 或旧的 node_ids 批量查询节点详情。成功时返回原始数组，不包裹通用响应体。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ListNodeInfoReq true "节点 ID 列表"
// @Success 200 {array} model.ArtifactNodeInfo "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/nodeInfo [post]
func (c *Controller) ListNodeInfo(ctx *gin.Context) {
	var reqBody model.ListNodeInfoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	list, err := c.artSvc.ListNodeInfo(ctx, &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, list)
}
