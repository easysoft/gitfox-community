package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"github.com/guregu/null"
	"gorm.io/gorm"
)

var _ AssetInterface = (*assets)(nil)

type assets struct{ db *gorm.DB }

func (c *assets) Create(ctx context.Context, newObj *model.CreateArtifactAsset) (*model.ArtifactAsset, error) {
	// version/view/group 含可空字段，需先用代码补做查重，避免唯一约束在 NULL 语义下失效。
	if newObj.VersionID.Ptr() == nil {
		var err error
		if newObj.ArtifactID.Ptr() == nil {
			_, err = c.GetPath(ctx, newObj.Path, newObj.Format)
		} else if newObj.GroupID.Ptr() != nil {
			_, err = c.GetGroupAsset(ctx, newObj.Path, newObj.GroupID.Int64)
		} else {
			_, err = c.GetMetaAsset(ctx, newObj.Path, newObj.ArtifactID.Int64, newObj.Format)
		}
		if err == nil {
			return nil, pkgErrors.ErrDuplicate
		}
	}
	if err := model.ValidateAssetKind(newObj.Kind); err != nil {
		return nil, err
	}
	var row artifactAsset
	// DDL 中 versionID / artifactID / groupID 均为 NOT NULL DEFAULT 0，
	// null.Int 在 valid=false 时会写出字面 NULL，触发 MySQL not null 约束。
	// 显式回填为 0，保持与默认值一致，避免插入失败。
	row.VersionID = nullIntOrZero(newObj.VersionID)
	row.ArtifactID = nullIntOrZero(newObj.ArtifactID)
	row.GroupID = nullIntOrZero(newObj.GroupID)
	row.Format = string(newObj.Format)
	row.Path = newObj.Path
	row.ContentType = newObj.ContentType
	row.Kind = string(newObj.Kind)
	row.BlobID = newObj.BlobID
	row.CheckSum = newObj.CheckSum
	row.EditedBy = newObj.EditedBy
	row.CreatedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetToModel(&row), nil
}

// nullIntOrZero 将 null.Int 在 invalid 时归一化为 IntFrom(0)，
// 用于 NOT NULL DEFAULT 0 的列，避免 GORM 写出 NULL 触发约束。
func nullIntOrZero(v null.Int) null.Int {
	if !v.Valid {
		return null.IntFrom(0)
	}
	return v
}

func (c *assets) GetById(ctx context.Context, assetID int64) (*model.ArtifactAsset, error) {
	var row artifactAsset
	if err := c.db.WithContext(ctx).First(&row, assetID).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetToModel(&row), nil
}

func (c *assets) find(ctx context.Context, query *artifactAsset) (*model.ArtifactAsset, error) {
	var row artifactAsset
	if err := c.db.WithContext(ctx).Where(query).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetToModel(&row), nil
}

func (c *assets) GetPath(ctx context.Context, path string, format model.ArtifactFormat) (*model.ArtifactAsset, error) {
	return c.find(ctx, &artifactAsset{Path: path, Format: string(format)})
}

func (c *assets) GetMetaAsset(ctx context.Context, path string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactAsset, error) {
	return c.find(ctx, &artifactAsset{Path: path, Format: string(format), ArtifactID: null.IntFrom(artifactID)})
}

func (c *assets) GetVersionAsset(ctx context.Context, path string, versionID int64) (*model.ArtifactAsset, error) {
	return c.find(ctx, &artifactAsset{Path: path, VersionID: null.IntFrom(versionID)})
}

func (c *assets) GetGroupAsset(ctx context.Context, path string, groupID int64) (*model.ArtifactAsset, error) {
	return c.find(ctx, &artifactAsset{Path: path, GroupID: null.IntFrom(groupID)})
}

func (c *assets) FindMain(ctx context.Context, versionID int64) ([]*model.ArtifactAsset, error) {
	if err := validatorEmpty(versionID); err != nil {
		return nil, err
	}
	var rows []*artifactAsset
	q := artifactAsset{VersionID: null.IntFrom(versionID)}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetsToModel(rows), nil
}

func (c *assets) UpdateBlobId(ctx context.Context, assetID int64, blobID int64) error {
	if err := validatorEmpty(assetID); err != nil {
		return err
	}
	result := c.db.WithContext(ctx).Model(&artifactAsset{ID: assetID}).
		Updates(&artifactAsset{BlobID: blobID})
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgErrors.ErrResourceNotFound
	}
	return nil
}

func (c *assets) Update(ctx context.Context, upObj *model.ArtifactAsset) error {
	row := assetFromModel(upObj)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// SoftDeleteById 软删除 asset 并记录操作人到 editedBy（即"删除人"，回收站列表 deletedBy 取此字段）。
func (c *assets) SoftDeleteById(ctx context.Context, assetID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	err := c.db.WithContext(ctx).Model(&artifactAsset{ID: assetID}).
		Updates(updates).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// SoftDeleteExcludeById 软删指定 version 下除排除 ID 外的所有活资产，并记录操作人。
func (c *assets) SoftDeleteExcludeById(ctx context.Context, versionID int64, assetID int64, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	err := c.db.WithContext(ctx).Model(&artifactAsset{}).
		Where("versionID = ?", versionID).
		Where("id <> ?", assetID).
		Where("deleted = 0").
		Updates(updates).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *assets) DeleteById(ctx context.Context, assetID int64) error {
	result := c.db.WithContext(ctx).Delete(&artifactAsset{}, assetID)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrAssetNoItemDeleted
	}
	return nil
}

func (c *assets) Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAsset, error) {
	db := c.db.WithContext(ctx)
	for _, opt := range options {
		db = opt.Apply(db)
	}
	var rows []*artifactAsset
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetsToModel(rows), nil
}

func (c *assets) SearchExtendBlob(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAssetExtendBlob, error) {
	db := c.db.WithContext(ctx)
	for _, opt := range options {
		db = opt.Apply(db)
	}
	// 用内部 row 承接 join 结果：列 rank 已重命名，必须由 store 控制 gorm tag，
	// 否则上层 model.ArtifactAsset.Kind 默认走蛇形 column "kind"，扫不到值。
	type assetExtendRow struct {
		artifactAsset
		StorageID int64  `gorm:"column:storageID"`
		Ref       string `gorm:"column:ref"`
		Size      int64  `gorm:"column:size"`
	}
	var rows []*assetExtendRow
	result := db.Model(&artifactAsset{}).
		Select("ops_artifact_assets.*, ops_artifact_blobs.storageID, ops_artifact_blobs.ref, ops_artifact_blobs.size").
		Joins("left join ops_artifact_blobs on ops_artifact_blobs.id = ops_artifact_assets.blobID").
		Order("id asc").Scan(&rows)
	if result.Error != nil {
		return nil, database.ProcessSQLErrorf(ctx, result.Error)
	}
	data := make([]*model.ArtifactAssetExtendBlob, 0, len(rows))
	for _, r := range rows {
		base := assetToModel(&r.artifactAsset)
		data = append(data, &model.ArtifactAssetExtendBlob{
			ArtifactAsset: *base,
			BlobID:        r.BlobID,
			StorageID:     r.StorageID,
			Ref:           r.Ref,
			Size:          r.Size,
		})
	}
	return data, nil
}

// DeleteByArtifactID 硬删指定 view 下的所有 asset 记录（含已软删的）。
// 注意：只删 asset 记录，不清理关联的 blob 和物理文件，需上层配合处理。
func (c *assets) DeleteByArtifactID(ctx context.Context, artifactID int64) (int64, error) {
	result := c.db.WithContext(ctx).Where("artifactLibID = ?", artifactID).Delete(&artifactAsset{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
