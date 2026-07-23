package commit

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	gitenum "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/enum"
)

const NilSHA = "0000000000000000000000000000000000000000"

type ReqForCreate struct {
	Message   string   `json:"message"`
	Branch    string   `json:"branch"`
	NewBranch string   `json:"newBranch"`
	Actions   []Action `json:"actions"`
	Author    Author   `json:"author"`
}

type RespForCreate struct {
	SHA string `json:"sha"`
}

type Action struct {
	Action   string                       `json:"action"`
	Path     string                       `json:"path"`
	Payload  string                       `json:"payload"`
	Encoding constant.ContentEncodingType `json:"encoding"`
	SHA      string                       `json:"sha"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// PaginationFilter stores pagination query parameters.
type PaginationFilter struct {
	Page     int `json:"page" form:"page" default:"1"`
	PageSize int `json:"pageSize" form:"pageSize" default:"20"`
}

// ReqFilter stores commit query parameters.
type ReqFilter struct {
	PaginationFilter
	After        string  `json:"after" form:"after"`
	Path         string  `json:"path" form:"path"`
	Since        int64   `json:"since" form:"since"`
	Until        int64   `json:"until" form:"until"`
	Committer    string  `json:"committer" form:"committer"`
	CommitterIDs []int64 `json:"committerIDs" form:"committerIDs"`
	Author       string  `json:"author" form:"author"`
	AuthorIDs    []int64 `json:"authorIDs" form:"authorIDs"`
	IncludeStats bool    `json:"includeStats" form:"includeStats"`
	GitRef       string  `json:"gitRef" form:"gitRef"`
}

type RespForList struct {
	Commits       []Commit        `json:"commits"`
	RenameDetails []RenameDetails `json:"renameDetails"`
	TotalCommits  int             `json:"totalCommits,omitempty"`
}

type RenameDetails struct {
	OldPath         string `json:"oldPath"`
	NewPath         string `json:"newPath"`
	CommitShaBefore string `json:"commitShaBefore"`
	CommitShaAfter  string `json:"commitShaAfter"`
}

type ReqForGet struct {
	SHA string `json:"sha" form:"sha"`
}

type Commit struct {
	SHA        string       `json:"sha"`
	ParentSHAs []string     `json:"parentShas,omitempty"`
	Title      string       `json:"title"`
	Message    string       `json:"message"`
	Author     Signature    `json:"author"`
	Committer  Signature    `json:"committer"`
	Stats      *CommitStats `json:"stats,omitempty"`
}

type CommitStats struct {
	Total ChangeStats       `json:"total,omitempty"`
	Files []CommitFileStats `json:"files,omitempty"`
}

type CommitFileStats struct {
	Path    string                 `json:"path"`
	OldPath string                 `json:"oldPath,omitempty"`
	Status  gitenum.FileDiffStatus `json:"status"`
	ChangeStats
}

type ChangeStats struct {
	Insertions int64 `json:"insertions"`
	Deletions  int64 `json:"deletions"`
	Changes    int64 `json:"changes"`
}

type Signature struct {
	Identity Identity  `json:"identity"`
	When     time.Time `json:"when"`
}

type Identity struct {
	Name  string `json:"name"`
	Email string `json:"email"`
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

type ReqForDivergence struct {
	MaxDivergence uint            `json:"maxDivergence"`
	Requests      []CompareGitRef `json:"requests"`
}

type CommitDivergence struct {
	// Ahead is the count of commits the 'From' ref is ahead of the 'To' ref.
	Ahead int32 `json:"ahead"`
	// Behind is the count of commits the 'From' ref is behind the 'To' ref.
	Behind int32 `json:"behind"`
}

type CompareGitRef struct {
	From string `json:"from"`
	To   string `json:"to"`
}
