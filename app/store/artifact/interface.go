package artifact

import (
	"context"
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"

	"gorm.io/gorm"
)

// SearchOption 是 artifact 查询的 functional option。
// SearchOption is a functional option for searching artifacts.
type SearchOption interface {
	Apply(db *gorm.DB) *gorm.DB
}

// Interface 是 artifact store 的主入口，聚合各类子仓储与跨表查询。
// Interface is the main artifact store interface.
type Interface interface {
	Groups() GroupInterface
	Views() ViewInterface
	Packages() PackageInterface
	Versions() VersionInterface
	Assets() AssetInterface
	MetaAssets() MetaAssetInterface
	Blobs() BlobInterface
	Nodes() TreeNodeInterface
	FindPackages(ctx context.Context, artifactID int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error)
	GetVersion(ctx context.Context, artifactID int64, packageName, groupName, versionName string, format model.ArtifactFormat) (*model.ArtifactVersion, error)
	GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetsRes, error)
	ListAssets(ctx context.Context, VersionID int64) ([]*model.ArtifactAssetsRes, error)
	ListGroupAssets(ctx context.Context, groupID int64) ([]*model.ArtifactAssetsRes, error)
	StatisticAssets(ctx context.Context, retentionTime time.Duration, options ...SearchOption) (*model.ArtifactStatisticReport, error)
}

type ViewInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactView) (*model.ArtifactView, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*model.ArtifactView, error)
	GetDefault(ctx context.Context, spaceID int64) (*model.ArtifactView, error)
	GetDefaultByRepoAndSpace(ctx context.Context, repoID, spaceID int64) (*model.ArtifactView, error)
	GetByName(ctx context.Context, spaceID, repoID int64, name string) (*model.ArtifactView, error)

	// FindByNameInScope 在指定作用域下查询所有同名 view（不包含已软删的）。
	// 返回结果包含 active（deleted=0）
	FindByNameInScope(ctx context.Context, spaceID, repoID int64, name string) ([]*model.ArtifactView, error)

	// SoftDelete 软删除视图（将 deleted 置为 1）。
	// 注意：GORM 的 autoUpdateTime 会自动更新 editedDate，
	// 因此 editedDate 即为删除时间，用于回收站保留期计算。
	SoftDelete(ctx context.Context, view *model.ArtifactView) error

	// FindExpiredIDs 查询 editedDate 早于 before 的已软删视图 ID 列表。
	// 只做查询，不删除，供上层先清理关联资源、再逐条删除 view 记录的正确流程使用。
	FindExpiredIDs(ctx context.Context, before time.Time) ([]int64, error)

	// HardDeleteExpired 硬删除 editedDate 早于 before 的所有已软删视图。
	// 返回被删除的 view ID 列表，供上层级联清理关联资源。
	// Deprecated: 直接使用会先删 view 再清理资源，崩溃时产生孤儿。
	// 请改用 FindExpiredIDs + hardDeleteViewResources + Delete。
	HardDeleteExpired(ctx context.Context, before time.Time) ([]int64, error)
}

type PackageInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactPackage) (*model.ArtifactPackage, error)
	GetByID(ctx context.Context, packageId int64) (*model.ArtifactPackage, error)
	GetByName(ctx context.Context, name, namespace string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactPackage, error)
	List(ctx context.Context, artifactID int64, includeDeleted bool) ([]*model.ArtifactPackage, error)
	ListBatch(ctx context.Context, includeDeleted bool, batchSize int, fn func(pkg *model.ArtifactPackage) error) error
	ListByNamespace(ctx context.Context, artifactID int64, namespace string, format model.ArtifactFormat, recurse, includeDeleted bool) ([]*model.ArtifactPackage, error)
	Update(ctx context.Context, upObj *model.ArtifactPackage) error
	DeleteById(ctx context.Context, packageId int64) error
	DeleteByIds(ctx context.Context, packageIds ...int64) (int64, error)
	SoftDelete(ctx context.Context, pkg *model.ArtifactPackage) error
	UnDelete(ctx context.Context, pkg *model.ArtifactPackage) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactPackage, error)
	GetByAssetID(ctx context.Context, assetID int64) (*model.ArtifactPackage, error)
}

type VersionInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactVersion) (*model.ArtifactVersion, error)
	GetByID(ctx context.Context, VersionID int64) (*model.ArtifactVersion, error)
	GetByVersion(ctx context.Context, packageId int64, version string) (*model.ArtifactVersion, error)
	Update(ctx context.Context, ver *model.ArtifactVersion) error
	DeleteById(ctx context.Context, VersionID int64) error
	DeleteByIds(ctx context.Context, VersionIDs ...int64) (int64, error)
	Find(ctx context.Context, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error)
	SoftDelete(ctx context.Context, ver *model.ArtifactVersion) error
	UnDelete(ctx context.Context, ver *model.ArtifactVersion) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactVersionInfo, error)
}

type AssetInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactAsset) (*model.ArtifactAsset, error)
	GetById(ctx context.Context, assetID int64) (*model.ArtifactAsset, error)
	GetPath(ctx context.Context, path string, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetMetaAsset(ctx context.Context, path string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetVersionAsset(ctx context.Context, path string, VersionID int64) (*model.ArtifactAsset, error)
	GetGroupAsset(ctx context.Context, path string, groupID int64) (*model.ArtifactAsset, error)
	FindMain(ctx context.Context, VersionID int64) ([]*model.ArtifactAsset, error)
	UpdateBlobId(ctx context.Context, assetID int64, blobID int64) error
	Update(ctx context.Context, asset *model.ArtifactAsset) error
	SoftDeleteById(ctx context.Context, assetID int64) error
	SoftDeleteExcludeById(ctx context.Context, VersionID int64, assetID int64) error
	DeleteById(ctx context.Context, assetID int64) error
	Search(ctx context.Context, option ...SearchOption) ([]*model.ArtifactAsset, error)
	SearchExtendBlob(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAssetExtendBlob, error)
	DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error)
}

type MetaAssetInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactMetaAsset) (*model.ArtifactMetaAsset, error)
	GetById(ctx context.Context, metaAssetId int64) (*model.ArtifactMetaAsset, error)
	GetByPath(ctx context.Context, path string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactMetaAsset, error)
	Update(ctx context.Context, metaAsset *model.ArtifactMetaAsset) error
	DeleteById(ctx context.Context, metaAssetId int64) error
	DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error)
}

type BlobInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactBlob) (*model.ArtifactBlob, error)
	GetById(ctx context.Context, blobID int64) (*model.ArtifactBlob, error)
	GetByRef(ctx context.Context, ref string, storageID int64) (*model.ArtifactBlob, error)
	GetByHistory(ctx context.Context, assetID, createdAt int64) (*model.ArtifactBlob, error)
	Update(ctx context.Context, upObj *model.ArtifactBlob) error
	UpdateOptLock(ctx context.Context, upObj *model.ArtifactBlob, mutateFn func(blob *model.ArtifactBlob) error) (*model.ArtifactBlob, error)
	SoftDeleteById(ctx context.Context, blobID int64) error
	DeleteById(ctx context.Context, blobID int64) error
	History(ctx context.Context, assetID int64) ([]*model.ArtifactBlobRes, error)
}

type TreeNodeInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactTreeNode) (*model.ArtifactTreeNode, error)
	RecurseCreate(ctx context.Context, obj *model.ArtifactTreeNode) error
	GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error)
	GetByPath(ctx context.Context, artifactID int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error)
	GetByTarget(ctx context.Context, artifactID int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error)
	ListFormats(ctx context.Context, artifactID int64) ([]*model.ArtifactTreeNode, error)
	ListByParentId(ctx context.Context, parentID int64) ([]*model.ArtifactTreeNode, error)
	Update(ctx context.Context, upObj *model.ArtifactTreeNode) error
	DeleteById(ctx context.Context, nodeID int64) error
	RecurseDeleteById(ctx context.Context, nodeID int64) error
	DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error)
}

type GroupInterface interface {
	// Create 在指定 view 下创建一个 artifact 分组。
	Create(ctx context.Context, newObj *model.CreateArtifactGroup) (*model.ArtifactGroup, error)

	// GetById 按主键查找分组。
	GetById(ctx context.Context, groupID int64) (*model.ArtifactGroup, error)

	// GetRoot 在指定 view 下查找指定名称和格式的根分组（parentID IS NULL）。
	// 供 container push 时通过 (artifactID, groupName) 精确定位 namespace group。
	GetRoot(ctx context.Context, artifactID int64, name string, format model.ArtifactFormat) (*model.ArtifactGroup, error)

	// GetSubNode 查找指定父分组下的子分组（按名称）。
	GetSubNode(ctx context.Context, parentID int64, name string) (*model.ArtifactGroup, error)

	// GetEndpointByNames 按路径名数组查找（或递归建立）分组链路，返回末级分组及路径。
	// 用于 raw/generic 格式按命名空间路径索引分组。
	GetEndpointByNames(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)

	// GetEndpointById 按 group ID 查找分组端点（含完整路径信息）。
	GetEndpointById(ctx context.Context, id int64) (*model.ArtifactGroupEndpoint, error)

	// RecursiveCreate 按名称路径递归创建分组，返回末级分组端点。
	// 用于 raw 格式在上传时自动建立 namespace 目录树。
	RecursiveCreate(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)

	// All 返回所有未软删的分组（不推荐在生产路径直接使用，仅供管理/同步任务）。
	All(ctx context.Context) ([]*model.ArtifactGroup, error)

	// ListDescendantsForOwner 列出指定 view 下指定格式的所有后代分组（用于树形展示）。
	ListDescendantsForOwner(ctx context.Context, artifactID int64, format model.ArtifactFormat) ([]*model.ArtifactGroup, error)

	// ListDescendants 列出指定分组的所有后代分组（不限格式）。
	ListDescendants(ctx context.Context, id int64) ([]*model.ArtifactGroup, error)

	// SoftDeleteById 软删除分组（将 deleted 置为 1）。
	SoftDeleteById(ctx context.Context, groupID int64) error

	// DeleteByArtifactID 物理删除指定 view 下的所有分组，供 view 删除时级联使用。
	DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error)
}
