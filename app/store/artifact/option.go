package artifact

import (
	"time"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"

	"gorm.io/gorm"
)

type AssetVersionedOption struct{}

func (opt AssetVersionedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("versionID IS NOT NULL").Where("versionID <> 0")
}

type AssetWithDeletedOption struct{}

func (opt AssetWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type AssetExcludeDeletedOption struct{}

func (opt AssetExcludeDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 0")
}

type AssetWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt AssetWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type ContainerFormatOption struct{}

func (opt ContainerFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.format = ?", model.ArtifactContainerFormat)
}

type ExcludeContainerOption struct{}

func (opt ExcludeContainerOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("versionID IS NOT NULL").Where("versionID <> 0").
		Where("ops_artifact_assets.format <> ?", model.ArtifactContainerFormat)
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

type PackageWithDeletedOption struct{}

func (opt PackageWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type PackageWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt PackageWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type VersionWithDeletedOption struct{}

func (opt VersionWithDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

type VersionWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt VersionWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("deleted = 1")
}

// AssetViewOption 按 artifactID 过滤 assets。
type AssetViewOption struct {
	ArtifactID int64
}

func (opt AssetViewOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.artifactID = ?", opt.ArtifactID)
}
