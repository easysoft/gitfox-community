package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 按实体批量软删除制品
// @Description 按 entityIDs 批量软删除 asset/version/package/group
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SoftRemoveEntitiesReq true "实体ID列表，仅支持 entityIDs，元素格式为 type.id"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities [delete]
func (c *Controller) SoftRemoveByEntity(ctx *gin.Context) {
	var reqBody model.SoftRemoveEntitiesReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	spaceID, _, err := req.QueryParamAsPositiveInt64(ctx, "spaceID")
	if err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	data, err := c.artSvc.SoftRemove(ctx.Request.Context(), spaceID, &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactDeleteSuccess, data)
}

// @Summary 触发制品垃圾回收
// @Description 创建已软删且超过回收站保留期的制品垃圾回收后台任务；entity 为空时清理全部，支持 container/raw/file
// @Tags artifacts
// @Produce application/json
// @Param entity path string false "制品类型，留空表示全部，可选: container/raw/file"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/garbage [get]
// @Router /artifacts/garbage/{entity} [get]
func (c *Controller) GarbageCollect(ctx *gin.Context) {
	data, err := c.artSvc.TriggerGarbageCollect(ctx, model.ArtifactFormat(ctx.Param("entity")))
	if err != nil {
		if errors.GetI18nKey(err) == i18n.ArtifactViewFormatInvalid {
			resp.BadRequest(ctx, err)
			return
		}
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactGarbageCollectSuccess, data)
}
