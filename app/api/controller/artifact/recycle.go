package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 回收站列表
// @Description 分页查询回收站中的实体记录。kind 必填，默认 asset；支持 artifact/group/package/version/asset 五种维度。
// @Description 仅查 deleted=1 的记录，不再做时间窗过滤；前端按 expiresAt 自行展示距离 GC 多久。
// @Tags artifacts
// @Produce application/json
// @Param kind query string false "实体类型，可选：artifact/group/package/version/asset，默认 asset"
// @Param page query int false "页码，默认 1"
// @Param pageSize query int false "每页数量，默认 20，最大 200"
// @Success 200 {object} resp.ListResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/recycle/list [get]
func (c *Controller) ListRecycleBin(ctx *gin.Context) {
	var reqBody model.ArtifactRecycleListReq
	if err := ctx.ShouldBindQuery(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactQueryParamError)
		resp.BadRequest(ctx, err)
		return
	}
	// kind 缺省走 asset；非法值由 service 层校验。
	if reqBody.Kind == "" {
		reqBody.Kind = model.ArtifactRecycleKindAsset
	}

	data, total, err := c.artSvc.ListRecycleBin(ctx, &reqBody)
	if err != nil {
		// kind 不合法时是参数错误，其它当作内部错误返回。
		if errors.GetI18nKey(err) == i18n.ArtifactRecycleKindInvalid {
			resp.BadRequest(ctx, err)
			return
		}
		resp.InternalError(ctx, err)
		return
	}

	page, pageSize := reqBody.Page, reqBody.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	resp.ListSuccess(ctx, data, resp.ListArgs{
		Total:    total,
		Page:     int64(page),
		PageSize: int64(pageSize),
	}, i18n.ArtifactRecycleListSuccess, nil)
}
