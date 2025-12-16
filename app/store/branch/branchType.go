package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

// CreateBranchType 创建分支类型
func (s *Store) CreateBranchType(ctx context.Context, branchType *BranchType) (*branchModel.BranchType, error) {
	if err := s.db.WithContext(ctx).Create(branchType).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}
	return mapBranchTypeToModel(*branchType), nil
}

// BatchCreateBranchType 批量创建分支类型(带事务)
func (s *Store) BatchCreateBranchType(ctx context.Context, branchTypes []*BranchType) ([]branchModel.BranchType, error) {
	results := make([]branchModel.BranchType, 0, len(branchTypes))

	// 开启事务
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, branchType := range branchTypes {
			if err := tx.Create(branchType).Error; err != nil {
				return database.ProcessSQLErrorf(ctx, err)
			}
			results = append(results, *mapBranchTypeToModel(*branchType))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// UpdateBranchType 更新分支类型
func (s *Store) UpdateBranchType(ctx context.Context, branchType *BranchType, id uint) (*branchModel.BranchType, error) {
	branchType.ID = id
	if err := s.db.WithContext(ctx).Model(branchType).Updates(branchType).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 重新查询完整记录,因为 Updates 只更新指定非零值字段,不会回填其他字段
	var updated BranchType
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&updated).Error; err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapBranchTypeToModel(updated), nil
}

// GetBranchTypeByID 根据 ID 获取分支类型
func (s *Store) GetBranchTypeByID(ctx context.Context, id uint) (*branchModel.BranchType, error) {
	var branchType BranchType
	err := s.db.WithContext(ctx).Where("id = ? AND deleted = 0", id).First(&branchType).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	return mapBranchTypeToModel(branchType), nil
}

// DeleteBranchType 删除分支类型(软删除)
func (s *Store) DeleteBranchType(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&BranchType{}).Where("id = ?", id).Update("deleted", 1).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListBranchTypes 分页查询分支类型列表
func (s *Store) ListBranchTypes(ctx context.Context, repoID uint, filter branchModel.ReqFilterBranchType) ([]branchModel.BranchType, int64, error) {
	var branchTypes []BranchType
	var total int64

	query := s.db.WithContext(ctx).Model(&BranchType{}).
		Where("repoID = ? AND deleted = 0", repoID)

	// 如果有查询条件,添加模糊查询
	if filter.Name != "" {
		query = query.Where("name LIKE ?", "%"+filter.Name+"%")
	}
	if filter.Key != "" {
		query = query.Where("`key` LIKE ?", "%"+filter.Key+"%")
	}
	if filter.Prefix != "" {
		query = query.Where("prefix LIKE ?", "%"+filter.Prefix+"%")
	}
	if filter.Desc != "" {
		query = query.Where("desc LIKE ?", "%"+filter.Desc+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}

	// 应用分页
	offset := (filter.Page - 1) * filter.PageSize
	limit := filter.PageSize

	if err := query.Offset(offset).Limit(limit).Order("createdDate DESC").Find(&branchTypes).Error; err != nil {
		return nil, 0, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]branchModel.BranchType, len(branchTypes))
	for i, bt := range branchTypes {
		result[i] = *mapBranchTypeToModel(bt)
	}

	return result, total, nil
}

// mapBranchTypeToModel 将 store 模型转换为 API 模型
func mapBranchTypeToModel(data BranchType) *branchModel.BranchType {
	return &branchModel.BranchType{
		ID:          data.ID,
		RepoID:      data.RepoID,
		Name:        data.Name,
		Key:         data.Key,
		Prefixes:    data.Prefixes,
		Desc:        data.Desc,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
	}
}
