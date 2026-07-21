package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

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
	GetAsset(ctx context.Context, assetID int64) (*model.ArtifactAssetsRes, error)
	ListAssets(ctx context.Context, versionID int64) ([]*model.ArtifactAssetsRes, error)
	ListAssetsPage(ctx context.Context, req *model.ListAssetReq) (*database.PageResult[*model.ArtifactAssetsRes], error)
	ListGroupAssets(ctx context.Context, groupID int64) ([]*model.ArtifactAssetsRes, error)
}

type ViewInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactView) (*model.ArtifactView, error)

	// SoftDelete 软删除视图（将 deleted 置为 1），并把 operator 写入 editedBy（即"删除人"）。
	SoftDelete(ctx context.Context, view *model.ArtifactView, operator string) error

	UpdateName(ctx context.Context, id int64, name string) error

	GetByID(ctx context.Context, id int64) (*model.ArtifactView, error)            // 不包含已软删的
	GetByIDWithDeleted(ctx context.Context, id int64) (*model.ArtifactView, error) // 包含已软删的

	GetByName(ctx context.Context, spaceID, repoID int64, name string) (*model.ArtifactView, error)

	// GetByCode 按 (spaceID, repoID, code) 查询视图的唯一 code 标识。
	GetByCode(ctx context.Context, spaceID, repoID int64, code string) (*model.ArtifactView, error)

	// FindByNameInScope 在指定作用域下查询所有同名 view
	FindByNameInScope(ctx context.Context, spaceID, repoID int64, name string) ([]*model.ArtifactView, error) // 不包含已软删的

	// Restore 恢复制品库视图（将 deleted 置为 0），并把 operator 写入 editedBy（即"恢复人"）。
	Restore(ctx context.Context, viewID int64, operator string) error

	// ListRecycle 列出回收站中的制品库（deleted=1），按 editedDate 倒序分页。
	ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactView, int64, error)
}

type PackageInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactPackage) (*model.ArtifactPackage, error)
	GetByID(ctx context.Context, packageId int64) (*model.ArtifactPackage, error)
	GetByName(ctx context.Context, name, namespace string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactPackage, error)
	GetByNameInArtifactScope(ctx context.Context, name, namespace string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactPackage, error)
	LockArtifactScopePackage(ctx context.Context, artifactID int64, namespace, name string, format model.ArtifactFormat) (func(), error)
	List(ctx context.Context, artifactID int64, includeDeleted bool) ([]*model.ArtifactPackage, error)
	ListBatch(ctx context.Context, includeDeleted bool, batchSize int, fn func(pkg *model.ArtifactPackage) error) error
	ListByNamespace(ctx context.Context, artifactID int64, namespace string, format model.ArtifactFormat, recurse, includeDeleted bool) ([]*model.ArtifactPackage, error)
	Update(ctx context.Context, upObj *model.ArtifactPackage) error
	DeleteById(ctx context.Context, packageId int64) error
	DeleteByIds(ctx context.Context, packageIds ...int64) (int64, error)
	SoftDelete(ctx context.Context, pkg *model.ArtifactPackage, operator string) error
	// Restore 恢复制品包（将 deleted 置为 0），operator 写入 editedBy。
	Restore(ctx context.Context, packageID int64, operator string) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactPackage, error)

	// ListRecycle 列出回收站中的制品包（deleted=1），按 editedDate 倒序分页。
	ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactPackage, int64, error)

	// MoveArtifactLibByGroupIDs 跨 view 移动 group 子树时，批量改 group 下挂 package 的 artifactLibID。
	MoveArtifactLibByGroupIDs(ctx context.Context, groupIDs []int64, targetArtifactID int64) error
}

type VersionInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactVersion) (*model.ArtifactVersion, error)
	GetByID(ctx context.Context, versionID int64) (*model.ArtifactVersion, error)
	GetByVersion(ctx context.Context, packageID, artifactLibID int64, version string) (*model.ArtifactVersion, error)
	Update(ctx context.Context, ver *model.ArtifactVersion) error
	DeleteById(ctx context.Context, versionID int64) error
	DeleteByIds(ctx context.Context, versionIDs ...int64) (int64, error)
	Find(ctx context.Context, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error)
	SoftDelete(ctx context.Context, ver *model.ArtifactVersion, operator string) error
	// Restore 恢复制品版本（将 deleted 置为 0），operator 写入 editedBy。
	Restore(ctx context.Context, versionID int64, operator string) error
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactVersionInfo, error)

	// ListRecycle 列出回收站中的制品版本（deleted=1），按 editedDate 倒序分页。
	ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactVersion, int64, error)
}

type AssetInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactAsset) (*model.ArtifactAsset, error)
	GetById(ctx context.Context, assetID int64) (*model.ArtifactAsset, error)
	GetPath(ctx context.Context, path string, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetMetaAsset(ctx context.Context, path string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactAsset, error)
	GetVersionAsset(ctx context.Context, path string, versionID int64) (*model.ArtifactAsset, error)
	GetGroupAsset(ctx context.Context, path string, groupID int64) (*model.ArtifactAsset, error)
	FindMain(ctx context.Context, versionID int64) ([]*model.ArtifactAsset, error)
	UpdateBlobId(ctx context.Context, assetID int64, blobID int64) error
	Update(ctx context.Context, asset *model.ArtifactAsset) error
	SoftDeleteById(ctx context.Context, assetID int64, operator string) error
	SoftDeleteExcludeById(ctx context.Context, versionID int64, assetID int64, operator string) error
	// RestoreById 恢复制品资产（将 deleted 置为 0），operator 写入 editedBy。
	RestoreById(ctx context.Context, assetID int64, operator string) error
	DeleteById(ctx context.Context, assetID int64) error
	Search(ctx context.Context, option ...SearchOption) ([]*model.ArtifactAsset, error)
	SearchExtendBlob(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAssetExtendBlob, error)
	DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error)

	// ListRecycle 列出回收站中的制品资产（deleted=1），按 editedDate 倒序分页。
	ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactAsset, int64, error)

	// MoveArtifactLibByGroupIDs 跨 view 移动 group 子树时，批量改归属在该子树下的 asset
	// 的 artifactLibID。仅按 groupID 命中（直挂 asset 场景）。
	MoveArtifactLibByGroupIDs(ctx context.Context, groupIDs []int64, targetArtifactID int64) error
}

type MetaAssetInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactMetaAsset) (*model.ArtifactMetaAsset, error)
	DeleteById(ctx context.Context, metaAssetId int64) error
	Update(ctx context.Context, metaAsset *model.ArtifactMetaAsset) error
	GetById(ctx context.Context, metaAssetId int64) (*model.ArtifactMetaAsset, error)
	GetByPath(ctx context.Context, path string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactMetaAsset, error)
}

type BlobInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactBlob) (*model.ArtifactBlob, error)
	SoftDeleteById(ctx context.Context, blobID int64, operator string) error
	DeleteById(ctx context.Context, blobID int64) error
	Update(ctx context.Context, upObj *model.ArtifactBlob) error
	UpdateOptLock(ctx context.Context, upObj *model.ArtifactBlob, mutateFn func(blob *model.ArtifactBlob) error) (*model.ArtifactBlob, error)
	RestoreById(ctx context.Context, blobID int64, operator string) error
	GetById(ctx context.Context, blobID int64) (*model.ArtifactBlob, error)
	GetByRef(ctx context.Context, ref string, storageID int64) (*model.ArtifactBlob, error)

	GetActiveByChecksum(ctx context.Context, checksum string, storageID int64) (*model.ArtifactBlob, error)
}

type TreeNodeInterface interface {
	Create(ctx context.Context, newObj *model.CreateArtifactTreeNode) (*model.ArtifactTreeNode, error)
	DeleteById(ctx context.Context, nodeID int64) error
	RecurseDeleteById(ctx context.Context, nodeID int64) error
	Update(ctx context.Context, upObj *model.ArtifactTreeNode) error
	GetById(ctx context.Context, id int64) (*model.ArtifactTreeNode, error)
	GetByPath(ctx context.Context, artifactID int64, path string, format model.ArtifactFormat) (*model.ArtifactTreeNode, error)
	GetByTarget(ctx context.Context, artifactID int64, nodeType model.ArtifactTreeNodeType, id int64) (*model.ArtifactTreeNode, error)
	ListFormats(ctx context.Context, artifactID int64) ([]*model.ArtifactTreeNode, error)
	ListByParentId(ctx context.Context, parentID int64) ([]*model.ArtifactTreeNode, error)
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
	RecursiveCreate(ctx context.Context, artifactID int64, names []string, format model.ArtifactFormat, operator string) (*model.ArtifactGroupEndpoint, error)

	// All 返回所有未软删的分组（不推荐在生产路径直接使用，仅供管理/同步任务）。
	All(ctx context.Context) ([]*model.ArtifactGroup, error)

	// ListDescendantsForOwner 列出指定 view 下指定格式的所有后代分组（用于树形展示）。
	ListDescendantsForOwner(ctx context.Context, artifactID int64, format model.ArtifactFormat) ([]*model.ArtifactGroup, error)

	// ListDescendants 列出指定分组的所有后代分组（不限格式）。
	ListDescendants(ctx context.Context, id int64) ([]*model.ArtifactGroup, error)

	// ListChildren 列出指定分组的直接子分组，包含已软删记录。
	ListChildren(ctx context.Context, id int64) ([]*model.ArtifactGroup, error)

	// Search 按 SearchOption 组合条件查询分组列表。
	// GC 链路依赖该方法按「软删 + 过保留期」筛选过期 group。
	Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactGroup, error)

	// FindRootByName 按 (artifactID, format, name) 查找根级 group（parentID=0），含已软删记录。
	// 主要用于回收站恢复链路：CreateView 自动建的默认根 group 与 view 同名同 format，
	// 软删后通过此方法精确定位以便对称恢复。
	FindRootByName(ctx context.Context, artifactID int64, name string, format model.ArtifactFormat) (*model.ArtifactGroup, error)

	// SoftDeleteById 软删除分组（将 deleted 置为 1），并把 operator 写入 editedBy（即"删除人"）。
	SoftDeleteById(ctx context.Context, groupID int64, operator string) error

	// RestoreById 恢复分组（将 deleted 置为 0），operator 写入 editedBy。
	RestoreById(ctx context.Context, groupID int64, operator string) error

	// UpdateName 更新分组名称。
	UpdateName(ctx context.Context, id int64, name string) error

	// UpdateNameByCondition 按条件批量更新分组名称。
	UpdateNameByCondition(ctx context.Context, artifactID int64, oldName, newName, format string) error

	// MoveLocation 调整分组归属：把指定 group 的 artifactLibID / parentID 指向新位置。
	// 跨 view 移动时调用方需另行同步子树资产的 artifactLibID（见 MoveArtifactLibByIDs）。
	MoveLocation(ctx context.Context, groupID, artifactID, parentID int64) error

	// MoveArtifactLibByIDs 批量把一批 group 的 artifactLibID 改到目标 view，
	// 用于跨 view 移动 group 时同步子树。
	MoveArtifactLibByIDs(ctx context.Context, groupIDs []int64, targetArtifactID int64) error

	// DeleteById 物理删除指定分组。
	DeleteById(ctx context.Context, groupID int64) error

	// ListRecycle 列出回收站中的分组（deleted=1），按 editedDate 倒序分页。
	ListRecycle(ctx context.Context, page, pageSize int) ([]*model.ArtifactGroup, int64, error)
}
