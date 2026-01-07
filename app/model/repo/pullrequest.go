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
	MergeConflicts    *string           `json:"mergeConflicts"`
	RebaseCheckStatus RebaseCheckStatus `json:"rebaseCheckStatus"`
	RebaseConflicts   *string           `json:"rebaseConflicts"`
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

// DiffStats shows total number of commits and modified files.
type DiffStats struct {
	Commits      uint `json:"commits,omitempty"`       // 该 PR 包含的提交（Commit）总数
	FilesChanged uint `json:"files_changed,omitempty"` // 发生变更的文件总数
	Additions    uint `json:"additions,omitempty"`     // 新增的代码行数
	Deletions    uint `json:"deletions,omitempty"`     // 删除的代码行数
}
