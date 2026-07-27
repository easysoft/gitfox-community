package artifact

import (
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
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

	data, err := c.artSvc.SoftRemove(ctx.Request.Context(), &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactDeleteSuccess, data)
}

func (c *Controller) GarbageCollect(ctx *gin.Context) {
	req := ctx.Query("format")

	raw := strings.TrimSpace(req)
	if raw != "" {
		_, ok := constant.ArtifactFormat(raw).Sanitize()
		if !ok {
			resp.BadRequest(ctx, errors.WithI18nKey(fmt.Errorf("unsupported garbage collect format: %s", raw), i18n.ArtifactLibFormatInvalid))
			return
		}
	}

	data, err := c.artSvc.TriggerGarbageCollect(ctx, req)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ArtifactGarbageCollectSuccess, data)
}
