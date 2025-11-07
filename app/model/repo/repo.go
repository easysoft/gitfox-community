package repo

import (
	"fmt"
	"regexp"
	"time"
)

// CreateRemoteRepoRequest 创建远端仓库请求参数
type CreateRemoteRepoRequest struct {
	SpaceID       uint64 `json:"spaceID" binding:"required"`                // 空间ID（字符串类型，后续转int）
	Identifier    string `json:"identifier" binding:"required"`             // 仓库唯一标识
	DefaultBranch string `json:"defaultBranch"`                             // 默认分支（默认main）
	Desc          string `json:"desc" binding:"max=500"`                    // 仓库描述（最大500字符）
	ForkID        *uint  `json:"forkID" binding:"omitempty"`                // 派生来源ID（可选）
	ACL           string `json:"acl" binding:"required,oneof=open private"` // 权限（仅支持open/private）
	Ignore        string `json:"ignore"`                                    // GitIgnore类型（如go/java）
	License       string `json:"license"`                                   // 许可证类型（如mit/apache）
	Readme        bool   `json:"readme"`                                    // 是否生成README.md
}

// CreateRemoteRepoResponse 创建仓库响应结果
type CreateRemoteRepoResponse struct {
	ID            uint      `json:"id"`            // 数据库仓库ID
	GitUID        string    `json:"gitUID"`        // Git全局唯一标识
	RepoURL       string    `json:"repoURL"`       // 远端仓库克隆地址（HTTP/SSH）
	DefaultBranch string    `json:"defaultBranch"` // 默认分支
	Identifier    string    `json:"identifier"`    // 仓库唯一标识
	SpaceID       uint      `json:"spaceID"`       // 空间ID（转换后int类型）
	CreatedAt     time.Time `json:"createdAt"`     // 创建时间
}


// 自定义参数校验：Identifier 仅允许字母、数字、下划线、连字符
func (req *CreateRemoteRepoRequest) Validate() error {
	// 校验Identifier字符规则（符合Git仓库命名规范）
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]+$`).MatchString(req.Identifier) {
		return fmt.Errorf("identifier仅支持字母、数字、下划线(_)、连字符(-)")
	}
	// 校验Identifier长度（数据库字段varchar(255)，建议限制64字符内）
	if len(req.Identifier) > 64 {
		return fmt.Errorf("identifier长度不能超过64字符")
	}
	// 默认分支处理（未传则设为main）
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}
	return nil
}