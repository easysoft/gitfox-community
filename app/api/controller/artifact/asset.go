package artifact

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/easysoft/gitfox-community/app/api/utils/req"
	"github.com/easysoft/gitfox-community/app/api/utils/resp"
	artifactModel "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

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
	asset, err := c.artSvc.GetAsset(ctx.Request.Context(), assetID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactQuerySuccess, asset)
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
	reader, blob, fielname, err := c.artSvc.DownloadAsset(ctx.Request.Context(), assetID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fielname))
	ctx.Header("Content-Length", strconv.FormatInt(blob.Size, 10))
	ctx.DataFromReader(200, blob.Size, "application/octet-stream", reader, nil)
}

// @Summary 按路径下载文件
// @Description 通过制品完整路径直接下载文件，无需鉴权。
// @Tags artifacts
// @Produce application/octet-stream
// @Param filepath path string true "制品路径，格式: viewName/groupPath/packageName/version/fileName"
// @Success 200 {file} file "文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /_artifacts/space/{code}/{filepath} [get]
func (c *Controller) DownloadAssetByPath(ctx *gin.Context) {
	spaceCode,err := req.PathParam(ctx, "code")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("spaceCode is required"),
			i18n.ArtifactPathParamError,
		))
		return
	}
	filepath,err := req.PathParam(ctx,"filepath")
	if err != nil {
		resp.BadRequest(ctx, errors.WithI18nKey(
			fmt.Errorf("filepath is required"),
			i18n.ArtifactPathParamError,
		))
		return
	}
	filepath = strings.TrimPrefix(filepath, "/")
	reader, blob, filename, err := c.artSvc.DownloadAssetByPath(ctx.Request.Context(), spaceCode, filepath)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	ctx.Header("Content-Length", strconv.FormatInt(blob.Size, 10))
	ctx.DataFromReader(200, blob.Size, "application/octet-stream", reader, nil)
}

// @Summary 制品资产分页列表
// @Description 按 entityID（group.{id} 或 version.{id}）分页查询资产列表，支持过滤和排序；group 查询传 more=true 时递归包含子目录资产
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body artifactModel.ListAssetReq true "分页查询请求"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/list [post]
func (c *Controller) ListAssetPage(ctx *gin.Context) {
	var body artifactModel.ListAssetReq
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	pageResult, err := c.artSvc.ListAssetPage(ctx.Request.Context(), &body)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.ListSuccess(ctx, pageResult.Data, resp.ListArgs{
		Total:    pageResult.Total,
		Page:     int64(pageResult.Page),
		PageSize: int64(pageResult.PageSize),
	}, i18n.ArtifactListSuccess, nil)
}
