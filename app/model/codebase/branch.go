package codebase

import (
	"errors"
	"time"
)

// Branch Query
type BranchQuery struct {
	Name   string `json:"name"`    // 分支名称
	RepoID string `json:"repo_id"` // 所属代码库ID
}

// Validate 验证分支查询参数合法性
func (bq *BranchQuery) Validate() error {
	return nil
}

// Branch 代码库分支领域模型
type Branch struct {
	ID         string    `json:"id"`         // 分支唯一标识
	Name       string    `json:"name"`       // 分支名称
	RepoID     string    `json:"repo_id"`    // 所属代码库ID
	CreatorID  string    `json:"creator_id"` // 创建者ID
	Source     string    `json:"source"`     // 来源: manual(手动), auto(自动), webhook(Webhook触发)
	CreatedAt  time.Time `json:"created_at"` // 创建时间
	UpdatedAt  time.Time `json:"updated_at"` // 更新时间
	Protection bool      `json:"protection"` // 是否受保护
}

// Validate 验证分支信息合法性
func (b *Branch) Validate() error {
	if b.Name == "" {
		return errors.New("分支名称不能为空")
	}
	if b.RepoID == "" {
		return errors.New("代码库ID不能为空")
	}
	if b.CreatorID == "" {
		return errors.New("创建者ID不能为空")
	}
	return nil
}

// IsProtected 判断分支是否受保护
func (b *Branch) IsProtected() bool {
	return b.Protection
}
