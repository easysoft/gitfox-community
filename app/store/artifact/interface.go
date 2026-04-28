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
	FindPackages(ctx context.Context, viewId int64, filter *model.ArtifactFilter) ([]*model.ArtifactListItem, error)
	GetVersion(ctx context.Context, viewId int64, packageName, groupName, versionName string, format model.ArtifactFormat) (*model.ArtifactVersion, error)
	GetAsset(ctx context.Context, assetId int64) (*model.ArtifactAssetsRes, error)
	ListAssets(ctx context.Context, versionId int64) ([]*model.ArtifactAssetsRes, error)
	ListGroupAssets(ctx context.Context, groupId int64) ([]*model.ArtifactAssetsRes, error)
	StatisticAssets(ctx context.Context, retentionTime time.Duration, options ...SearchOption) (*model.ArtifactStatisticReport, error)
}



type ViewInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactView) error
	Delete(ctx context.Context, id int64) error
	GetDefault(ctx context.Context, spaceId int64) (*model.ArtifactView, error)
	GetDefaultByRepoAndSpace(ctx context.Context, repoId, spaceId int64) (*model.ArtifactView, error)
	GetByName(ctx context.Context, spaceId, repoId int64, name string) (*model.ArtifactView, error)

	// FindByNameInScope 在指定作用域（spaceId + repoId）下查询所有同名 view（含已软删的）。
	// 用于 CreateView 时判断是否存在活跃的或在回收站保留期内的同名 view。
	FindByNameInScope(ctx context.Context, spaceId, repoId int64, name string) ([]*model.ArtifactView, error)

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
	Create(ctx context.Context, newObj *model.ArtifactPackage) error
	GetByID(ctx context.Context, packageId int64) (*model.ArtifactPackage, error)
	GetByName(ctx context.Context, name, namespace string, viewId int64, format model.ArtifactFormat) (*model.ArtifactPackage, error)
	List(ctx context.Context, viewId int64, includeDeleted bool) ([]*model.ArtifactPackage, error)
	ListBatch(ctx context.Context, includeDeleted bool, batchSize int, fn func(pkg *model.ArtifactPackage) error) error
	ListByNamespace(ctx context.Context, viewId int64, namespace string, format model.ArtifactFormat, recurse, includeDeleted bool) ([]*model.ArtifactPackage, error)
	Update(ctx context.Context, upObj *model.ArtifactPackage) error
	DeleteById(ctx context.Context, packageId int64) error
	DeleteByIds(ctx context.Context, packageIds ...int64) (int64, error)
	SoftDelete(ctx context.Context, pkg *model.ArtifactPackage) error
	UnDelete(ctx context.Context, pkg *model.ArtifactPackage) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactPackage, error)
	GetByAssetID(ctx context.Context, assetId int64) (*model.ArtifactPackage, error)
}

type VersionInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactVersion) error
	GetByID(ctx context.Context, versionId int64) (*model.ArtifactVersion, error)
	GetByVersion(ctx context.Context, packageId, viewId int64, version string) (*model.ArtifactVersion, error)
	Update(ctx context.Context, ver *model.ArtifactVersion) error
	DeleteById(ctx context.Context, versionId int64) error
	DeleteByIds(ctx context.Context, versionIds ...int64) (int64, error)
	Find(ctx context.Context, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error)
	SoftDelete(ctx context.Context, ver *model.ArtifactVersion) error
	UnDelete(ctx context.Context, ver *model.ArtifactVersion) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactVersionInfo, error)
}

type AssetInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactAsset) error
	GetById(ctx context.Context, assetId int64) (*model.ArtifactAsset, error)
	GetPath(ctx context.Context, path string, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetMetaAsset(ctx context.Context, path string, viewId int64, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetVersionAsset(ctx context.Context, path string, versionId int64) (*model.ArtifactAsset, error)
	GetGroupAsset(ctx context.Context, path string, groupId int64) (*model.ArtifactAsset, error)
	FindMain(ctx context.Context, versionId int64) ([]*model.ArtifactAsset, error)
	UpdateBlobId(ctx context.Context, assetId int64, blobId int64) error
	Update(ctx context.Context, asset *model.ArtifactAsset) error
	SoftDeleteById(ctx context.Context, assetId int64) error
	SoftDeleteExcludeById(ctx context.Context, versionId int64, assetId int64) error
	DeleteById(ctx context.Context, assetId int64) error
	Search(ctx context.Context, option ...SearchOption) ([]*model.ArtifactAsset, error)
	SearchExtendBlob(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAssetExtendBlob, error)
	DeleteByViewID(ctx context.Context, viewId int64) (int64, error)
}

type MetaAssetInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactMetaAsset) error
	GetById(ctx context.Context, metaAssetId int64) (*model.ArtifactMetaAsset, error)
	GetByPath(ctx context.Context, path string, viewId int64, format model.ArtifactFormat) (*model.ArtifactMetaAsset, error)
	Update(ctx context.Context, metaAsset *model.ArtifactMetaAsset) error
	DeleteById(ctx context.Context, metaAssetId int64) error
	DeleteByViewID(ctx context.Context, viewId int64) (int64, error)
}

type BlobInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactBlob) error
	GetById(ctx context.Context, blobId int64) (*model.ArtifactBlob, error)
	GetByRef(ctx context.Context, ref string, storageId int64) (*model.ArtifactBlob, error)
	GetByHistory(ctx context.Context, assetId, createdAt int64) (*model.ArtifactBlob, error)
	Update(ctx context.Context, upObj *model.ArtifactBlob) error
	UpdateOptLock(ctx context.Context, upObj *model.ArtifactBlob, mutateFn func(blob *model.ArtifactBlob) error) (*model.ArtifactBlob, error)
	SoftDeleteById(ctx context.Context, blobId int64) error
	DeleteById(ctx context.Context, blobId int64) error
	History(ctx context.Context, assetId int64) ([]*model.ArtifactBlobRes, error)
}

type TreeNodeInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactTreeNode) error
	RecurseCreate(ctx context.Context, obj *model.ArtifactTreeNode) error
	GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error)
	GetByPath(ctx context.Context, viewId int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error)
	GetByTarget(ctx context.Context, viewId int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error)
	ListFormats(ctx context.Context, viewId int64) ([]*model.ArtifactTreeNode, error)
	ListByParentId(ctx context.Context, parentId int64) ([]*model.ArtifactTreeNode, error)
	Update(ctx context.Context, upObj *model.ArtifactTreeNode) error
	DeleteById(ctx context.Context, nodeId int64) error
	RecurseDeleteById(ctx context.Context, nodeId int64) error
	DeleteByViewID(ctx context.Context, viewId int64) (int64, error)
}

type GroupInterface interface {
	Create(ctx context.Context, newObj *model.ArtifactGroup) error
	GetById(ctx context.Context, groupId int64) (*model.ArtifactGroup, error)
	GetRoot(ctx context.Context, viewId int64, name string, format model.ArtifactFormat) (*model.ArtifactGroup, error)
	GetSubNode(ctx context.Context, parentId int64, name string) (*model.ArtifactGroup, error)
	GetEndpointByNames(ctx context.Context, viewId int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)
	GetEndpointById(ctx context.Context, id int64) (*model.ArtifactGroupEndpoint, error)
	RecursiveCreate(ctx context.Context, viewId int64, names []string, format model.ArtifactFormat) (*model.ArtifactGroupEndpoint, error)
	All(ctx context.Context) ([]*model.ArtifactGroup, error)
	ListDescendantsForOwner(ctx context.Context, viewId int64, format model.ArtifactFormat) ([]*model.ArtifactGroup, error)
	ListDescendants(ctx context.Context, id int64) ([]*model.ArtifactGroup, error)
	SoftDeleteById(ctx context.Context, groupId int64) error
	DeleteByViewID(ctx context.Context, viewId int64) (int64, error)
}
