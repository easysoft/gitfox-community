package artifact

import (
	"context"
	"fmt"
	"strings"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gorm.io/gorm"
)

var _ VersionInterface = (*versions)(nil)

type versions struct{ db *gorm.DB }

func (c *versions) Create(ctx context.Context, newObj *model.ArtifactVersion) error {
	if err := c.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *versions) GetByID(ctx context.Context, versionId int64) (*model.ArtifactVersion, error) {
	var ver model.ArtifactVersion
	if err := c.db.WithContext(ctx).First(&ver, versionId).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &ver, nil
}

func (c *versions) GetByVersion(ctx context.Context, packageId, viewId int64, version string) (*model.ArtifactVersion, error) {
	var ver model.ArtifactVersion
	q := model.ArtifactVersion{PackageID: packageId, ViewID: viewId, Version: version}
	if err := c.db.WithContext(ctx).Where(q).First(&ver).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &ver, nil
}

func (c *versions) Update(ctx context.Context, ver *model.ArtifactVersion) error {
	if err := c.db.WithContext(ctx).Save(ver).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *versions) DeleteById(ctx context.Context, versionId int64) error {
	delNum, err := c.deleteByIds(ctx, versionId)
	if err != nil {
		return err
	}
	if delNum == 0 {
		return model.ErrPkgVersionNoItemDeleted
	}
	return nil
}

func (c *versions) DeleteByIds(ctx context.Context, versionIds ...int64) (int64, error) {
	return c.deleteByIds(ctx, versionIds...)
}

func (c *versions) deleteByIds(ctx context.Context, versionIds ...int64) (int64, error) {
	result := c.db.WithContext(ctx).Delete(&model.ArtifactVersion{}, versionIds)
	if result.Error != nil {
		return result.RowsAffected, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}

func (c *versions) Find(ctx context.Context, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error) {
	stmt := opt.Apply(c.db.WithContext(ctx))
	if opt.Query != "" {
		stmt = stmt.Where("version LIKE ?", fmt.Sprintf("%%%s%%", opt.Query))
	}
	if !opt.IncludeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	stmt = stmt.Limit(int(database.Limit(opt.Size))).Offset(int(database.Offset(opt.Page, opt.Size)))

	var data []*model.ArtifactVersion
	if err := stmt.Order("editedDate desc").Find(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return data, nil
}

func (c *versions) SoftDelete(ctx context.Context, ver *model.ArtifactVersion) error {
	ver.Deleted = 1
	return c.Update(ctx, ver)
}

func (c *versions) UnDelete(ctx context.Context, ver *model.ArtifactVersion) error {
	ver.Deleted = 0
	return c.Update(ctx, ver)
}

func (c *versions) Search(ctx context.Context, options ...SearchOption) ([]*model.ArtifactVersionInfo, error) {
	db := c.db.WithContext(ctx)
	for _, opt := range options {
		db = opt.Apply(db)
	}

	selectFields := []string{
		"ops_artifact_versions.id", "ops_artifact_versions.version", "ops_artifact_versions.viewID", "ops_artifact_versions.deleted",
		"ops_artifact_packages.id AS packageID", "ops_artifact_packages.name",
		"ops_artifact_packages.namespace", "ops_artifact_packages.format",
		"ops_artifact_views.spaceID", "ops_artifact_views.repoID",
		"ops_space.name AS space_name", "ops_repo.name AS repo_name",
	}

	var data []*model.ArtifactVersionInfo
	result := db.Model(&model.ArtifactVersion{}).
		Select(strings.Join(selectFields, ",")).
		Joins("left join ops_artifact_packages on ops_artifact_packages.id = ops_artifact_versions.packageID").
		Joins("left join ops_artifact_views on ops_artifact_views.id = ops_artifact_versions.viewID").
		Joins("left join ops_space on ops_space.id = ops_artifact_views.spaceID").
		Joins("left join ops_repo on ops_repo.id = ops_artifact_views.repoID").
		Scan(&data)

	if result.Error != nil {
		return nil, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return data, nil
}
