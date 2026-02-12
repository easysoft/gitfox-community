package repo

import (
	"regexp"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	gitTypes "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/api"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// CreateRemoteRepoRequest 创建远端仓库请求参数
type CreateRepoReq struct {
	SpaceID       uint   `json:"spaceID" binding:"required"`                // 空间ID（字符串类型，后续转int）
	Identifier    string `json:"identifier" binding:"required"`             // 仓库唯一标识
	DefaultBranch string `json:"defaultBranch"`                             // 默认分支（默认main）
	Desc          string `json:"desc" binding:"max=500"`                    // 仓库描述（最大500字符）
	Product       string `json:"product"`                                   // 关联产品
	ForkID        *uint  `json:"forkID" binding:"omitempty"`                // 派生来源ID（可选）
	ACL           string `json:"acl" binding:"required,oneof=open private"` // 权限（仅支持open/private）
	Ignore        string `json:"ignore"`                                    // GitIgnore类型（如go/java）
	License       string `json:"license"`                                   // 许可证类型（如mit/apache）
	Readme        bool   `json:"readme"`                                    // 是否生成README.md
}

// ReqFilter stores repo query parameters.
type ReqFilter struct {
	Query    string `json:"query" form:"query"`
	Page     int    `json:"page" form:"page" default:"1"`
	PageSize int    `json:"pageSize" form:"pageSize" default:"20"`
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

// ReqForUpdate 更新仓库的请求参数
type ReqForUpdate struct {
	SpaceID       uint   `json:"spaceID" binding:"omitempty"`                // 空间ID（字符串类型，后续转int）
	Identifier    string `json:"identifier" binding:"omitempty"`             // 仓库唯一标识
	DefaultBranch string `json:"defaultBranch" binding:"omitempty"`          // 默认分支（默认main）
	Product       string `json:"product" binding:"omitempty"`                // 关联产品
	Desc          string `json:"desc" binding:"omitempty,max=500"`           // 仓库描述（最大500字符）
	ACL           string `json:"acl" binding:"omitempty,oneof=open private"` // 权限（仅支持open/private）
}

type Repo struct {
	ID            uint   `json:"id"`
	SpaceID       uint   `json:"spaceID"`
	Identifier    string `json:"identifier"`
	Desc          string `json:"desc"`
	Product       string `json:"product"`
	GitUID        string `json:"gitUID"`
	DefaultBranch string `json:"defaultBranch"`
	ForkID        *uint  `json:"forkID"` // 允许 NULL，用指针类型
	Empty         bool   `json:"empty"`
	ACL           string `json:"acl"`
	Mirror        bool   `json:"mirror"`
	Synced        uint8  `json:"synced"`
	Status        int    `json:"status"` // 状态码用 int 类型（兼容枚举映射）
	Mutex         int    `json:"mutex"`
	// git urls
	GitURL      string          `json:"gitURL,omitempty"`
	GitSSHURL   string          `json:"gitSSHURL,omitempty"`
	Path        string          `json:"path"`
	CreatedBy   string          `json:"createdBy"`
	CreatedDate *datetime.Local `json:"createdDate"` // 允许 NULL，用指针类型
	EditedBy    string          `json:"editedBy"`
	EditedDate  *datetime.Local `json:"editedDate"` // 允许 NULL，用指针类型
	Deleted     uint            `json:"deleted"`
}

type ReqForDiff struct {
	Range string                     `json:"range"`
	Files []gitTypes.FileDiffRequest `json:"files"`
}

type BlameFilter struct {
	GitRef   string `json:"gitRef" binding:"required" form:"gitRef"`      // 非必填，omitempty 允许空值
	LineFrom int    `json:"lineFrom" binding:"omitempty" form:"lineFrom"` // 非必填，omitempty 允许空值
	LineTo   int    `json:"lineTo" binding:"omitempty" form:"lineTo"`     // 非必填，omitempty 允许空值
}

type CompareInfo struct {
	BaseRef   string
	HeadRef   string
	MergeBase bool
}
