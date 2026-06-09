// Package artifact 汇聚制品库相关的 HTTP 入口。
//
// 控制器按资源域拆成多个文件，便于阅读与维护：
//
//	artifact.go    - Controller 基座、构造、Service()
//	asset.go       - GetAsset / DownloadAsset / ListAssetPage
//	group.go       - CreateGroup  / ArchiveGroup
//	tree.go        - ListTree / RefreshTree
//	upload.go      - Upload
//	admin.go       - SoftRemoveByEntity
package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	artifactModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gin-gonic/gin"
)

// Controller 处理制品库相关的 HTTP 请求。
type Controller struct {
	artSvc artifactSvc.Interface
}

// NewController 创建制品库处理器实例。
func NewController(artSvc artifactSvc.Interface) *Controller {
	return &Controller{artSvc: artSvc}
}

// Service 暴露 artifact service，供路由中间件复用统一的视图解析逻辑。
func (c *Controller) Service() artifactSvc.Interface {
	return c.artSvc
}

// @Summary 创建制品库视图
// @Description 创建一个制品库视图 (测试用)
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.CreateArtifactViewReq true "创建参数"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts [post]
func (c *Controller) CreateArtifact(ctx *gin.Context) {
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

// @Summary 移动并/或重命名制品实体
// @Description 同时支持移动和重命名 package / asset / group。targetGroupID 显式传 0 表示根目录（仅 file group 支持）；
// @Description 不传 targetGroupID 表示不移动，仍可正常执行重命名；两者都不变则直接成功。整个过程在单事务内完成。
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body artifactModel.RelocateEntityReq true "移动并/或重命名请求"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Router /artifacts/entities/relocate [post]
func (c *Controller) RelocateEntity(ctx *gin.Context) {
	var body artifactModel.RelocateEntityReq
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}

	if err := body.Validate(); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	if err := c.artSvc.RelocateEntity(ctx.Request.Context(), &body); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	resp.OK(ctx, i18n.ArtifactRelocateSuccess, nil)
}
