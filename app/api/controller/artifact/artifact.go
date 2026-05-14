// Package artifact 汇聚制品库相关的 HTTP 入口。
//
// 控制器按资源域拆成多个文件，便于阅读与维护：
//
//	artifact.go    - Controller 基座、构造、Service()
//	package.go     - ListPackages / ListFormats / ListNodeInfo
//	version.go     - ListVersions / CreateView
//	asset.go       - ListAssets / GetAsset / GetBlobHistory / DownloadAsset
//	group.go       - CreateGroup / ListGroupAssets / ArchiveGroup
//	tree.go        - ListTree / ListTreeNodes / RefreshTree
//	upload.go      - Upload
//	statistic.go   - StatisticCapacity / StatisticContainerCapacity / StatisticSummary
//	admin.go       - SoftRemoveByEntity / GarbageCollectContainer / GarbageCollectSoftRemove
//	permission.go  - 权限相关 handler（原文件保留）
package artifact

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	artifactModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
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

// @Summary 移动制品实体
// @Description 将 package / asset / group 移动到目标制品库或目标分组
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body artifactModel.MoveEntityReq true "移动请求"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Router /artifacts/entities/move [post]
func (c *Controller) MoveEntity(ctx *gin.Context) {
	var body artifactModel.MoveEntityReq
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	if err := c.artSvc.MoveEntity(ctx, &body); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactMoveSuccess, nil)
}

// @Summary 重命名制品实体
// @Description 重命名 asset 或 group（package 不支持重命名）
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body artifactModel.RenameEntityReq true "重命名请求"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误或业务校验失败"
// @Router /artifacts/entities/rename [post]
func (c *Controller) RenameEntity(ctx *gin.Context) {
	var body artifactModel.RenameEntityReq
	if err := ctx.ShouldBindBodyWithJSON(&body); err != nil {
		err = errors.WithI18nKey(err, i18n.ArtifactBodyParamError)
		resp.BadRequest(ctx, err)
		return
	}
	if err := c.artSvc.RenameEntity(ctx, &body); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactRenameSuccess, nil)
}

// @Summary 移动并/或重命名制品实体
// @Description 同时支持移动和重命名 asset 或 group。targetGroupID/newName 任一传入即生效；
// @Description 与当前一致或为空则跳过对应操作；两者都不变则返回错误。整个过程在单事务内完成。
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
	if err := c.artSvc.RelocateEntity(ctx, &body); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactRelocateSuccess, nil)
}
