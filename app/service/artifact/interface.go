package artifact

import (
	"context"
	"io"
	"net/http"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/artifact/adapter"
	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	artifactStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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
	ResolveView(ctx context.Context, req *model.SpaceRepoReq) (*model.ArtifactView, error)

	// ParseSpaceView 解析空间+仓库参数，返回视图描述符（包含 artifactID 和路由信息）。
	ParseSpaceView(ctx context.Context, req *model.SpaceRepoReq) (*adapter.ViewDescriptor, error)

	// ---- 制品查询 ----

	// ListAssetPage 按 entityID（group / version / package / asset）分页查询。
	ListAssetPage(ctx context.Context, req *model.ListAssetReq) (*database.PageResult[*model.ArtifactAssetInfo], error)

	// GetAsset 获取单个制品资源的详细信息。
	GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetInfo, error)

	// ---- 上传/下载 ----

	// DownloadAsset 下载制品资源，返回文件流和 blob 元信息。
	DownloadAsset(ctx context.Context, assetID int64) (io.ReadCloser, *model.ArtifactBlob, string, error)

	// DownloadAssetByPath 按制品路径下载资源，返回文件流和 blob 元信息。
	// spaceCode 为空间 code，filepath 为 _format/viewName/groupPath/packageName/version/fileName
	DownloadAssetByPath(ctx context.Context, spaceCode string, filepath string) (io.ReadCloser, *model.ArtifactBlob, string, error)

	// Upload 上传制品，自动根据 format 分发到对应的适配器处理。
	Upload(ctx context.Context, view *adapter.ViewDescriptor, format model.ArtifactFormat, req *http.Request) ([]*adapter.SavedAsset, error)

	// ---- 分组管理 ----
	// ActiveGroupPathExists 检查指定视图下是否已存在未软删的同名分组路径。
	ActiveGroupPathExists(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (bool, error)

	// CreateGroup 在指定视图下创建分组层级。
	CreateGroup(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)

	// ---- 制品树 ----

	// ListTree 按 path/format/level 聚合查询制品树，兼容参考项目的万能树接口语义。
	ListTree(ctx context.Context, artifactID int64, filter *model.ArtifactTreeFilter) ([]*model.ArtifactTreeRes, error)

	// RefreshTree 全量重建制品树节点（groups → packages → versions）。
	// format 必填：传入具体 format 时仅刷新该 format 下的节点；传入 ArtifactAllFormat ("all") 时全量刷新。
	RefreshTree(ctx context.Context, format string) error

	// ---- 删除操作 ----

	// SoftRemove 软删除指定的实体节点（group/package/version/asset），支持级联删除子节点。
	SoftRemove(ctx context.Context, req *model.SoftRemoveEntitiesReq) (*model.ArtifactNodeRemoveReport, error)

	// ---- 回收站 ----

	// ListRecycleBin 按 kind 分页查询回收站记录（kind ∈ artifact/group/package/version/asset，默认 asset）。
	ListRecycleBin(ctx context.Context, req *model.ArtifactRecycleListReq) ([]*model.ArtifactRecycleItem, int64, error)

	// BatchRestore 批量恢复回收站中的实体（entityIDs 为 "kind.id" 串）。单条失败继续，最终返回汇总报告。
	BatchRestore(ctx context.Context, req *model.ArtifactEntitiesReq) (*model.ArtifactRestoreReport, error)

	// Restore 恢复回收站中的实体（entityID）
	Restore(ctx context.Context, req model.ArtifactEntitieReq) (*model.ArtifactRestoreRes, error)
	// ---- 归档下载 ----

	// PrepareArchiveGroup 准备分组归档下载所需的数据。
	PrepareArchiveGroup(ctx context.Context, session *authModel.Session, groupID int64) (*ArchiveGroup, error)

	// ArchiveGroup 将已准备好的分组数据写入 zip 流。
	ArchiveGroup(ctx context.Context, archive *ArchiveGroup, output io.Writer) error

	// ---- GC ----
	StatisticAssets(ctx context.Context, retentionTime time.Duration, options ...artifactStore.SearchOption) (*model.ArtifactStatisticReport, error)

	// TriggerGarbageCollect 触发指定格式或全量软删资源的 GC 异步任务。
	TriggerGarbageCollect(ctx context.Context, format string) (*model.JobUIDResponse, error)

	// GarbageCollect 执行指定格式或全量软删资源 GC（cron job / 用户主动 API 共用）。
	// format 留空 → 清全部格式；非空 → 仅清指定 format。
	GarbageCollect(ctx context.Context, format string) (*model.ArtifactGarbageReportRes, error)

	// ---- 容器镜像 ----

	// ResolveGlobalArtifact 按 artifactName 解析 /v2/global 路由对应的制品库上下文。
	ResolveGlobalArtifact(ctx context.Context, artifactName string) (*ContainerArtifactCtx, error)

	// ResolveSpaceArtifactByCode 按 (spaceCode, artifactName) 解析 /v2/{spaceCode} 路由对应的制品库上下文。
	ResolveSpaceArtifactByCode(ctx context.Context, spaceCode string, artifactName string) (*ContainerArtifactCtx, error)

	// ResolveRepoArtifact 按 (repoID, artifactName) 解析 /v2/repo{ID} 路由对应的制品库上下文。
	ResolveRepoArtifact(ctx context.Context, repoID uint64, artifactName string) (*ContainerArtifactCtx, error)

	// ---- 移动 & 重命名 ----

	// RelocateEntity 同时支持移动 + 重命名，单事务内完成；通用入口。
	RelocateEntity(ctx context.Context, req *model.RelocateEntityReq) error

	// FileStore 返回制品库内容存储，供容器镜像控制器直接使用。
	FileStore() storage.ContentStorage
}
