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

func (c *views) Create(ctx context.Context, newObj *model.ArtifactView) error {
	if err := c.db.WithContext(ctx).Create(newObj).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *views) Delete(ctx context.Context, id int64) error {
	if err := c.db.WithContext(ctx).Delete(&model.ArtifactView{}, id).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

func (c *views) GetDefault(ctx context.Context, spaceId int64) (*model.ArtifactView, error) {
	var view model.ArtifactView
	q := model.ArtifactView{SpaceID: spaceId, Default: true}
	if err := c.db.WithContext(ctx).Where(q).First(&view).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &view, nil
}

func (c *views) GetByName(ctx context.Context, spaceId, repoId int64, name string) (*model.ArtifactView, error) {
	var view model.ArtifactView
	q := model.ArtifactView{SpaceID: spaceId, RepoID: repoId, Name: name}
	if err := c.db.WithContext(ctx).Where(q).First(&view).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &view, nil
}

func (c *views) GetDefaultByRepoAndSpace(ctx context.Context, repoId, spaceId int64) (*model.ArtifactView, error) {
	var view model.ArtifactView
	q := model.ArtifactView{RepoID: repoId, SpaceID: spaceId, Default: true}
	if err := c.db.WithContext(ctx).Where(q).First(&view).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return &view, nil
}

// FindByNameInScope 在指定作用域下查询所有同名 view（包含已软删的）。
// 返回结果包含 active（deleted=0）和 soft-deleted（deleted=1）的 view，
// 由调用方根据 deleted 状态和 editedDate 判断是否允许创建同名 view。
func (c *views) FindByNameInScope(ctx context.Context, spaceId, repoId int64, name string) ([]*model.ArtifactView, error) {
	var list []*model.ArtifactView
	err := c.db.WithContext(ctx).
		Where("spaceID = ? AND repoID = ? AND name = ?", spaceId, repoId, name).
		Find(&list).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return list, nil
}

// SoftDelete 软删除视图（将 deleted 标记为 1）。
// GORM 的 autoUpdateTime 标签会自动将 editedDate 更新为当前时间，
// 因此 editedDate 复用为"删除时间"，用于回收站保留期的计算基准。
// 重要：软删后不应再更新该 view 的其他内容，以保证 editedDate 即删除时间。
func (c *views) SoftDelete(ctx context.Context, view *model.ArtifactView) error {
	view.Deleted = 1
	if err := c.db.WithContext(ctx).Save(view).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// HardDeleteExpired 物理删除所有 editedDate 早于 before 的已软删视图。
// 返回被删除的 view ID 列表，供上层级联清理关联的 packages、versions、assets、blobs 等资源。
func (c *views) HardDeleteExpired(ctx context.Context, before time.Time) ([]int64, error) {
	// 先查出待删除的 view ID，用于返回给调用方做级联清理
	var ids []int64
	err := c.db.WithContext(ctx).
		Model(&model.ArtifactView{}).
		Where("deleted = 1 AND editedDate < ?", before).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	if len(ids) == 0 {
		return nil, nil
	}

	// 物理删除这些过期的已软删 view
	if err = c.db.WithContext(ctx).
		Where("id IN ?", ids).
		Delete(&model.ArtifactView{}).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return ids, nil
}

