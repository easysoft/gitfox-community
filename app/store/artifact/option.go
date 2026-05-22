package artifact

import (
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"

	"gorm.io/gorm"
)

type WithDeletedOption struct{}

func (opt WithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type ExcludeDeletedOption struct{}

func (opt ExcludeDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 0")
}

type AssetVersionedOption struct{}

func (opt AssetVersionedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("versionID IS NOT NULL").Where("versionID <> 0")
}

type AssetWithDeletedOption struct {
	WithDeletedOption
}

type AssetExcludeDeletedOption struct {
	ExcludeDeletedOption
}

// GroupAssetsOption 按 groupID 查询 assets。
type GroupAssetsOption struct {
	GroupID int64
}

func (opt GroupAssetsOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("groupID = ?", opt.GroupID)
}

type AssetVersionOption struct {
	VersionID int64
}

func (opt AssetVersionOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("versionID = ?", opt.VersionID)
}

type AssetWithDeletedBeforeOption struct {
	WithDeletedOption
	Before time.Time
}

type ContainerFormatOption struct{}

func (opt ContainerFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.type = ?", model.ArtifactContainerFormat)
}

type AssetFormatOption struct {
	Format model.ArtifactFormat
}

func (opt AssetFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.type = ?", opt.Format)
}

type ExcludeContainerOption struct{}

func (opt ExcludeContainerOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.type <> ?", model.ArtifactContainerFormat)
}

type VersionIDsOption struct {
	IDs []int64
}

func (opt VersionIDsOption) Apply(db *gorm.DB) *gorm.DB {
	if len(opt.IDs) == 0 {
		return db.Where("1 = 0")
	}
	return db.Where("ops_artifact_versions.id IN ?", opt.IDs).Order("ops_artifact_versions.id asc")
}

type PackageWithDeletedOption struct {
	WithDeletedOption
}

// GroupPackagesOption 按 groupID 查询未删除 package。
type GroupPackagesOption struct {
	GroupID int64
}

func (opt GroupPackagesOption) Apply(db *gorm.DB) *gorm.DB {
	return ExcludeDeletedOption{}.Apply(db.Where("groupID = ?", opt.GroupID))
}

type PackageGroupOption struct {
	GroupID int64
}

func (opt PackageGroupOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("groupID = ?", opt.GroupID)
}

type VersionPackageOption struct {
	PackageID int64
}

func (opt VersionPackageOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("packageID = ?", opt.PackageID)
}

type PackageWithDeletedBeforeOption struct {
	WithDeletedOption
	Before time.Time
}

type VersionWithDeletedOption struct {
	WithDeletedOption
}

type VersionWithDeletedBeforeOption struct {
	WithDeletedOption
	Before time.Time
}

// AssetViewOption 按 artifactID 过滤 assets。
type AssetViewOption struct {
	ArtifactID int64
}

func (opt AssetViewOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.artifactLibID = ?", opt.ArtifactID)
}

// BlobAssetsOption 按 blobID 查询 assets。
type BlobAssetsOption struct {
	BlobID int64
}

func (opt BlobAssetsOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("blobID = ?", opt.BlobID)
}
