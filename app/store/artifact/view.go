package artifact

import (
	"context"

	model "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/artifact"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"

	"gorm.io/gorm"
)

var _ ViewInterface = (*views)(nil)

type views struct{ db *gorm.DB }

func (c *views) Create(ctx context.Context, newObj *model.CreateArtifactView) (*model.ArtifactView, error) {
	var view artifactView
	view.SpaceID = newObj.SpaceID
	view.RepoID = newObj.RepoID
	view.Type = string(newObj.Type)
	view.Default = newObj.Default
	view.Name = newObj.Name
	view.Format = string(newObj.Format)
	view.Description = newObj.Description
	view.EditedBy = newObj.EditedBy
	view.CreatedBy = newObj.CreatedBy
	if err := c.db.WithContext(ctx).Create(&view).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return viewToModel(&view), nil
}

func (c *views) Delete(ctx context.Context, id int64) error {
	if err := c.db.WithContext(ctx).Delete(&artifactView{}, id).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *views) GetByID(ctx context.Context, id int64) (*model.ArtifactView, error) {
	var row artifactView
	if err := c.db.WithContext(ctx).Where("deleted = 0").Take(&row, id).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

func (c *views) GetDefault(ctx context.Context, spaceID int64) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{SpaceID: spaceID, Default: true}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

func (c *views) GetByName(ctx context.Context, spaceID, repoID int64, name string) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{SpaceID: spaceID, RepoID: repoID, Name: name}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

func (c *views) GetDefaultByRepoAndSpace(ctx context.Context, repoID, spaceID int64) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{RepoID: repoID, SpaceID: spaceID, Default: true}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// FindByNameInScope 在指定作用域下查询所有同名 view（包含已软删的）。
// 返回结果包含 active（deleted=0）和 soft-deleted（deleted=1）的 view，
// 由调用方根据 deleted 状态和 editedDate 判断是否允许创建同名 view。
func (c *views) FindByNameInScope(ctx context.Context, spaceID, repoID int64, name string) ([]*model.ArtifactView, error) {
	var rows []*artifactView
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

// SoftDelete 软删除视图（将 deleted 标记为 1）。
// BeforeUpdate 钩子会自动更新 editedDate，复用为"删除时间"，用于回收站保留期的计算基准。
// 重要：软删后不应再更新该 view 的其他内容，以保证 editedDate 即删除时间。
func (c *views) SoftDelete(ctx context.Context, view *model.ArtifactView) error {
	view.Deleted = 1
	row := viewFromModel(view)
	if err := c.db.WithContext(ctx).Save(row).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// UpdateName 更新视图名称。
func (c *views) UpdateName(ctx context.Context, id int64, name string) error {
	if err := c.db.WithContext(ctx).Model(&artifactView{}).Where("id = ?", id).
		UpdateColumn("name", name).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}
