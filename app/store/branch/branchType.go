package branch

import (
	"context"

	branchModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/slice"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/str"
	"gorm.io/gorm"
)

// ***** BranchType(zt_ops_branch_type) 表在devops 库中 *****
// ***** BranchType(zt_ops_branch_type) 表 Repo 字段 存储的是 devops 库中 zt_repo 表的 id 字段*****

// BatchCreateBranchType 批量创建分支类型(带事务)
func (s *Store) BatchCreateBranchType(ctx context.Context, branchTypes []*BranchType) ([]branchModel.BranchType, error) {
	// 开启事务，使用 CreateInBatches 批量插入
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		for _, branchType := range branchTypes {
			// 检查前缀是否重复
			check, err := s.CheckBranchTypePrefix(ctx, tx, branchType.Repo, 0, str.SplitEx(branchType.Prefixes, ",", true))
			if err != nil {
				err = errors.WithI18nKey(err, i18n.BranchTypeCreateErrorPrefix)
				return err
			}
			if check {
				return errors.WithI18nKey(errors.ErrDuplicate, i18n.BranchTypeCreateErrorPrefix)
			}
		}

		// 每批100条记录
		if err := tx.CreateInBatches(branchTypes, 100).Error; err != nil {
			if errors.Is(database.ProcessSQLErrorf(ctx, err), errors.ErrDuplicate) {
				return errors.WithI18nKey(err, i18n.BranchTypeCreateErrorKey)
			}
			return errors.WithI18nKey(database.ProcessSQLErrorf(ctx, err), i18n.BranchTypeCreateError)
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

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		check, err := s.CheckBranchTypePrefix(ctx, tx, branchType.Repo, branchType.ID, str.SplitEx(branchType.Prefixes, ",", true))
		if err != nil {
			return err
		}
		if check {
			return errors.WithI18nKey(errors.ErrDuplicate, i18n.BranchTypeUpdateErrorPrefix)
		}
		if err := tx.Model(branchType).Updates(branchType).Error; err != nil {
			if errors.Is(err, errors.ErrDuplicate) {
				return errors.WithI18nKey(database.ProcessSQLErrorf(ctx, err), i18n.BranchTypeUpdateErrorKey)
			}
			return errors.WithI18nKey(database.ProcessSQLErrorf(ctx, err), i18n.BranchTypeUpdateError)
		}
		return nil
	})

	if err != nil {
		return nil, err
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

// GetBranchTypesByIDs 根据 ID 列表批量获取分支类型
func (s *Store) GetBranchTypesByIDs(ctx context.Context, repoID uint, ids []uint) ([]branchModel.BranchType, error) {
	var branchTypes []BranchType
	err := s.db.WithContext(ctx).
		Where("id IN ? AND repo = ? AND deleted = 0", ids, repoID).
		Find(&branchTypes).Error
	if err != nil {
		return nil, database.ProcessSQLErrorf(ctx, err)
	}

	result := make([]branchModel.BranchType, len(branchTypes))
	for i, bt := range branchTypes {
		result[i] = *mapBranchTypeToModel(bt)
	}
	return result, nil
}

// DeleteBranchType 删除分支类型
func (s *Store) DeleteBranchType(ctx context.Context, id uint) error {
	if err := s.db.WithContext(ctx).Model(&BranchType{}).Delete("id = ?", id).Error; err != nil {
		return database.ProcessSQLErrorf(ctx, err)
	}
	return nil
}

// ListBranchTypes 分页查询分支类型列表
func (s *Store) ListBranchTypes(ctx context.Context, repoID uint, filter branchModel.ReqFilterBranchType) (*database.PageResult[branchModel.BranchType], error) {
	query := s.db.WithContext(ctx).Model(&BranchType{}).
		Where("repo = ? AND deleted = 0", repoID)

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

// CheckBranchTypePrefix 检查分支类型前缀是否重复
func (s *Store) CheckBranchTypePrefix(ctx context.Context, tx *gorm.DB, repo uint, branchTypeID uint, prefixes []string) (bool, error) {
	var branchTypes []BranchType

	// 可能同前缀的分支类型
	err := tx.WithContext(ctx).Model(&BranchType{}).
		Where("repo = ?", repo).
		Where(func(db *gorm.DB) *gorm.DB {
			if branchTypeID > 0 {
				db = db.Where("id <> ?", branchTypeID)
			}
			return db
		}(tx)).
		Where(func(db *gorm.DB) *gorm.DB {
			for i, p := range prefixes {
				if i == 0 {
					db = db.Where("prefix LIKE ?", "%"+p+"%")
				} else {
					db = db.Or("prefix LIKE ?", "%"+p+"%")
				}
			}
			return db
		}(tx)).
		Find(&branchTypes).Error
	if err != nil {
		return false, database.ProcessSQLErrorf(ctx, err)
	}

	for _, branchType := range branchTypes {
		if len(slice.Intersection(str.SplitEx(branchType.Prefixes, ",", true), prefixes)) > 0 {
			return true, nil
		}
	}

	return false, nil
}

// mapBranchTypeToModel 将 store 模型转换为 API 模型
func mapBranchTypeToModel(data BranchType) *branchModel.BranchType {
	return &branchModel.BranchType{
		ID:          data.ID,
		Repo:        data.Repo,
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
