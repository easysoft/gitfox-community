package branch

import (
	"context"
	"errors"

	"gitfox.zcorp.cc/easysoft/gitfox-pms/internal/domain/codebase"
)

// Repository 分支数据访问层接口
type Repository interface {
	// Create 新增分支记录
	Create(ctx context.Context, branch *codebase.Branch) error

	// GetByID 根据ID查询分支
	GetByID(ctx context.Context, id string) (*codebase.Branch, error)

	// ListByRepoID 按代码库ID查询分支列表
	ListByRepoID(ctx context.Context, repoID string, opts ListOptions) ([]*codebase.Branch, int64, error)

	// Update 更新分支信息
	Update(ctx context.Context, branch *codebase.Branch) error

	// Delete 删除分支记录
	Delete(ctx context.Context, id string) error
}

// ListOptions 列表查询选项
type ListOptions struct {
	Page     int    // 页码（从1开始）
	PageSize int    // 每页条数
	Search   string // 搜索关键词（分支名）
	Status   int    // 状态筛选（0表示不筛选）
	SortBy   string // 排序字段
	SortDesc bool   // 是否倒序
}

// 自定义错误变量
var (
	ErrBranchNotFound = errors.New("分支记录不存在")
)
