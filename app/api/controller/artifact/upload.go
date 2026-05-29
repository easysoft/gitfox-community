package artifact

import (
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 上传制品
// @Description 上传制品文件（multipart/form-data）。artifactID 必传，用于精确定位制品库视图。
// @Tags artifacts
// @Accept multipart/form-data
// @Produce application/json
// @Param format path string true "制品格式"
// @Param artifactID formData int true "视图ID（必传，> 0）"
// @Param name formData string true "文件名（覆盖 asset.path）"
// @Param packageName formData string false "raw 包名（raw 上传必传）"
// @Param file formData file true "制品文件"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/upload/{format} [post]
func (c *Controller) Upload(ctx *gin.Context) {
	var reqBody model.SpaceRepoFormReq
	if err := ctx.ShouldBind(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	if err := reqBody.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	view, err := c.artSvc.ParseSpaceView(ctx.Request.Context(), &model.SpaceRepoReq{
		ArtifactID: reqBody.ArtifactID,
	})
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	formatStr := ctx.Param("format")
	format := model.ArtifactFormat(formatStr)

	assets, err := c.artSvc.Upload(ctx.Request.Context(), view, format, ctx.Request)
	if err != nil {
		if errors.GetHttpStatus(err) == http.StatusOK {
			resp.InternalError(ctx, err)
		} else {
			resp.Error(ctx, err)
		}
		return
	}
	// 同名同内容命中：assets[0].AlreadyExists=true，切换为"文件已存在"成功提示，
	// 仍按 200 渲染，前端拿到现有 asset 即可。
	if len(assets) > 0 && assets[0].AlreadyExists {
		resp.OK(ctx, i18n.ArtifactUploadFileDuplicate, assets)
		return
	}
	resp.OK(ctx, i18n.ArtifactUploadSuccess, assets)
}
