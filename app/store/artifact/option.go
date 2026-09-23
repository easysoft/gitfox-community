// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package artifact

import (
	"time"

	model "github.com/easysoft/gitfox-community/app/model/artifact"

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

// AssetWithDeletedBeforeOption 过滤「已软删 且 editedDate 早于 Before」的 asset。
type AssetWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt AssetWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.deleted = 1").Where("ops_artifact_assets.editedDate < ?", opt.Before)
}

type ContainerFormatOption struct{}

func (opt ContainerFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.type = ?", model.ArtifactContainerFormat)
}

type AssetFormatOption struct {
	Format string
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

// PackageWithDeletedBeforeOption 过滤「已软删 且 editedDate 早于 Before」的 package。
type PackageWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt PackageWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_packages.deleted = 1").
		Where("ops_artifact_packages.editedDate < ?", opt.Before)
}

type VersionWithDeletedOption struct {
	WithDeletedOption
}

// VersionWithDeletedBeforeOption 过滤「已软删 且 editedDate 早于 Before」的 version。
type VersionWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt VersionWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_versions.deleted = 1").
		Where("ops_artifact_versions.editedDate < ?", opt.Before)
}

// GroupWithDeletedBeforeOption 过滤「已软删 且 editedDate 早于 Before」的 group。
type GroupWithDeletedBeforeOption struct {
	Before time.Time
}

func (opt GroupWithDeletedBeforeOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_groups.deleted = 1").
		Where("ops_artifact_groups.editedDate < ?", opt.Before)
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
