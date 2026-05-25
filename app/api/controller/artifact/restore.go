package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 恢复回收站中的实体
// @Description 批量恢复回收站中的 artifact / group / package / version / asset。
// @Description entityIDs 复用 "kind.id" 串（与软删一致）；恢复时向上级联祖先链路（不向下级联）。
// @Description 单条失败不阻断后续，整体返回汇总报告。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ArtifactRestoreEntitiesReq true "恢复请求体"
// @Success 200 {object} resp.ObjectResponse "成功（部分恢复时 status 字段标识失败条目）"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/recycle/restore [post]
func (c *Controller) Restore(ctx *gin.Context) {
	var reqBody model.ArtifactRestoreEntitiesReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	report, err := c.artSvc.Restore(ctx, &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}

	// 按整体成败分布选不同 i18n message：
	// - 全部失败：ArtifactRestoreAllFailed
	// - 部分失败：ArtifactRestorePartialFailed
	// - 全部成功：ArtifactRestoreSuccess
	msg := i18n.ArtifactRestoreSuccess
	switch {
	case report.Total > 0 && report.Success == 0:
		msg = i18n.ArtifactRestoreAllFailed
	case report.Failed > 0:
		msg = i18n.ArtifactRestorePartialFailed
	}
	resp.OK(ctx, msg, report)
}
