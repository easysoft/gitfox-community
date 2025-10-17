package codebase

import (
	"time"
)

// Branch 代码库分支领域模型
type Branch struct {
	ID         string    `gorm:"id"`         // 分支唯一标识
	Name       string    `gorm:"name"`       // 分支名称
	RepoID     string    `gorm:"repo_id"`    // 所属代码库ID
	CreatorID  string    `gorm:"creator_id"` // 创建者ID
	Source     string    `gorm:"source"`     // 来源: manual(手动), auto(自动), webhook(Webhook触发)
	CreatedAt  time.Time `gorm:"created_at"` // 创建时间
	UpdatedAt  time.Time `gorm:"updated_at"` // 更新时间
	Protection bool      `gorm:"protection"` // 是否受保护
}

// Update 更新分支信息
func (b *Branch) Update(name string, protection bool) {
	if name != "" {
		b.Name = name
	}
	b.Protection = protection
	b.UpdatedAt = time.Now()
}
