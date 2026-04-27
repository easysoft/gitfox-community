package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gorm.io/gorm"
)

var _ PackageInterface = (*packages)(nil)

type packages struct{ db *gorm.DB }

func (c *packages) Create(ctx context.Context, newObj *model.ArtifactPackage) error {
	if err := c.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *packages) GetByID(ctx context.Context, packageId int64) (*model.ArtifactPackage, error) {
	var pkg model.ArtifactPackage
	if err := c.db.WithContext(ctx).First(&pkg, packageId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &pkg, nil
}

func (c *packages) GetByName(ctx context.Context, name, namespace string, viewId int64, format model.ArtifactFormat) (*model.ArtifactPackage, error) {
	var pkg model.ArtifactPackage
	q := model.ArtifactPackage{Name: name, Namespace: namespace, ViewID: viewId, Format: format}
	if err := c.db.WithContext(ctx).Where(q).Where("namespace = ?", namespace).First(&pkg).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &pkg, nil
}

func (c *packages) List(ctx context.Context, viewId int64, includeDeleted bool) ([]*model.ArtifactPackage, error) {
	q := model.ArtifactPackage{ViewID: viewId}
	stmt := c.db.WithContext(ctx).Where(q)
	if !includeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	var dst []*model.ArtifactPackage
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}

func (c *packages) ListBatch(ctx context.Context, includeDeleted bool, batchSize int, fn func(pkg *model.ArtifactPackage) error) error {
	var results []*model.ArtifactPackage
	stmt := c.db.WithContext(ctx)
	if !includeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	res := stmt.FindInBatches(&results, batchSize, func(tx *gorm.DB, batch int) error {
		for _, result := range results {
			if err := fn(result); err != nil {
				return err
			}
		}
		return nil
	})
	if res.Error != nil {
		return database.ProcessSQLErrorf(ctx, res.Error)
	}
	return nil
}

func (c *packages) ListByNamespace(ctx context.Context, viewId int64, namespace string, format model.ArtifactFormat, recurse, includeDeleted bool) ([]*model.ArtifactPackage, error) {
	q := model.ArtifactPackage{ViewID: viewId}
	stmt := c.db.WithContext(ctx).Where(q)
	if recurse {
		stmt = stmt.Where("namespace LIKE ?", namespace+"%")
	} else {
		stmt = stmt.Where("namespace = ?", namespace)
	}
	if format != model.ArtifactAllFormat {
		stmt = stmt.Where("format = ?", format)
	}
	if !includeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	var dst []*model.ArtifactPackage
	if err := stmt.Find(&dst).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return dst, nil
}

func (c *packages) Update(ctx context.Context, upObj *model.ArtifactPackage) error {
	if err := c.db.WithContext(ctx).Save(&upObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *packages) SoftDelete(ctx context.Context, pkg *model.ArtifactPackage) error {
	pkg.Deleted = 1
	return c.Update(ctx, pkg)
}

func (c *packages) UnDelete(ctx context.Context, pkg *model.ArtifactPackage) error {
	pkg.Deleted = 0
	return c.Update(ctx, pkg)
}

func (c *packages) DeleteById(ctx context.Context, packageId int64) error {
	delNum, err := c.deleteByIds(ctx, packageId)
	if err != nil {
		return err
	}
	if delNum == 0 {
		return model.ErrPkgNoItemDeleted
	}
	return nil
}

func (c *packages) DeleteByIds(ctx context.Context, packageIds ...int64) (int64, error) {
	return c.deleteByIds(ctx, packageIds...)
}

func (c *packages) deleteByIds(ctx context.Context, packageIds ...int64) (int64, error) {
	result := c.db.WithContext(ctx).Delete(&model.ArtifactPackage{}, packageIds)
	if result.Error != nil {
		return result.RowsAffected, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}

func (c *packages) Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactPackage, error) {
	db := c.db.WithContext(ctx)
	for _, opt := range options {
		db = opt.Apply(db)
	}
	var data []*model.ArtifactPackage
	if err := db.Order("id asc").Find(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return data, nil
}

func (c *packages) GetByAssetID(ctx context.Context, assetId int64) (*model.ArtifactPackage, error) {
	var pkg model.ArtifactPackage
	db := c.db.WithContext(ctx).Model(&model.ArtifactPackage{}).
		Select("ops_artifact_packages.*").
		Joins("JOIN ops_artifact_versions ON ops_artifact_packages.id = ops_artifact_versions.packageID").
		Joins("JOIN ops_artifact_assets ON ops_artifact_versions.id = ops_artifact_assets.versionID").
		Where("ops_artifact_assets.id = ?", assetId).
		Where("ops_artifact_assets.deleted = 0")
	if err := db.First(&pkg).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &pkg, nil
}
