// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package gc

import (
	"time"

	model "github.com/easysoft/gitfox-community/app/model/artifact"

	"gorm.io/gorm"
)

// TagAssetOption selects container tag assets. In this schema, tag manifests are
// version-bound assets; layer/config blobs and digest-only manifests have versionID=0.
type TagAssetOption struct{}

func (TagAssetOption) Apply(db *gorm.DB) *gorm.DB {
	return db.
		Where("ops_artifact_assets.type = ?", model.ArtifactContainerFormat).
		Where("ops_artifact_assets.versionID IS NOT NULL").
		Where("ops_artifact_assets.versionID <> 0")
}

// ContainerFormatOption selects all container assets, including tag manifests,
// digest-only manifests, config blobs, and layer blobs.
type ContainerFormatOption struct{}

func (ContainerFormatOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.type = ?", model.ArtifactContainerFormat)
}

// TagActiveOption keeps only active tag assets for capacity statistics.
type TagActiveOption struct{}

func (TagActiveOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where("ops_artifact_assets.deleted = 0")
}

// TagUndeletedAndRecentDeletedOption keeps active tags plus recently soft-deleted
// tags as GC roots. Current schema stores the soft-delete marker in deleted=1
// and the deletion time in editedDate.
type TagUndeletedAndRecentDeletedOption struct {
	After time.Time
}

func (opt TagUndeletedAndRecentDeletedOption) Apply(db *gorm.DB) *gorm.DB {
	return db.Where(
		"(ops_artifact_assets.deleted = 0 OR (ops_artifact_assets.deleted > 0 AND ops_artifact_assets.editedDate > ?))",
		opt.After,
	)
}
