package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gorm.io/gorm"
)

var _ MetaAssetInterface = (*metaAssets)(nil)

type metaAssets struct{ db *gorm.DB }

func (s *metaAssets) Create(ctx context.Context, newObj *model.ArtifactMetaAsset) error {
	if err := validatorEmpty(newObj.ViewID, newObj.Path, newObj.ViewID, newObj.Format); err != nil {
		return err
	}
	if err := model.ValidateAssetKind(newObj.Kind); err != nil {
		return err
	}
	if err := s.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *metaAssets) GetById(ctx context.Context, metaAssetId int64) (*model.ArtifactMetaAsset, error) {
	var metaAsset model.ArtifactMetaAsset
	if err := s.db.WithContext(ctx).First(&metaAsset, metaAssetId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &metaAsset, nil
}

func (s *metaAssets) GetByPath(ctx context.Context, path string, viewId int64, format model.ArtifactFormat) (*model.ArtifactMetaAsset, error) {
	if err := validatorEmpty(path, viewId, format); err != nil {
		return nil, err
	}
	var metaAsset model.ArtifactMetaAsset
	q := model.ArtifactMetaAsset{Path: path, ViewID: viewId, Format: format}
	if err := s.db.WithContext(ctx).Where(q).First(&metaAsset).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &metaAsset, nil
}

func (s *metaAssets) Update(ctx context.Context, metaAsset *model.ArtifactMetaAsset) error {
	if err := s.db.WithContext(ctx).Save(metaAsset).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (s *metaAssets) DeleteById(ctx context.Context, metaAssetId int64) error {
	result := s.db.WithContext(ctx).Delete(&model.ArtifactMetaAsset{}, metaAssetId)
	if result.Error != nil {
		return database.ProcessSQLErrorf(ctx, result.Error)
	}
	if result.RowsAffected == 0 {
		return model.ErrMetaAssetNoItemDeleted
	}
	return nil
}

// DeleteByViewID 硬删指定 view 下的所有 meta_asset 记录。
func (s *metaAssets) DeleteByViewID(ctx context.Context, viewId int64) (int64, error) {
	result := s.db.WithContext(ctx).Where("viewID = ?", viewId).Delete(&model.ArtifactMetaAsset{})
	if result.Error != nil {
		return 0, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}
