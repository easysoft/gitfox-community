package repo

import (
	"regexp"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

// Repository represents a code repository.
type Repository struct {
	// TODO: int64 ID doesn't match DB
	ID          int64  `json:"id" yaml:"id"`
	Version     int64  `json:"-" yaml:"-"`
	ParentID    int64  `json:"parent_id" yaml:"parent_id"`
	Identifier  string `json:"identifier" yaml:"identifier"`
	Path        string `json:"path" yaml:"path"`
	Description string `json:"description" yaml:"description"`
	CreatedBy   int64  `json:"created_by" yaml:"created_by"`
	Created     int64  `json:"created" yaml:"created"`
	Updated     int64  `json:"updated" yaml:"updated"`
	Deleted     *int64 `json:"deleted,omitempty" yaml:"deleted"`

	// Size of the repository in KiB.
	Size int64 `json:"size" yaml:"size"`
	// SizeUpdated is the time when the Size was last updated.
	SizeUpdated int64 `json:"size_updated" yaml:"size_updated"`

	GitUID        string `json:"-" yaml:"-"`
	DefaultBranch string `json:"default_branch" yaml:"default_branch"`
	ForkID        int64  `json:"fork_id" yaml:"fork_id"`
	PullReqSeq    int64  `json:"-" yaml:"-"`

	NumForks       int `json:"num_forks" yaml:"num_forks"`
	NumPulls       int `json:"num_pulls" yaml:"num_pulls"`
	NumClosedPulls int `json:"num_closed_pulls" yaml:"num_closed_pulls"`
	NumOpenPulls   int `json:"num_open_pulls" yaml:"num_open_pulls"`
	NumMergedPulls int `json:"num_merged_pulls" yaml:"num_merged_pulls"`

	Mirror  bool               `json:"mirror,omitempty" yaml:"mirror"`
	State   constant.RepoState `json:"state" yaml:"-"`
	IsEmpty bool               `json:"is_empty,omitempty" yaml:"is_empty"`

	// git urls
	GitURL    string `json:"git_url" yaml:"-"`
	GitSSHURL string `json:"git_ssh_url,omitempty" yaml:"-"`
}

// CreateRemoteRepoRequest 创建远端仓库请求参数
type CreateRepoReq struct {
	SpaceID       uint   `json:"spaceID" binding:"required"`                // 空间ID（字符串类型，后续转int）
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
type CreateRepoResp struct {
	ID            uint       `json:"id"`
	SpaceID       uint       `json:"spaceID"`
	Identifier    string     `json:"identifier"`
	Desc          string     `json:"desc"`
	GitUID        string     `json:"gitUID"`
	DefaultBranch string     `json:"defaultBranch"`
	ForkID        *uint      `json:"forkID"` // 允许 NULL，用指针类型
	Empty         bool       `json:"empty"`
	ACL           string     `json:"acl"`
	Mirror        bool       `json:"mirror"`
	Status        int        `json:"status"` // 状态码用 int 类型（兼容枚举映射）
	Mutex         int        `json:"mutex"`
	CreatedBy     string     `json:"createdBy"`
	CreatedDate   *time.Time `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy      string     `json:"editedBy"`
	EditedDate    *time.Time `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted       bool       `json:"deleted"`
}

// 自定义参数校验：Identifier 仅允许字母、数字、下划线、连字符
func (req *CreateRepoReq) Validate() error {

	// 校验Identifier字符规则（符合Git仓库命名规范）
	if !regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString(req.Identifier) {
		return errors.Newf("identifier仅支持字母、数字、下划线(_)、连字符(-)、点号(.)")
	}

	if len(req.Identifier) > 64 {
		return errors.Newf("identifier长度不能超过64字符")
	}

	identifierLower := strings.ToLower(req.Identifier)
	if strings.HasSuffix(identifierLower, ".git") {
		return errors.Newf("repository identifiers cannot end with .git")
	}

	if len(req.Desc) > constant.MaxDescByteCountUTF8MB4 {
		return errors.Newf("Description can be at most %d in length.", constant.MaxDescByteCountUTF8MB4)
	}

	// 默认分支处理（未传则设为main）
	if req.DefaultBranch == "" {
		req.DefaultBranch = "main"
	}
	return nil
}

type Repo struct {
	ID            uint       `json:"id"`
	SpaceID       int        `json:"spaceID"`
	Identifier    string     `json:"identifier"`
	Desc          string     `json:"desc"`
	GitUID        string     `json:"gitUID"`
	DefaultBranch string     `json:"defaultBranch"`
	ForkID        *uint      `json:"forkID"` // 允许 NULL，用指针类型
	Empty         bool       `json:"empty"`
	ACL           string     `json:"acl"`
	Mirror        bool       `json:"mirror"`
	Status        int        `json:"status"` // 状态码用 int 类型（兼容枚举映射）
	Mutex         int        `json:"mutex"`
	CreatedBy     string     `json:"createdBy"`
	CreatedDate   *time.Time `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy      string     `json:"editedBy"`
	EditedDate    *time.Time `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted       bool       `json:"deleted"`
}

type BlameFilter struct {
	GitRef   string `json:"gitRef" binding:"omitempty"`   // 非必填，omitempty 允许空值
	LineFrom int    `json:"lineFrom" binding:"omitempty"` // 非必填，omitempty 允许空值
	LineTo   int    `json:"lineTo" binding:"omitempty"`   // 非必填，omitempty 允许空值
}
