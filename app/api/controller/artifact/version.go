package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"

	"github.com/gin-gonic/gin"
)

// @Summary 创建制品库视图
// @Description 创建一个制品库视图 (测试用)
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.CreateArtifactViewReq true "创建参数"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/views [post]
func (c *Controller) CreateView(ctx *gin.Context) {
	var reqBody model.CreateArtifactViewReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	// 业务规则校验（format 枚举）
	if err := reqBody.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	view, err := c.artSvc.CreateView(ctx.Request.Context(), &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactCreateViewSuccess, view)
}

// @Summary 查询版本列表
// @Description 列出指定制品包的版本列表。artifactID > 0 时优先按指定视图查询，否则按默认 view 逻辑查询
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.ArtifactVersionListReq true "查询条件"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/versions/list [post]
func (c *Controller) ListVersions(ctx *gin.Context) {
	var reqBody model.ArtifactVersionListReq
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

	versions, err := c.artSvc.ListVersions(ctx, view.ArtifactID,
		reqBody.PackageName, reqBody.GroupName, model.ArtifactFormat(reqBody.Format), model.SearchVersionOption{})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, versions)
}
