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

func (c *assets) Create(ctx context.Context, newObj *model.ArtifactAsset) error {
	// version/view/group 含可空字段，需先用代码补做查重，避免唯一约束在 NULL 语义下失效。
	if newObj.VersionID.Ptr() == nil {
		var err error
		if newObj.ViewID.Ptr() == nil {
			_, err = c.GetPath(ctx, newObj.Path, newObj.Format)
		} else if newObj.GroupID.Ptr() != nil {
			_, err = c.GetGroupAsset(ctx, newObj.Path, newObj.GroupID.Int64)
		} else {
			_, err = c.GetMetaAsset(ctx, newObj.Path, newObj.ViewID.Int64, newObj.Format)
		}
		if err == nil {
			return pkgErrors.ErrDuplicate
		}
	}
	if err := model.ValidateAssetKind(newObj.Kind); err != nil {
		return err
	}
	if err := c.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *assets) GetById(ctx context.Context, assetId int64) (*model.ArtifactAsset, error) {
	var asset model.ArtifactAsset
	if err := c.db.WithContext(ctx).First(&asset, assetId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &asset, nil
}

func (c *assets) find(ctx context.Context, query *model.ArtifactAsset) (*model.ArtifactAsset, error) {
	var asset model.ArtifactAsset
	if err := c.db.WithContext(ctx).Where(query).Where("deleted = 0").First(&asset).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &asset, nil
}

func (c *assets) GetPath(ctx context.Context, path string, format model.ArtifactFormat) (*model.ArtifactAsset, error) {
	return c.find(ctx, &model.ArtifactAsset{Path: path, Format: format})
}

func (c *assets) GetMetaAsset(ctx context.Context, path string, viewId int64, format model.ArtifactFormat) (*model.ArtifactAsset, error) {
	return c.find(ctx, &model.ArtifactAsset{Path: path, Format: format, ViewID: null.IntFrom(viewId)})
}

func (c *assets) GetVersionAsset(ctx context.Context, path string, versionId int64) (*model.ArtifactAsset, error) {
	return c.find(ctx, &model.ArtifactAsset{Path: path, VersionID: null.IntFrom(versionId)})
}

func (c *assets) GetGroupAsset(ctx context.Context, path string, groupId int64) (*model.ArtifactAsset, error) {
	return c.find(ctx, &model.ArtifactAsset{Path: path, GroupID: null.IntFrom(groupId)})
}

func (c *assets) FindMain(ctx context.Context, versionId int64) ([]*model.ArtifactAsset, error) {
	if err := validatorEmpty(versionId); err != nil {
		return nil, err
	}
	var assetList []*model.ArtifactAsset
	q := model.ArtifactAsset{VersionID: null.IntFrom(versionId)}
	if err := c.db.WithContext(ctx).Where(q).Where("deleted = 0").Find(&assetList).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return assetList, nil
}

func (c *assets) UpdateBlobId(ctx context.Context, assetId int64, blobId int64) error {
	if err := validatorEmpty(assetId); err != nil {
		return err
	}
	result := c.db.WithContext(ctx).Model(&model.ArtifactAsset{ID: assetId}).
		Updates(&model.ArtifactAsset{BlobID: blobId})
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return pkgErrors.ErrResourceNotFound
	}
	return nil
}

func (c *assets) Update(ctx context.Context, upObj *model.ArtifactAsset) error {
	if err := c.db.WithContext(ctx).Save(&upObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *assets) SoftDeleteById(ctx context.Context, assetId int64) error {
	err := c.db.WithContext(ctx).Model(&model.ArtifactAsset{ID: assetId}).
		UpdateColumn("deleted", 1).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *assets) SoftDeleteExcludeById(ctx context.Context, versionId int64, assetId int64) error {
	err := c.db.WithContext(ctx).Model(&model.ArtifactAsset{}).
		Where("versionID = ?", versionId).
		Where("id <> ?", assetId).
		Where("deleted = 0").
		UpdateColumn("deleted", 1).Error
	if err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *assets) DeleteById(ctx context.Context, assetId int64) error {
	result := c.db.WithContext(ctx).Delete(&model.ArtifactAsset{}, assetId)
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
	var data []*model.ArtifactAsset
	if err := db.Order("id asc").Find(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return data, nil
}

func (c *assets) SearchExtendBlob(ctx context.Context, options ...SearchOption) ([]*model.ArtifactAssetExtendBlob, error) {
	db := c.db.WithContext(ctx)
	for _, opt := range options {
		db = opt.Apply(db)
	}
	var data []*model.ArtifactAssetExtendBlob
	result := db.Model(&model.ArtifactAsset{}).
		Select("ops_artifact_assets.*, ops_artifact_blobs.storageID, ops_artifact_blobs.ref, ops_artifact_blobs.size").
		Joins("left join ops_artifact_blobs on ops_artifact_blobs.id = ops_artifact_assets.blobID").
		Order("id asc").Scan(&data)
	if result.Error != nil {
		return nil, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return data, nil
}

// DeleteByViewID 硬删指定 view 下的所有 asset 记录（含已软删的）。
// 注意：只删 asset 记录，不清理关联的 blob 和物理文件，需上层配合处理。
func (c *assets) DeleteByViewID(ctx context.Context, viewId int64) (int64, error) {
	result := c.db.WithContext(ctx).Where("viewID = ?", viewId).Delete(&model.ArtifactAsset{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
