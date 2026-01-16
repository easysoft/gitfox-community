package repo

import (
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

// PullRequestFilter PR查询过滤器
type PullRequestFilter struct {
	RepoID       uint
	SourceRepoID uint
	TargetRepoID uint
	SourceBranch string
	TargetBranch string
	Status       PullRequestStatus
	CreatedBy    string
	Page         int
	PageSize     int
}

// PullRequestStatus PR状态枚举
type PullRequestStatus string

const (
	PullRequestStatusOpen   PullRequestStatus = "opened"
	PullRequestStatusClosed PullRequestStatus = "closed"
	PullRequestStatusMerged PullRequestStatus = "merged"
)

type MergeCheckStatus string

const (
	// MergeCheckStatusUnchecked merge status has not been checked.
	MergeCheckStatusUnchecked MergeCheckStatus = "unchecked"
	// MergeCheckStatusConflict can’t merge into the target branch due to a potential conflict.
	MergeCheckStatusConflict MergeCheckStatus = "conflict"
	// MergeCheckStatusMergeable branch can merged cleanly into the target branch.
	MergeCheckStatusMergeable MergeCheckStatus = "mergeable"
)

// MergeMethod 合并方法枚举
type MergeMethod string

const (
	MergeMethodMerge       MergeMethod = "merge"
	MergeMethodSquash      MergeMethod = "squash"
	MergeMethodRebase      MergeMethod = "rebase"
	MergeMethodFastForward MergeMethod = "fast-forward"
)

var MergeMethods = []MergeMethod{
	MergeMethodMerge,
	MergeMethodSquash,
	MergeMethodRebase,
	MergeMethodFastForward,
}

func (m MergeMethod) Check() (MergeMethod, bool) {
	switch m {
	case MergeMethodMerge, MergeMethodSquash, MergeMethodRebase, MergeMethodFastForward:
		return m, true
	default:
		return MergeMethodMerge, false
	}
}

// PullRequest 合并请求表（ops_ppmrequests）Store层结构体
type PullRequest struct {
	ID                uint              `json:"id"`
	Title             string            `json:"title"`
	Desc              string            `json:"desc"`
	RepoID            uint              `json:"repoID"`
	SourceRepoID      uint              `json:"sourceRepoID"`
	SourceBranch      string            `json:"sourceBranch"`
	SourceSHA         string            `json:"sourceSHA"`
	TargetRepoID      uint              `json:"targetRepoID"`
	TargetBranch      string            `json:"targetBranch"`
	MergedBy          *string           `json:"mergedBy"`
	MergeMethod       MergeMethod       `json:"mergeMethod"`
	MergeTargetSHA    string            `json:"mergeTargetSHA"`
	MergeBaseSHA      string            `json:"mergeBaseSHA"`
	MergeSHA          string            `json:"mergeSHA"`
	MergeCheckStatus  MergeCheckStatus  `json:"mergeCheckStatus"`
	MergeConflicts    []string          `json:"mergeConflicts"`
	RebaseCheckStatus MergeCheckStatus  `json:"rebaseCheckStatus"`
	RebaseConflicts   []string          `json:"rebaseConflicts"`
	CommitCount       *uint             `json:"commitCount"`
	FileCount         *uint             `json:"fileCount"`
	Flow              int8              `json:"flow"`
	ReviewFlowID      uint              `json:"reviewFlowID"`
	Additions         *uint             `json:"additions"`
	Deletions         *uint             `json:"deletions"`
	Status            PullRequestStatus `json:"status"`
	CreatedBy         string            `json:"createdBy"`
	CreatedDate       *datetime.Local   `json:"createdDate"`
	EditedBy          string            `json:"editedBy"`
	EditedDate        *datetime.Local   `json:"editedDate"`
}

// DiffStats shows total number of commits and modified files. 展示该 PR 包含的提交（Commit）总数、发生变更的文件总数、新增的代码行数、删除的代码行数
type DiffStats struct {
	HeadHash     string `json:"headHash,omitempty"`     // 该 PR 最新的 Commit SHA
	Commits      uint   `json:"commits,omitempty"`      // 该 PR 包含的提交（Commit）总数
	FilesChanged uint   `json:"filesChanged,omitempty"` // 发生变更的文件总数
	Additions    uint   `json:"additions,omitempty"`    // 新增的代码行数
	Deletions    uint   `json:"deletions,omitempty"`    // 删除的代码行数
}

type MergeAfter struct {
	DiffStats

	MergedBy       *string     `json:"mergedBy"`
	MergeMethod    MergeMethod `json:"mergeMethod"`
	MergeTargetSHA string      `json:"mergeTargetSHA"`
	MergeBaseSHA   string      `json:"mergeBaseSHA"`
	MergeSHA       string      `json:"mergeSHA"`
}

// MergeCheck 合并检查结果
type MergeCheck struct {
	Mergeable     bool     `json:"mergeable"`               // 是否可合并
	ConflictFiles []string `json:"conflictFiles,omitempty"` // 冲突文件列表
	Message       string   `json:"message"`                 // 合并检查消息列表
	SourceSHA     string   `json:"sourceSHA"`               // 源分支最新提交 SHA
	TargetSHA     string   `json:"targetSHA"`               // 目标分支最新提交 SHA
}
type FileDiff struct {
	Path      string `json:"path"`
	StartLine int    `json:"start_line"`
}

type Merge struct {
	Method    MergeMethod `json:"method"`    // 合并策略: merge , squash , rebase , fast-forward (*DryRun 时可选)
	SourceSHA string      `json:"sourceSHA"` // 源分支当前的 Commit SHA，用于乐观锁并发控制
	Title     string      `json:"title"`     // 合并提交的标题
	Message   string      `json:"message"`   // 合并提交的详细描述

	BypassRules bool `json:"bypassRules"` // 是否强制绕过分支保护规则（需要管理员权限）
	DryRun      bool `json:"dryRun"`      // 试运行模式 。仅模拟合并，不修改数据
}

func (in *Merge) Check() error {
	var err error
	if in.Method == "" && !in.DryRun {
		err = errors.WithI18nKey(err, i18n.PullrequestMergeCheckMethodError)
		return err
	}

	if in.SourceSHA == "" {
		err = errors.WithI18nKey(err, i18n.PullrequestMergeCheckSourceShaError)
		return err
	}

	if in.Method != "" {
		method, ok := in.Method.Check()
		if !ok {
			err = errors.WithI18nKey(err, i18n.PullrequestMergeCheckMethodError)
			return err
		}

		in.Method = method
	}

	// cleanup title / message (NOTE: git doesn't support white space only)
	in.Title = strings.TrimSpace(in.Title)
	in.Message = strings.TrimSpace(in.Message)

	if (in.Method == MergeMethodRebase || in.Method == MergeMethodFastForward) &&
		(in.Title != "" || in.Message != "") {
		err = errors.WithI18nKey(err, i18n.PullrequestMergeCheckTitleError)
		return err
	}

	return nil
}

// MergeResponse 合并响应
type MergeResponse struct {
	// --- 核心结果字段 ---
	SHA           string `json:"sha"`           // 合并成功后生成的 Commit SHA
	BranchDeleted bool   `json:"branchDeleted"` // 是否自动删除了源分支

	// --- 规则相关 ---
	// 即使合并成功，也可能存在被"绕过"的规则（比如管理员强制合并）
	// 这里列出的是那些"虽然违规但被允许绕过"的规则详情
	RuleViolations []RuleViolations `json:"ruleViolations"`

	// --- Dry Run (试运行) 专用字段 ---
	// 当请求参数 dry_run=true 时，下面这些字段才会有值
	DryRun         bool               `json:"dryRun"`         // 是否为试运行模式
	Mergeable      bool               `json:"mergeable"`      // 是否无代码冲突
	ConflictFiles  []string           `json:"conflictFiles"`  // 如果有冲突，列出文件名
	AllowedMethods []enum.MergeMethod `json:"allowedMethods"` // 当前规则允许的合并方式（如只允许 squash）
}

// MergeViolations 合并冲突和规则违规
type MergeViolations struct {
	Message        string           `json:"message"`        // 错误摘要信息
	ConflictFiles  []string         `json:"conflictFiles"`  // 代码冲突的文件列表（Git 层面冲突）
	RuleViolations []RuleViolations `json:"ruleViolations"` // 规则拦截详情（业务规则冲突）
}

// RuleViolations 规则违规详情
type RuleViolations struct {
	Rule       RuleInfo    `json:"rule"`       // 违规的规则元数据
	Bypassable bool        `json:"bypassable"` // 是否可绕过
	Bypassed   bool        `json:"bypassed"`   // 是否已绕过
	Violations []Violation `json:"violations"` // 具体违规信息列表
}

// RuleInfo 规则元数据
type RuleInfo struct {
	ID      int64 `json:"-"`
	Buildin bool  `json:"buildin"` // 是否为内置规则
}

// Violation 具体违规信息
type Violation struct {
	Code    string `json:"code"`    // 违规码，用于唯一标识违规类型
	Message string `json:"message"` // 违规详情描述
	Params  []any  `json:"params"`  // 违规参数，用于格式化违规详情描述
}
