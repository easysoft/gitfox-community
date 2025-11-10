package commit

import (
	"time"

	gitenum "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
)

const NilSHA = "0000000000000000000000000000000000000000"

// PaginationFilter stores pagination query parameters.
type PaginationFilter struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// CommitFilter stores commit query parameters.
type CommitFilter struct {
	PaginationFilter
	After        string  `json:"after"`
	Path         string  `json:"path"`
	Since        int64   `json:"since"`
	Until        int64   `json:"until"`
	Committer    string  `json:"committer"`
	CommitterIDs []int64 `json:"committer_ids"`
	Author       string  `json:"author"`
	AuthorIDs    []int64 `json:"author_ids"`
	IncludeStats bool    `json:"include_stats"`
}

type ChangeStats struct {
	Insertions int64 `json:"insertions"`
	Deletions  int64 `json:"deletions"`
	Changes    int64 `json:"changes"`
}

type CommitFileStats struct {
	Path    string                 `json:"path"`
	OldPath string                 `json:"old_path,omitempty"`
	Status  gitenum.FileDiffStatus `json:"status"`
	ChangeStats
}

type CommitStats struct {
	Total ChangeStats       `json:"total,omitempty"`
	Files []CommitFileStats `json:"files,omitempty"`
}

type Commit struct {
	SHA        string       `json:"sha"`
	ParentSHAs []string     `json:"parent_shas,omitempty"`
	Title      string       `json:"title"`
	Message    string       `json:"message"`
	Author     Signature    `json:"author"`
	Committer  Signature    `json:"committer"`
	Stats      *CommitStats `json:"stats,omitempty"`
}

type CommitDivergence struct {
	// Ahead is the count of commits the 'From' ref is ahead of the 'To' ref.
	Ahead int32 `json:"ahead"`
	// Behind is the count of commits the 'From' ref is behind the 'To' ref.
	Behind int32 `json:"behind"`
}

type Signature struct {
	Identity Identity  `json:"identity"`
	When     time.Time `json:"when"`
}

type Identity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type RenameDetails struct {
	OldPath         string `json:"old_path"`
	NewPath         string `json:"new_path"`
	CommitShaBefore string `json:"commit_sha_before"`
	CommitShaAfter  string `json:"commit_sha_after"`
}

type ListCommitResponse struct {
	Commits       []Commit        `json:"commits"`
	RenameDetails []RenameDetails `json:"rename_details"`
	TotalCommits  int             `json:"total_commits,omitempty"`
}

type CommitCount struct {
	Date   string `json:"date"`
	Author string `json:"author"`
	CommitCountData
}

type CommitCountData struct {
	Commits   int `json:"commits"`
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Changes   int `json:"changes"`
}
