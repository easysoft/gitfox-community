package artifact

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/req"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	artifactSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/artifact"
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

// @Summary 查询制品包列表
// @Description 列出指定空间/视图下的制品包
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SpaceRepoReq true "查询范围"
// @Param format query string false "制品格式"
// @Param query query string false "查询关键字"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/packages [post]
func (c *Controller) ListPackages(ctx *gin.Context) {
	var reqBody model.SpaceRepoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	filter := &model.ArtifactFilter{
		Format: req.QueryParamOrDefault(ctx, "format", ""),
		Query:  req.QueryParamOrDefault(ctx, "query", ""),
	}
	list, err := c.artSvc.ListPackages(ctx, view.ViewID, filter)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, list)
}

// @Summary 创建视图
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
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.CreateView(ctx, &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactCreateViewSuccess, view)
}

// @Summary 查询版本列表
// @Description 列出指定制品包的版本列表
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
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody.SpaceRepoReq)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	versions, err := c.artSvc.ListVersions(ctx, view.ViewID,
		reqBody.PackageName, reqBody.GroupName, model.ArtifactFormat(reqBody.Format), model.SearchVersionOption{})
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, versions)
}

// @Summary 查询版本文件列表
// @Description 列出指定版本的制品文件列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param versionId path int true "版本ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/versions/{versionId}/assets [get]
func (c *Controller) ListAssets(ctx *gin.Context) {
	versionId, err := req.PathParamAsPositiveInt64(ctx, "versionId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	assets, err := c.artSvc.ListAssets(ctx, versionId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, assets)
}

// @Summary 查询分组文件列表
// @Description 列出指定分组的制品文件列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param groupId path int true "分组ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups/{groupId}/assets [get]
func (c *Controller) ListGroupAssets(ctx *gin.Context) {
	groupId, err := req.PathParamAsPositiveInt64(ctx, "groupId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	assets, err := c.artSvc.ListGroupAssets(ctx, groupId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, assets)
}

// @Summary 查询文件详情
// @Description 获取指定制品文件的详细信息
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param assetId path int true "文件ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetId} [get]
func (c *Controller) GetAsset(ctx *gin.Context) {
	assetId, err := req.PathParamAsPositiveInt64(ctx, "assetId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	asset, err := c.artSvc.GetAsset(ctx, assetId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactQuerySuccess, asset)
}

// @Summary 查询文件历史
// @Description 获取指定制品文件的历史 blob 列表
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param assetId path int true "文件ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetId}/history [get]
func (c *Controller) GetBlobHistory(ctx *gin.Context) {
	assetId, err := req.PathParamAsPositiveInt64(ctx, "assetId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	blobs, err := c.artSvc.GetBlobHistory(ctx, assetId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactQuerySuccess, blobs)
}

// @Summary 下载文件
// @Description 下载制品文件
// @Tags artifacts
// @Produce application/octet-stream
// @Param assetId path int true "文件ID"
// @Success 200 {file} file "文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/assets/{assetId}/download [get]
func (c *Controller) DownloadAsset(ctx *gin.Context) {
	assetId, err := req.PathParamAsPositiveInt64(ctx, "assetId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	reader, blob, err := c.artSvc.DownloadAsset(ctx, assetId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	defer reader.Close()

	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", blob.Ref))
	ctx.Header("Content-Length", strconv.FormatInt(blob.Size, 10))
	ctx.DataFromReader(200, blob.Size, "application/octet-stream", reader, nil)
}

// @Summary 上传制品
// @Description 上传制品文件（multipart/form-data）
// @Tags artifacts
// @Accept multipart/form-data
// @Produce application/json
// @Param format path string true "制品格式"
// @Param spaceId formData int false "空间ID"
// @Param repoId formData int false "仓库ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/upload/{format} [post]
func (c *Controller) Upload(ctx *gin.Context) {
	var reqBody model.SpaceRepoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	formatStr := ctx.Param("format")
	format := model.ArtifactFormat(formatStr)

	assets, err := c.artSvc.Upload(ctx, view, format, ctx.Request)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactUploadSuccess, assets)
}

// @Summary 创建分组
// @Description 创建 group 节点
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.CreateGroupReq true "请求参数(包含 spaceId, repoId, names 和 format)"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups [post]
func (c *Controller) CreateGroup(ctx *gin.Context) {
	var params model.CreateGroupReq
	if err := ctx.ShouldBindBodyWithJSON(&params); err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	reqBody := &model.SpaceRepoReq{SpaceID: params.SpaceID, RepoID: params.RepoID}
	view, err := c.artSvc.ParseSpaceView(ctx, reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}

	names := strings.Split(params.Names, ".")
	ge, err := c.artSvc.CreateGroup(ctx, view.ViewID, names, model.ArtifactFormat(params.Format))
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactCreateGroupSuccess, ge)
}

// @Summary 查询格式列表
// @Description 列出指定空间的已用格式
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SpaceRepoReq true "查询范围"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/formats [post]
func (c *Controller) ListFormats(ctx *gin.Context) {
	var reqBody model.SpaceRepoReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	view, err := c.artSvc.ParseSpaceView(ctx, &reqBody)
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	formats, err := c.artSvc.ListFormats(ctx, view.ViewID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, formats)
}

// @Summary 查询树节点列表
// @Description 列出树节点的子节点
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param nodeId path int true "节点ID"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/nodes/{nodeId}/children [get]
func (c *Controller) ListTreeNodes(ctx *gin.Context) {
	parentId, err := req.PathParamAsPositiveInt64(ctx, "nodeId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	nodes, err := c.artSvc.ListTreeNodes(ctx, parentId)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactListSuccess, nodes)
}

// @Summary 容量统计
// @Description 返回非容器格式制品容量统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/capacity/all [get]
func (c *Controller) StatisticCapacity(ctx *gin.Context) {
	data, err := c.artSvc.StatisticCapacity(ctx)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
}

// @Summary 容器容量统计
// @Description 返回 container 格式制品容量统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/capacity/container [get]
func (c *Controller) StatisticContainerCapacity(ctx *gin.Context) {
	data, err := c.artSvc.StatisticContainerCapacity(ctx)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
}

// @Summary 制品汇总统计
// @Description 返回制品资产汇总统计
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/statistic [get]
func (c *Controller) StatisticSummary(ctx *gin.Context) {
	data, err := c.artSvc.StatisticSummary(ctx)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactStatisticSuccess, data)
}

// @Summary 按实体批量软删除制品
// @Description 按 entity_id 批量软删除 asset/version/package/group
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Param body body model.SoftRemoveEntitiesReq true "实体ID列表"
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/entities [delete]
func (c *Controller) SoftRemoveByEntity(ctx *gin.Context) {
	var reqBody model.SoftRemoveEntitiesReq
	if err := ctx.ShouldBindBodyWithJSON(&reqBody); err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	spaceID, _, err := req.QueryParamAsPositiveInt64(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	data, err := c.artSvc.SoftRemove(ctx, spaceID, &reqBody)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactDeleteSuccess, data)
}

// @Summary 归档 generic 分组
// @Description 将 generic 分组资产打包为 zip 下载
// @Tags artifacts
// @Produce application/zip
// @Param groupId path int true "分组ID"
// @Param spaceId query int true "空间ID"
// @Success 200 {file} file "ZIP 文件流"
// @Failure 400 {object} resp.ObjectResponse "参数错误"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/groups/{groupId}/archive [get]
func (c *Controller) ArchiveGroup(ctx *gin.Context) {
	groupID, err := req.PathParamAsPositiveInt64(ctx, "groupId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	spaceID, err := req.QueryParamAsPositiveInt64OrError(ctx, "spaceId")
	if err != nil {
		resp.BadRequest(ctx, err)
		return
	}
	session, _ := req.GetAuthSession(ctx)
	var repoID int64
	if rID, _, _ := req.QueryParamAsPositiveInt64(ctx, "repoId"); rID > 0 {
		repoID = rID
	}
	archive, err := c.artSvc.PrepareArchiveGroup(ctx, session, spaceID, repoID, groupID)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	ctx.Header("Content-Type", "application/zip")
	ctx.Header("Content-Disposition", "attachment; filename="+archive.Name())
	ctx.Status(http.StatusOK)
	if err := c.artSvc.ArchiveGroup(ctx, archive, ctx.Writer); err != nil {
		resp.InternalError(ctx, err)
		return
	}
}

// @Summary 触发容器垃圾回收
// @Description 创建 container 制品垃圾回收后台任务
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/garbage/container [get]
func (c *Controller) GarbageCollectContainer(ctx *gin.Context) {
	data, err := c.artSvc.TriggerGarbageCollectContainer(ctx)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactGarbageCollectSuccess, data)
}

// @Summary 触发软删垃圾回收
// @Description 创建已软删制品垃圾回收后台任务
// @Tags artifacts
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/garbage/soft-remove [get]
func (c *Controller) GarbageCollectSoftRemove(ctx *gin.Context) {
	data, err := c.artSvc.TriggerGarbageCollectSoftRemove(ctx)
	if err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactGarbageCollectSuccess, data)
}

// @Summary 重建制品树
// @Description 重建所有制品树节点
// @Tags artifacts
// @Accept application/json
// @Produce application/json
// @Success 200 {object} resp.ObjectResponse "成功"
// @Failure 500 {object} resp.ObjectResponse "服务器错误"
// @Router /artifacts/tree/refresh [post]
func (c *Controller) RefreshTree(ctx *gin.Context) {
	if err := c.artSvc.RefreshTree(ctx); err != nil {
		resp.InternalError(ctx, err)
		return
	}
	resp.OK(ctx, i18n.ArtifactRefreshTreeSuccess, nil)
}
