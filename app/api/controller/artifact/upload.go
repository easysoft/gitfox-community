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

	view, err := c.artSvc.ParseSpaceView(ctx, &model.SpaceRepoReq{
		ArtifactID: reqBody.ArtifactID,
	})
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	formatStr := ctx.Param("format")
	format := model.ArtifactFormat(formatStr)

	assets, err := c.artSvc.Upload(ctx, view, format, ctx.Request)
	if err != nil {
		// 若 err 在 adapter / service 层已显式携带 HTTP 状态（如 400），按其渲染；
		// 否则按 500 渲染。i18n key 通过 GetI18nKey 透传到前端。
		if errors.GetHttpStatus(err) == http.StatusOK {
			resp.InternalError(ctx, err)
		} else {
			resp.Error(ctx, err)
		}
		return
	}
	resp.OK(ctx, i18n.ArtifactUploadSuccess, assets)
}
