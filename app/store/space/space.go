package space

import (
	"context"
	"fmt"

	spaceModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/space"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

type SpaceStore struct {
	db *gorm.DB
}

func NewSpaceStore(db *gorm.DB) *SpaceStore {
	return &SpaceStore{db: db}
}

func (s *SpaceStore) Create(ctx context.Context, req spaceModel.CreateSpaceRequest) (*spaceModel.SpaceResponse, error) {

	data := Space{
		Name:       req.Name,
		Owner:      req.Owner,
		Permission: req.Permission,
		Desc:       req.Desc,
		CreatedBy:  req.CreatedBy,
		EditedBy:   req.CreatedBy,
	}

	if err := s.db.WithContext(ctx).Create(&data).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapToModel(data), nil
}

func (s *SpaceStore) Update(ctx context.Context, id uint, req spaceModel.UpdateSpaceRequest) (*spaceModel.SpaceResponse, error) {
	var out Space
	out.ID = id

	updateData := make(map[string]any)

	// EditedBy 和 EditedDate 是必填字段
	updateData["editedBy"] = req.EditedBy

	// 只更新非空字段
	if req.Name != "" {
		updateData["name"] = req.Name
	}
	if req.Owner != "" {
		updateData["owner"] = req.Owner
	}
	if req.Permission != "" {
		updateData["acl"] = req.Permission
	}
	if req.Desc != "" {
		updateData["desc"] = req.Desc
	}

	// 使用事务确保查询和更新的原子性
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 先查询出要更新的记录
		if err := tx.First(&out).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 提取 map 中的所有字段名,用于 Select
		fields := make([]string, 0, len(updateData))
		for field := range updateData {
			fields = append(fields, field)
		}

		// 使用 Select 明确指定要更新的字段,支持零值更新
		// Updates 会直接更新 out 对象（传入指针）
		if err := tx.Model(&out).
			Select(fields).
			Updates(updateData).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// 将 Space 转换为 SpaceResponse
	return mapToModel(out), nil
}

func (s *SpaceStore) List(ctx context.Context, filter spaceModel.ListSpaceFilter) (*database.PageResult[*Space], error) {
	tx := s.db.WithContext(ctx).Model(&Space{}).Where("deleted = ?", 0)

	if filter.Name != "" {
		// 用LOWER函数实现不区分大小写匹配，避免数据库大小写敏感问题
		tx = tx.Where("LOWER(name) LIKE LOWER(?)", "%"+filter.Name+"%")
	}
	if filter.Permission != "" {
		tx = tx.Where("acl = ?", filter.Permission)
	}
	if filter.Owner != "" {
		tx = tx.Where("owner = ?", filter.Owner)
	}

	// 2. 调用通用分页工具（NeedTotal默认true，自动统计总条数）
	var spaces []*Space
	pageResult, err := database.PaginateDefault(ctx, tx, filter.Page, filter.PageSize, &spaces)	
	fmt.Println(pageResult)
	if err != nil {
		return nil,  database.ProcessSQLErrorf(ctx, err)
	}

	return pageResult, nil
}

// GetByID 根据ID查询单个空间（包括已删除的）
func (r *SpaceStore) GetByID(ctx context.Context, id uint) (*Space, error) {
	var out Space
	out.ID = id

	if err := r.db.WithContext(ctx).First(&out).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// GORM 会自动将 DATETIME 类型映射为 *time.Time，通常无需手动转换
	// 如果遇到类型问题，请检查 GORM 配置或数据库驱动设置
	return &out, nil
}

// HasRepositories 检查空间是否有关联的代码库
func (r *SpaceStore) HasRepositories(ctx context.Context, spaceID uint) (bool, error) {
	var count int64
	// TODO: 根据实际的代码库表结构修改此查询
	// 示例：如果代码库表有 space_id 字段
	// err := r.db.WithContext(ctx).Table("ops_repo").
	//     Where("space_id = ? AND deleted = 0", spaceID).
	//     Count(&count).Error

	// 暂时返回 false，表示没有关联代码库，待实现
	// 实际实现时需要检查代码库表
	count = 0
	return count > 0, nil
}

// HasArtifacts 检查空间是否有关联的制品库
func (r *SpaceStore) HasArtifacts(ctx context.Context, spaceID uint) (bool, error) {
	var count int64
	// TODO: 根据实际的制品库表结构修改此查询
	// 示例：如果制品库表有 space_id 字段
	// err := r.db.WithContext(ctx).Table("ops_artifact").
	//     Where("space_id = ? AND deleted = 0", spaceID).
	//     Count(&count).Error

	// 暂时返回 false，表示没有关联制品库，待实现
	// 实际实现时需要检查制品库表
	count = 0
	return count > 0, nil
}

// Delete 逻辑删除空间
func (r *SpaceStore) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		// 查询是否删除.
		var space Space
		if err := tx.Where("id = ?", id).Where("deleted = ?", 0).First(&space).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		// 删除操作.
		if err := tx.Model(&space).UpdateColumns(map[string]interface{}{
			"deleted":    1,
			"editedDate": tx.NowFunc(),
		}).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// Undelete 批量恢复已删除的空间
func (r *SpaceStore) Undelete(ctx context.Context, ids []uint, editedBy string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 状态校验
		var spaces []Space
		if err := tx.Where("id IN ?", ids).Where("deleted = ?", 1).Find(&spaces).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}

		if len(spaces) == 0 {
			return database.ProcessSQLErrorf(ctx, gorm.ErrRecordNotFound)
		}

		// 恢复空间
		updateData := map[string]interface{}{
			"deleted":    0,
			"editedBy":   editedBy,
			"editedDate": tx.NowFunc(),
		}
		if err := tx.Model(&Space{}).
			Where("id IN ?", ids).
			Where("deleted = ?", 1).
			Updates(updateData).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})
}

// mapToModel 将 Space 转换为 SpaceResponse
func mapToModel(data Space) *spaceModel.SpaceResponse {
	return &spaceModel.SpaceResponse{
		ID:          data.ID,
		Name:        data.Name,
		Owner:       data.Owner,
		Permission:  data.Permission,
		Desc:        data.Desc,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
	}
}
