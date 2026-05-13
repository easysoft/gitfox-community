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

func (c *versions) Create(ctx context.Context, newObj *model.CreateArtifactVersion) (*model.ArtifactVersion, error) {
	var row artifactVersion
	row.PackageID = newObj.PackageID
	row.Version = newObj.Version
	row.Metadata = newObj.Metadata
	row.CreatedBy = newObj.CreatedBy
	row.EditedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return versionToModel(&row), nil
}

func (c *versions) GetByID(ctx context.Context, versionID int64) (*model.ArtifactVersion, error) {
	var row artifactVersion
	if err := c.db.WithContext(ctx).First(&row, versionID).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return versionToModel(&row), nil
}

func (c *versions) GetByVersion(ctx context.Context, packageId int64, version string) (*model.ArtifactVersion, error) {
	var row artifactVersion
	q := artifactVersion{PackageID: packageId, Version: version}
	if err := c.db.WithContext(ctx).Where(&q).First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return versionToModel(&row), nil
}

func (c *versions) Update(ctx context.Context, ver *model.ArtifactVersion) error {
	row := versionFromModel(ver)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *versions) DeleteById(ctx context.Context, versionID int64) error {
	delNum, err := c.deleteByIds(ctx, versionID)
	if err != nil {
		return err
	}
	if delNum == 0 {
		return model.ErrPkgVersionNoItemDeleted
	}
	return nil
}

func (c *versions) DeleteByIds(ctx context.Context, versionIDs ...int64) (int64, error) {
	return c.deleteByIds(ctx, versionIDs...)
}

func (c *versions) deleteByIds(ctx context.Context, versionIDs ...int64) (int64, error) {
	result := c.db.WithContext(ctx).Delete(&artifactVersion{}, versionIDs)
	if result.Error != nil {
		return result.RowsAffected, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return result.RowsAffected, nil
}

func (c *versions) Find(ctx context.Context, opt model.SearchVersionOption) ([]*model.ArtifactVersion, error) {
	stmt := c.db.WithContext(ctx).Model(&artifactVersion{})
	// 应用 SearchVersionOption 的查询条件
	if opt.PackageID != 0 {
		stmt = stmt.Where("packageID = ?", opt.PackageID)
	}
	if opt.Query != "" {
		stmt = stmt.Where("version LIKE ?", fmt.Sprintf("%%%s%%", opt.Query))
	}
	if !opt.IncludeDeleted {
		stmt = stmt.Where("deleted = 0")
	}
	stmt = stmt.Limit(int(database.Limit(opt.Size))).Offset(int(database.Offset(opt.Page, opt.Size)))

	var rows []*artifactVersion
	if err := stmt.Order("editedDate desc").Find(&rows).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return versionsToModel(rows), nil
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
		"ops_artifact_versions.id", "ops_artifact_versions.version", "ops_artifact_versions.deleted",
		"ops_artifact_packages.id AS packageID", "ops_artifact_packages.name",
		"ops_artifact_packages.namespace", "ops_artifact_packages.format",
		"ops_artifact_packages.artifactID",
		"ops_artifact.spaceID", "ops_artifact.repoID",
		"ops_space.name AS space_name", "ops_repo.name AS repo_name",
	}

	var data []*model.ArtifactVersionInfo
	result := db.Model(&artifactVersion{}).
		Select(strings.Join(selectFields, ",")).
		Joins("left join ops_artifact_packages on ops_artifact_packages.id = ops_artifact_versions.packageID").
		Joins("left join ops_artifact on ops_artifact.id = ops_artifact_packages.artifactID").
		Joins("left join ops_space on ops_space.id = ops_artifact.spaceID").
		Joins("left join ops_repo on ops_repo.id = ops_artifact.repoID").
		Scan(&data)

	if result.Error != nil {
		return nil, database.ProcessSQLErrorf(ctx, result.Error)
	}
	return data, nil
}
