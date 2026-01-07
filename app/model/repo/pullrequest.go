package repo

import (
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
	PullRequestStatusOpen   PullRequestStatus = "open"
	PullRequestStatusClosed PullRequestStatus = "closed"
	PullRequestStatusMerged PullRequestStatus = "merged"
)

// MergeCheckStatus 合并检查状态枚举（passed/failed/pending/skipped）
type MergeCheckStatus string

const (
	MergeCheckStatusPassed  MergeCheckStatus = "passed"
	MergeCheckStatusFailed  MergeCheckStatus = "failed"
	MergeCheckStatusPending MergeCheckStatus = "pending"
	MergeCheckStatusSkipped MergeCheckStatus = "skipped"
)

// RebaseCheckStatus 变基检查状态枚举（unchecked/passed/failed/pending）
type RebaseCheckStatus string

const (
	RebaseCheckStatusUnchecked RebaseCheckStatus = "unchecked"
	RebaseCheckStatusPassed    RebaseCheckStatus = "passed"
	RebaseCheckStatusFailed    RebaseCheckStatus = "failed"
	RebaseCheckStatusPending   RebaseCheckStatus = "pending"
)

// MergeMethod 合并方法枚举
type MergeMethod string

const (
	MergeMethodMerge       MergeMethod = "merge"
	MergeMethodSquash      MergeMethod = "squash"
	MergeMethodRebase      MergeMethod = "rebase"
	MergeMethodFastForward MergeMethod = "fast-forward"
)

// PullRequest 合并请求表（ops_pullrequests）Store层结构体
type PullRequest struct {
	ID                uint              `gorm:"column:id;primaryKey;autoIncrement;comment:'自增主键'"`
	Title             string            `gorm:"column:title;not null;comment:'PR 标题'"`
	Desc              string            `gorm:"column:desc;type:text;not null;comment:'拉取请求详细描述'"`
	RepoID            uint              `gorm:"column:repoID;not null;comment:'关联zt_repo表的id'"`
	SourceRepoID      uint              `gorm:"column:sourceRepoID;not null;comment:'源仓库ID'"`
	SourceBranch      string            `gorm:"column:sourceBranch;not null;comment:'源分支名称'"`
	SourceSHA         string            `gorm:"column:sourceSHA;type:char(40);not null;comment:'源分支最新提交的SHA-1值'"`
	TargetRepoID      uint              `gorm:"column:targetRepoID;not null;comment:'目标仓库ID'"`
	TargetBranch      string            `gorm:"column:targetBranch;not null;comment:'目标分支名称'"`
	MergedBy          *string           `gorm:"column:mergedBy;comment:'执行合并操作的用户ID'"`
	MergeMethod       MergeMethod       `gorm:"column:mergeMethod;comment:'合并方法'"`
	MergeTargetSHA    string            `gorm:"column:mergeTargetSHA;type:char(40);comment:'合并时目标分支的SHA'"`
	MergeBaseSHA      string            `gorm:"column:mergeBaseSHA;type:char(40);not null;default:'';comment:'合并基准SHA'"`
	MergeSHA          string            `gorm:"column:mergeSHA;type:char(40);comment:'合并后生成的commit SHA'"`
	MergeCheckStatus  MergeCheckStatus  `gorm:"column:mergeCheckStatus;not null;comment:'合并前检查状态'"`
	MergeConflicts    *string           `gorm:"column:mergeConflicts;type:text;comment:'冲突文件列表'"`
	RebaseCheckStatus RebaseCheckStatus `gorm:"column:rebaseCheckStatus;not null;default:'unchecked';comment:'变基检查状态'"`
	RebaseConflicts   *string           `gorm:"column:rebaseConflicts;type:text;comment:'变基冲突详情'"`
	CommitCount       *uint             `gorm:"column:commitCount;comment:'PR包含的commit数量'"`
	FileCount         *uint             `gorm:"column:fileCount;comment:'PR修改的文件数量'"`
	Flow              int8              `gorm:"column:flow;not null;default:0;comment:'流程标识'"`
	ReviewFlowID      uint              `gorm:"column:reviewFlowID;not null;default:0;comment:'关联的评审流程ID'"`
	Additions         *uint             `gorm:"column:additions;comment:'新增的代码行数'"`
	Deletions         *uint             `gorm:"column:deletions;comment:'删除的代码行数'"`
	Status            PullRequestStatus `gorm:"column:status;not null;comment:'拉取请求状态'"`
	CreatedBy         string            `gorm:"column:createdBy;not null;default:'';comment:'创建人'"`
	CreatedDate       *datetime.Local   `gorm:"column:createdDate;comment:'创建时间'"`
	EditedBy          string            `gorm:"column:editedBy;not null;default:'';comment:'更新人'"`
	EditedDate        *datetime.Local   `gorm:"column:editedDate;comment:'更新时间'"`
}

// DiffStats shows total number of commits and modified files.
type DiffStats struct {
	Commits      uint `json:"commits,omitempty"`       // 该 PR 包含的提交（Commit）总数
	FilesChanged uint `json:"files_changed,omitempty"` // 发生变更的文件总数
	Additions    uint `json:"additions,omitempty"`     // 新增的代码行数
	Deletions    uint `json:"deletions,omitempty"`     // 删除的代码行数
}
