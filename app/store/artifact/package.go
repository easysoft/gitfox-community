package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

	"gorm.io/gorm"
)

var _ PackageInterface = (*packages)(nil)

type packages struct{ db *gorm.DB }

func (c *packages) Create(ctx context.Context, newObj *model.CreateArtifactPackage) (*model.ArtifactPackage, error) {
	var row artifactPackage
	row.ArtifactID = newObj.ArtifactID
	row.Name = newObj.Name
	row.Namespace = newObj.Namespace
	row.GroupID = newObj.GroupID
	row.Format = string(newObj.Format)
	row.CreatedBy = newObj.CreatedBy
	row.EditedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packageToModel(&row), nil
}

func (c *packages) GetByID(ctx context.Context, packageId int64) (*model.ArtifactPackage, error) {
	var row artifactPackage
	if err := c.db.WithContext(ctx).First(&row, packageId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packageToModel(&row), nil
}

func (c *packages) GetByName(ctx context.Context, name, namespace string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactPackage, error) {
	var row artifactPackage
	q := artifactPackage{Name: name, Namespace: namespace, ArtifactID: artifactID, Format: string(format)}
	if err := c.db.WithContext(ctx).Where(&q).Where("namespace = ?", namespace).First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packageToModel(&row), nil
}

func (c *packages) List(ctx context.Context, artifactID int64, includeDeleted bool) ([]*model.ArtifactPackage, error) {
	q := artifactPackage{ArtifactID: artifactID}
	stmt := c.db.WithContext(ctx).Where(&q)
	if !includeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	var rows []*artifactPackage
	if err := stmt.Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packagesToModel(rows), nil
}

func (c *packages) ListBatch(ctx context.Context, includeDeleted bool, batchSize int, fn func(pkg *model.ArtifactPackage) error) error {
	var rows []*artifactPackage
	stmt := c.db.WithContext(ctx)
	if !includeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	res := stmt.FindInBatches(&rows, batchSize, func(tx *gorm.DB, batch int) error {
		for _, row := range rows {
			if err := fn(packageToModel(row)); err != nil {
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

func (c *packages) ListByNamespace(ctx context.Context, artifactID int64, namespace string, format model.ArtifactFormat, recurse, includeDeleted bool) ([]*model.ArtifactPackage, error) {
	q := artifactPackage{ArtifactID: artifactID}
	stmt := c.db.WithContext(ctx).Where(&q)
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
	var rows []*artifactPackage
	if err := stmt.Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packagesToModel(rows), nil
}

func (c *packages) Update(ctx context.Context, upObj *model.ArtifactPackage) error {
	row := packageFromModel(upObj)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
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
	result := c.db.WithContext(ctx).Delete(&artifactPackage{}, packageIds)
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
	var rows []*artifactPackage
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packagesToModel(rows), nil
}

func (c *packages) GetByAssetID(ctx context.Context, assetID int64) (*model.ArtifactPackage, error) {
	var row artifactPackage
	db := c.db.WithContext(ctx).Model(&artifactPackage{}).
		Select("ops_artifact_packages.*").
		Joins("JOIN ops_artifact_versions ON ops_artifact_packages.id = ops_artifact_versions.packageID").
		Joins("JOIN ops_artifact_assets ON ops_artifact_versions.id = ops_artifact_assets.versionID").
		Where("ops_artifact_assets.id = ?", assetID).
		Where("ops_artifact_assets.deleted = 0")
	if err := db.First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packageToModel(&row), nil
}

// 仅用于 Search 的 SearchOption 中，避免循环引用
var _ = pkgErrors.ErrResourceNotFound

// UpdateNamespaceByCondition 按条件批量更新 package namespace（用于 container 制品库重命名时同步）。
func (c *packages) UpdateNamespaceByCondition(ctx context.Context, artifactID int64, oldNamespace, newNamespace, format string) error {
	if err := c.db.WithContext(ctx).Model(&artifactPackage{}).
		Where("artifactID = ? AND namespace = ? AND format = ?", artifactID, oldNamespace, format).
		UpdateColumn("namespace", newNamespace).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
