package branch

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"gitfox.zcorp.cc/demo/gitfox-pms/internal/domain/codebase"
)

// MySQLRepository 基于GORM的分支数据访问实现
type MySQLRepository struct {
	db *gorm.DB // GORM数据库实例
}

// 确保实现了BranchRepository接口
var _ Repository = (*MySQLRepository)(nil)

// NewMySQLRepository 创建MySQL仓库实例
func NewMySQLRepository(db *gorm.DB) *MySQLRepository {
	return &MySQLRepository{
		db: db,
	}
}

// Create 新增分支记录
func (r *MySQLRepository) Create(ctx context.Context, branch *codebase.Branch) error {
	// 使用GORM的Create方法插入记录，自动填充主键和时间戳
	result := r.db.WithContext(ctx).Create(branch)
	if result.Error != nil {
		return wrapDBError(result.Error, "创建分支记录失败")
	}
	return nil
}

// GetByID 根据ID查询分支
func (r *MySQLRepository) GetByID(ctx context.Context, id string) (*codebase.Branch, error) {
	var branch codebase.Branch
	// 使用GORM的First方法查询单条记录
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&branch)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrBranchNotFound // 自定义未找到错误
		}
		return nil, wrapDBError(result.Error, "查询分支记录失败")
	}
	return &branch, nil
}

// ListByRepoID 按代码库ID查询分支列表
func (r *MySQLRepository) ListByRepoID(ctx context.Context, repoID string, opts ListOptions) ([]*codebase.Branch, int64, error) {
	var (
		branches []*codebase.Branch
		total    int64
	)

	// 构建查询条件
	query := r.db.WithContext(ctx).
		Where("repo_id = ?", repoID)

	// 处理搜索条件
	if opts.Search != "" {
		query = query.Where("name LIKE ?", "%"+opts.Search+"%")
	}

	// 处理状态筛选
	if opts.Status > 0 {
		query = query.Where("status = ?", opts.Status)
	}

	// 计算总数（分页需要）
	if err := query.Model(&codebase.Branch{}).Count(&total).Error; err != nil {
		return nil, 0, wrapDBError(err, "统计分支总数失败")
	}

	// 处理分页
	if opts.Page > 0 && opts.PageSize > 0 {
		offset := (opts.Page - 1) * opts.PageSize
		query = query.Offset(offset).Limit(opts.PageSize)
	}

	// 处理排序
	if opts.SortBy != "" {
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: opts.SortBy},
			Desc:   true,
		})
	} else {
		// 默认按创建时间倒序
		query = query.Order("created_at DESC")
	}

	// 执行查询
	if err := query.Find(&branches).Error; err != nil {
		return nil, 0, wrapDBError(err, "查询分支列表失败")
	}

	return branches, total, nil
}

// Update 更新分支信息
func (r *MySQLRepository) Update(ctx context.Context, branch *codebase.Branch) error {
	// 使用GORM的Save方法更新记录（全量更新）
	// 如需部分更新，可使用Updates方法配合map或结构体
	result := r.db.WithContext(ctx).
		Where("id = ?", branch.ID).
		Save(branch)

	if result.Error != nil {
		return wrapDBError(result.Error, "更新分支记录失败")
	}

	// 检查是否有记录被更新
	if result.RowsAffected == 0 {
		return ErrBranchNotFound
	}

	return nil
}

// Delete 软删除分支（如果需要物理删除，可使用Unscoped()）
func (r *MySQLRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&codebase.Branch{})

	if result.Error != nil {
		return wrapDBError(result.Error, "删除分支记录失败")
	}

	if result.RowsAffected == 0 {
		return ErrBranchNotFound
	}

	return nil
}

// 包装数据库错误，添加业务上下文信息
func wrapDBError(err error, message string) error {
	return errors.New(message + ": " + err.Error())
}
