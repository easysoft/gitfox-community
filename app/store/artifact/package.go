package artifact

import (
	"context"
	"fmt"
	"time"

	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/lock"

	"gorm.io/gorm"
)

var _ PackageInterface = (*packages)(nil)

type packages struct{ db *gorm.DB }

var artifactPackageLockManager = lock.ProvideMutexManager()

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

// GetByNameInArtifactScope 查询 package 列表所在的 artifact scope 下是否存在指定 name 和 namespace 的 package
// 如果存在，返回该 package 的 model
// 如果不存在，返回 nil 和 nil 错误
// 如果存在多个 package，返回第一个 package 的 model
func (c *packages) GetByNameInArtifactScope(ctx context.Context, name, namespace string, artifactID int64, format model.ArtifactFormat) (*model.ArtifactPackage, error) {
	var row artifactPackage
	err := c.db.WithContext(ctx).
		Model(&artifactPackage{}).
		Joins("JOIN ops_artifact_libs package_lib ON package_lib.id = ops_artifact_packages.artifactLibID").
		Joins("JOIN ops_artifact_libs target_lib ON target_lib.id = ?", artifactID).
		Where("package_lib.spaceID = target_lib.spaceID").
		Where("package_lib.repoID = target_lib.repoID").
		Where("ops_artifact_packages.name = ?", name).
		Where("ops_artifact_packages.namespace = ?", namespace).
		Where("ops_artifact_packages.type = ?", format).
		Order("ops_artifact_packages.id ASC").
		First(&row).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return packageToModel(&row), nil
}

// LockArtifactScopePackage 锁定 artifact scope 下 package
func (c *packages) LockArtifactScopePackage(ctx context.Context, artifactID int64, namespace, name string, format model.ArtifactFormat) (func(), error) {
	lockKey, err := c.artifactScopePackageLockKey(ctx, artifactID, namespace, name, format)
	if err != nil {
		return nil, err
	}

	mx, err := artifactPackageLockManager.NewMutex(
		lockKey,
		lock.WithNamespace("artifact-package"),
		lock.WithExpiry(30*time.Second),
		lock.WithTries(20),
		lock.WithRetryDelay(100*time.Millisecond),
		lock.WithTimeoutFactor(1),
	)
	if err != nil {
		return nil, err
	}

	if err = mx.Lock(ctx); err != nil {
		return nil, err
	}

	return func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mx.Unlock(releaseCtx)
	}, nil
}

// artifactScopePackageLockKey 生成 artifact scope 下 package 锁的 key
// 锁的 key 格式为 space:spaceID/repo:repoID/format:format/namespace:namespace/name:name
func (c *packages) artifactScopePackageLockKey(ctx context.Context, artifactID int64, namespace, name string, format model.ArtifactFormat) (string, error) {
	var view artifactLib
	if err := c.db.WithContext(ctx).
		Select("spaceID", "repoID").
		Where("id = ?", artifactID).
		First(&view).Error; err != nil {
		return "", database.ProcessSQLErrorf(ctx, err)
	}

	return fmt.Sprintf("space:%d/repo:%d/format:%s/namespace:%s/name:%s", view.SpaceID, view.RepoID, format, namespace, name), nil
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
		stmt = stmt.Where("type = ?", format)
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

// SoftDelete 软删 package 并记录操作人到 editedBy（回收站列表 deletedBy 取此字段）。
// 仅更新 deleted/editedBy 两列，其它字段保持原值。
func (c *packages) SoftDelete(ctx context.Context, pkg *model.ArtifactPackage, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactPackage{ID: pkg.ID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
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

// MoveArtifactLibByGroupIDs 跨 view 移动 group 子树时，把挂在这些 group 上的 package 一并改归属。
func (c *packages) MoveArtifactLibByGroupIDs(ctx context.Context, groupIDs []int64, targetArtifactID int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	if err := c.db.WithContext(ctx).Model(&artifactPackage{}).
		Where("groupID IN ?", groupIDs).
		UpdateColumn("artifactLibID", targetArtifactID).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
