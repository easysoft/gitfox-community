package artifact

import (
	"fmt"
	"strconv"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 查询版本文件列表
// @Description 列出指定版本的制品文件列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param versionID path int true "版本ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/versions/{versionID}/assets [get]
func (c *Controller) ListAssets(ctx *gin.Context) {
	versionID, err := req.PathParamAsPositiveInt64(ctx, "versionID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	assets, err := c.artSvc.ListAssets(ctx, versionID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, assets)
}

// @Summary 查询文件详情
// @Description 获取指定制品文件的详细信息
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param assetID path int true "文件ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetID} [get]
func (c *Controller) GetAsset(ctx *gin.Context) {
	assetID, err := req.PathParamAsPositiveInt64(ctx, "assetID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	asset, err := c.artSvc.GetAsset(ctx, assetID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactQuerySuccess, asset)
}

// @Summary 查询文件历史
// @Description 获取指定制品文件的历史 blob 列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param assetID path int true "文件ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetID}/history [get]
func (c *Controller) GetBlobHistory(ctx *gin.Context) {
	assetID, err := req.PathParamAsPositiveInt64(ctx, "assetID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	blobs, err := c.artSvc.GetBlobHistory(ctx, assetID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactQuerySuccess, blobs)
}

// @Summary 下载文件
// @Description 下载制品文件
// @Tags artifacts
// @Produce application/octet-stream
// @Param assetID path int true "文件ID"
// @Success 200 {file} file "文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetID}/download [get]
func (c *Controller) DownloadAsset(ctx *gin.Context) {
	assetID, err := req.PathParamAsPositiveInt64(ctx, "assetID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactPathParamError)
		resp.BadRequest(ctx, err)
		return
	}
	reader, blob, err := c.artSvc.DownloadAsset(ctx, assetID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", blob.Ref))
	ctx.Header("Content-Length", strconv.FormatInt(blob.Size, 10))
	ctx.DataFromReader(200, blob.Size, "application/octet-stream", reader, nil)
}
