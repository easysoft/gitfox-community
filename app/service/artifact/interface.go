package artifact

import (
	"context"
	"io"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter/container"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

// JobScheduler 是后台任务调度器接口，用于触发异步 GC 等任务。
type JobScheduler interface {
	RunJob(ctx context.Context, jobName, jobType string, data string, maxDurationSeconds int) error
}

// Interface 是制品库服务的对外接口定义。
// 所有制品库的业务能力通过此接口暴露给 controller 和其他服务。
type Interface interface {
	// ---- 视图管理 ----

	// CreateView 创建制品库视图，支持全局/空间/仓库三级作用域。
	CreateView(ctx context.Context, req *model.CreateArtifactViewReq) (*model.ArtifactView, error)

	// ResolveView 按请求中的 artifactID 或默认视图规则解析实际视图。
	// 规则：
	//  1. artifactID > 0：优先按 ID 精确取库
	//  2. artifactID == 0 或未传：继续按 spaceId/repoId 走默认 view 逻辑
	ResolveView(ctx context.Context, req *model.SpaceRepoReq) (*model.ArtifactView, error)

	// ParseSpaceView 解析空间+仓库参数，返回视图描述符（包含 artifactID 和路由信息）。
	ParseSpaceView(ctx context.Context, req *model.SpaceRepoReq) (*adapter.ViewDescriptor, error)

	// ParseContainerView 按作用域和视图名称解析 ViewDescriptor，供容器镜像路由使用。
	// scope: "global" | "space" | "repo"
	// spaceName / scopeRepoName 在对应 scope 下有值，global 时均为空字符串。
	// view 不存在时直接返回 404，不自动创建。
	ParseContainerView(ctx context.Context, scope, spaceName, scopeRepoName, viewname string) (*adapter.ViewDescriptor, error)

	// ---- 制品查询 ----

	// ListPackages 列出指定视图下的制品包，支持格式、关键字过滤。
	ListPackages(ctx context.Context, artifactID int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error)

	// ListVersions 列出制品包的所有版本。
	ListVersions(ctx context.Context, artifactID int64, packageName, groupName string, format model.ArtifactFormat, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error)

	// ListAssets 列出指定版本下的所有制品资源文件。
	ListAssets(ctx context.Context, versionId int64) ([]*model.ArtifactAssetsRes, error)

	// ListGroupAssets 列出指定分组下直接挂载的资源文件（generic 格式）。
	ListGroupAssets(ctx context.Context, groupId int64) ([]*model.ArtifactAssetsRes, error)

	// GetAsset 获取单个制品资源的详细信息。
	GetAsset(ctx context.Context, assetId int64) (*model.ArtifactAssetsRes, error)

	// GetBlobHistory 获取制品资源的 blob 历史记录。
	GetBlobHistory(ctx context.Context, assetId int64) ([]*model.ArtifactBlobRes, error)

	// ---- 上传/下载 ----

	// DownloadAsset 下载制品资源，返回文件流和 blob 元信息。
	DownloadAsset(ctx context.Context, assetId int64) (io.ReadCloser, *model.ArtifactBlob, error)

	// Upload 上传制品，自动根据 format 分发到对应的适配器处理。
	Upload(ctx context.Context, view *adapter.ViewDescriptor, format model.ArtifactFormat, req *http.Request) ([]*adapter.SavedAsset, error)

	// ---- 分组管理 ----

	// ViewExists 检查指定 artifactID 是否存在（仅 active view）。
	ViewExists(ctx context.Context, artifactID int64) error

	// GroupPathExists 检查指定视图下是否已存在完全相同的分组路径。
	GroupPathExists(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (bool, error)

	// CreateGroup 在指定视图下创建分组层级。
	CreateGroup(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)

	// ---- 制品树 ----

	// ListFormats 列出指定视图下的格式根节点。
	ListFormats(ctx context.Context, artifactID int64) ([]*model.ArtifactTreeNode, error)

	// ListTreeNodes 列出指定父节点下的子节点。
	ListTreeNodes(ctx context.Context, parentId int64) ([]*model.ArtifactTreeNode, error)

	// RefreshTree 全量重建制品树节点（groups → packages → versions）。
	RefreshTree(ctx context.Context) error

	// ---- 删除操作 ----

	// SoftRemove 软删除指定的实体节点（group/package/version/asset），支持级联删除子节点。
	SoftRemove(ctx context.Context, spaceID int64, req *model.SoftRemoveEntitiesReq) (*model.ArtifactNodeRemoveReport, error)

	// ---- 归档下载 ----

	// PrepareArchiveGroup 准备分组归档下载所需的数据（鉴权 + 数据组装）。
	PrepareArchiveGroup(ctx context.Context, session *authModel.Session, spaceID, repoID, groupID int64) (*ArchiveGroup, error)

	// ArchiveGroup 将已准备好的分组数据写入 zip 流。
	ArchiveGroup(ctx context.Context, archive *ArchiveGroup, output io.Writer) error

	// ---- 统计与 GC ----

	// StatisticCapacity 统计全量制品资源的容量使用情况。
	StatisticCapacity(ctx context.Context) (*model.ArtifactStatisticResponse, error)

	// StatisticContainerCapacity 统计容器镜像格式的容量使用情况。
	StatisticContainerCapacity(ctx context.Context) (*model.ArtifactStatisticResponse, error)

	// StatisticSummary 汇总各格式的制品数量和容量摘要。
	StatisticSummary(ctx context.Context) (map[string]*model.ArtifactStatsObject, error)

	// TriggerGarbageCollectContainer 触发容器镜像格式的 GC 异步任务。
	TriggerGarbageCollectContainer(ctx context.Context) (*model.JobUIDResponse, error)

	// TriggerGarbageCollectSoftRemove 触发软删资源的 GC 异步任务。
	TriggerGarbageCollectSoftRemove(ctx context.Context) (*model.JobUIDResponse, error)

	// GarbageCollectContainer 执行容器镜像 GC，清理过期 blob。
	GarbageCollectContainer(ctx context.Context) (*model.ArtifactGarbageReportRes, error)

	// GarbageCollectSoftRemove 执行软删资源 GC，清理过期的 assets/versions/packages。
	GarbageCollectSoftRemove(ctx context.Context) (*model.ArtifactGarbageReportRes, error)

	// GarbageCollectExpiredViews 清理超过回收站保留期的已软删视图及其全部关联资源。
	GarbageCollectExpiredViews(ctx context.Context) (int, error)

	// ---- 容器镜像 ----

	// ListContainerImages 列出容器镜像的 tag 元数据。
	ListContainerImages(ctx context.Context, view *adapter.ViewDescriptor, packageName, version string) (*container.TagMetadata, error)

	// ---- 权限管理 ----

	// SetEntityPermission 设置实体节点级权限（package/version/group 级别）。
	SetEntityPermission(ctx context.Context, spaceID int64, entityId string, permReq *model.PermissionRequest) (interface{}, error)

	// GetEntityPermissions 获取实体节点级权限配置（含继承信息）。
	GetEntityPermissions(ctx context.Context, spaceID int64, entityId string) (*model.ArtifactEntityPermissions, error)

	// SetSpacePermission 设置空间/仓库级权限。
	SetSpacePermission(ctx context.Context, spaceID, repoID int64, permReq *model.PermissionRequest) (*model.ArtifactPermissions, error)

	// GetSpacePermissions 获取空间/仓库级权限配置。
	GetSpacePermissions(ctx context.Context, spaceID, repoID int64) (*model.ArtifactPermissions, error)

	// GetSpaceMembers 获取空间/仓库级的制品库成员列表。
	GetSpaceMembers(ctx context.Context, spaceID, repoID int64) ([]*model.CasbinSubject, error)

	// FileStore 返回制品库内容存储，供容器镜像控制器直接使用。
	FileStore() storage.ContentStorage
}
