package branch

import (
	"context"
	"fmt"
	"time"

	"gitfox.zcorp.cc/git/demo/gitfox-pms/app/store/codebase/branch"
	"gitfox.zcorp.cc/git/demo/gitfox-pms/pkg/logger"
)

// CreateBranchRequest 分支创建请求参数
type CreateBranchRequest struct {
	BranchName string // 分支名称
	RepoID     string // 代码库ID
	UserID     string // 请求用户ID
	// 企业版和旗舰版可能用到的扩展字段
	AutoCreateRule string      // 自动创建规则标识
	WebhookEvent   interface{} // Webhook事件数据
}

// CreateBranchResponse 分支创建响应
type CreateBranchResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	RepoID    string    `json:"repo_id"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// BranchService 分支服务接口定义
type BranchService interface {
	// 创建分支
	Create(ctx context.Context, req CreateBranchRequest) (*CreateBranchResponse, error)
	// 获取分支信息
	Get(ctx context.Context, branchID string) (*CreateBranchResponse, error)
	// // 列出代码库的所有分支
	ListByRepo(ctx context.Context, repoID string) ([]*CreateBranchResponse, error)
	// 删除分支
	Delete(ctx context.Context, branchID, userID string) error
}

// 基础服务结构体，包含通用依赖
type BaseBranchService struct {
	repo branch.Repository
}

// 初始化基础服务
func NewBaseBranchService(repo branch.Repository) *BaseBranchService {
	return &BaseBranchService{
		repo: repo,
	}
}

// Get 实现获取分支信息的通用逻辑
func (s *BaseBranchService) Get(ctx context.Context, branchID string) (*CreateBranchResponse, error) {
	branch, err := s.repo.GetByID(ctx, branchID)
	if err != nil {
		logger.Error("获取分支失败: %v", err)
		return nil, err
	}

	return &CreateBranchResponse{
		ID:        branch.ID,
		Name:      branch.Name,
		RepoID:    branch.RepoID,
		Source:    branch.Source,
		CreatedAt: branch.CreatedAt,
	}, nil
}

// Create 基础版创建分支的默认实现
// 可以被企业版和旗舰版重写
func (s *BaseBranchService) Create(ctx context.Context, req CreateBranchRequest) (*CreateBranchResponse, error) {

	// // 创建领域模型
	// now := time.Now()
	// branch := &codebase.Branch{
	// 	ID:        "ID",
	// 	Name:      req.BranchName,
	// 	RepoID:    req.RepoID,
	// 	CreatorID: req.UserID,
	// 	CreatedAt: now,
	// 	UpdatedAt: now,
	// }

	// // 保存到数据库
	// if err := s.repo.Create(ctx, branch); err != nil {
	// 	logger.ErrorCtx(ctx, "保存分支失败: %v", err)
	// 	return nil, fmt.Errorf("创建分支失败: %w", err)
	// }

	// logger.InfoCtx(ctx, "分支创建成功: %s, 创建者: %s", branch.ID, req.UserID)
	// return &CreateBranchResponse{}, nil

	fmt.Println("Base Service create")
	return &CreateBranchResponse{}, nil
}
