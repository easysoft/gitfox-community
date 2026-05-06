package artifact

import (
	"context"
	"time"

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
	view.Default = newObj.Default
	view.Name = newObj.Name
	view.Description = newObj.Description
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

func (c *views) GetDefault(ctx context.Context, spaceId int64) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{SpaceID: spaceId, Default: true}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

func (c *views) GetByName(ctx context.Context, spaceId, repoId int64, name string) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{SpaceID: spaceId, RepoID: repoId, Name: name}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

func (c *views) GetDefaultByRepoAndSpace(ctx context.Context, repoId, spaceId int64) (*model.ArtifactView, error) {
	var row artifactView
	q := artifactView{RepoID: repoId, SpaceID: spaceId, Default: true}
	if err := c.db.WithContext(ctx).Where(&q).Where("deleted = 0").First(&row).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return viewToModel(&row), nil
}

// FindByNameInScope 在指定作用域下查询所有同名 view（包含已软删的）。
// 返回结果包含 active（deleted=0）和 soft-deleted（deleted=1）的 view，
// 由调用方根据 deleted 状态和 editedDate 判断是否允许创建同名 view。
func (c *views) FindByNameInScope(ctx context.Context, spaceId, repoId int64, name string) ([]*model.ArtifactView, error) {
	var rows []*artifactView
	err := c.db.WithContext(ctx).
		Where("spaceID = ? AND repoID = ? AND name = ?", spaceId, repoId, name).
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

// FindExpiredIDs 查询所有 editedDate 早于 before 的已软删视图 ID 列表。
// 只做查询，不执行删除操作，供上层在级联清理关联资源后再逐条删除 view 记录。
func (c *views) FindExpiredIDs(ctx context.Context, before time.Time) ([]int64, error) {
	var ids []int64
	err := c.db.WithContext(ctx).
		Model(&artifactView{}).
		Where("deleted = 1 AND editedDate < ?", before).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return ids, nil
}

// HardDeleteExpired 物理删除所有 editedDate 早于 before 的已软删视图。
// 返回被删除的 view ID 列表，供上层级联清理关联的 packages、versions、assets、blobs 等资源。
//
// Deprecated: 直接使用此方法会导致先删 view 再清理关联资源，崩溃时产生孤儿资源。
// 请改用 FindExpiredIDs + hardDeleteViewResources + DeleteByID 的三步流程。
func (c *views) HardDeleteExpired(ctx context.Context, before time.Time) ([]int64, error) {
	ids, err := c.FindExpiredIDs(ctx, before)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, nil
	}

	// 物理删除这些过期的已软删 view
	if err = c.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&artifactView{}).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return ids, nil
}
