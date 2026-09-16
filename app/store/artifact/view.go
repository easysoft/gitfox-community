package artifact

import (
	"context"

	model "github.com/easysoft/gitfox-community/app/model/artifact"
	"github.com/easysoft/gitfox-community/pkg/database"

	"gorm.io/gorm"
)

var _ ViewInterface = (*libs)(nil)

type libs struct{ db *gorm.DB }

// Create 创建视图
func (c *libs) Create(ctx context.Context, newObj *model.CreateArtifactView) (*model.ArtifactView, error) {
	var view artifactLib
	view.SpaceID = newObj.SpaceID
	view.RepoID = newObj.RepoID
	view.Type = string(newObj.Type)
	view.Name = newObj.Name
	view.Code = newObj.Code
	view.Format = string(newObj.Format)
	view.Description = newObj.Description
	view.EditedBy = newObj.EditedBy
	view.CreatedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&view).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return viewToModel(&view), nil
}

// SoftDelete 软删除视图（将 deleted 标记为 1）并记录操作人到 editedBy（即删除人）。
func (c *libs) SoftDelete(ctx context.Context, view *model.ArtifactView, operator string) error {
	updates := map[string]interface{}{"deleted": 1}
	if operator != "" {
		updates["editedBy"] = operator
	}
	if err := c.db.WithContext(ctx).Model(&artifactLib{ID: view.ID}).
		Updates(updates).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateName 更新视图名称。
func (c *libs) UpdateName(ctx context.Context, id int64, name string) error {
	if err := c.db.WithContext(ctx).Model(&artifactLib{}).Where("id = ?", id).
		UpdateColumn("name", name).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// GetByID 按 ID 取视图，不过滤 deleted 标记，用于回收站恢复路径。
func (c *libs) GetByID(ctx context.Context, id int64) (*model.ArtifactView, error) {
	var row artifactLib
	if err := c.db.WithContext(ctx).Where("deleted = 0").Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// GetByIDWithDeleted 按 ID 取视图，不过滤 deleted 标记，用于回收站恢复路径。
func (c *libs) GetByIDWithDeleted(ctx context.Context, id int64) (*model.ArtifactView, error) {
	var row artifactLib
	if err := c.db.WithContext(ctx).Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// GetByCode 按 (spaceID, repoID, code) 取视图。
func (c *libs) GetByCode(ctx context.Context, spaceID, repoID int64, code string) (*model.ArtifactView, error) {
	var row artifactLib
	q := artifactLib{SpaceID: spaceID, RepoID: repoID, Code: code}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// GetByName 按名称查询视图 未包含已软删的
func (c *libs) GetByName(ctx context.Context, spaceID, repoID int64, name string) (*model.ArtifactView, error) {
	var row artifactLib
	q := artifactLib{SpaceID: spaceID, RepoID: repoID, Name: name}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// FindByNameInScope 在指定作用域下查询所有同名 view
// 返回结果包含 active（deleted=0）和 soft-deleted（deleted=1）的 view，
// 由调用方根据 deleted 状态和 editedDate 判断是否允许创建同名 view。
func (c *libs) FindByNameInScope(ctx context.Context, spaceID, repoID int64, name string) ([]*model.ArtifactView, error) {
	var rows []*artifactLib
	err := c.db.WithContext(ctx).
		Where("spaceID = ? AND repoID = ? AND name = ? and deleted = 0", spaceID, repoID, name).
		Find(&rows).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	result := make([]*model.ArtifactView, 0, len(rows))
	for _, r := range rows {
		result = append(result, viewToModel(r))
	}
	return result, nil
}
