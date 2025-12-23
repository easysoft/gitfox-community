package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/str"
	"gorm.io/gorm"
)

// BatchCreateBranchType 批量创建分支类型(带事务)
func (s *Store) BatchCreateBranchType(ctx context.Context, branchTypes []*BranchType) ([]branchModel.BranchType, error) {
	// 开启事务，使用 CreateInBatches 批量插入
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 每批100条记录
		if err := tx.CreateInBatches(branchTypes, 100).Error; err != nil {
			return database.ProcessSQLErrorf(ctx, err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 转换为 model
	results := make([]branchModel.BranchType, len(branchTypes))
	for i, bt := range branchTypes {
		results[i] = *mapBranchTypeToModel(*bt)
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
func (s *Store) ListBranchTypes(ctx context.Context, repoID uint, filter branchModel.ReqFilterBranchType) (*database.PageResult[branchModel.BranchType], error) {
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

	// 执行分页查询
	var branchTypes []BranchType
	pageResult, err := database.PaginateDefault(ctx, query, filter.Page, filter.PageSize, &branchTypes)
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	// 转换为 model 类型
	result := make([]branchModel.BranchType, len(branchTypes))
	for i, bt := range branchTypes {
		result[i] = *mapBranchTypeToModel(bt)
	}

	// 构造返回结果
	return &database.PageResult[branchModel.BranchType]{
		Data:      result,
		Total:     pageResult.Total,
		Page:      pageResult.Page,
		PageSize:  pageResult.PageSize,
		TotalPage: pageResult.TotalPage,
	}, nil
}

// mapBranchTypeToModel 将 store 模型转换为 API 模型
func mapBranchTypeToModel(data BranchType) *branchModel.BranchType {
	return &branchModel.BranchType{
		ID:          data.ID,
		RepoID:      data.RepoID,
		Name:        data.Name,
		Key:         data.Key,
		Prefixes:    str.SplitEx(data.Prefixes, ",", true),
		Desc:        data.Desc,
		CreatedBy:   data.CreatedBy,
		CreatedDate: data.CreatedDate,
		EditedBy:    data.EditedBy,
		EditedDate:  data.EditedDate,
	}
}
