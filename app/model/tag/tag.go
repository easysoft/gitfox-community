package tag

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
)

// ReqForCreate used for tag creation apis.
type ReqForCreate struct {
	Name    string `json:"name" binding:"required"`
	Source  string `json:"source"`
	Message string `json:"message"`
	RepoID  uint   `json:"repoID" binding:"required"`
}

type ReqFilter struct {
	Query         string `json:"query"`
	Sort          string `json:"sort"`
	Order         string `json:"order"`
	Page          int    `json:"page"`
	Size          int    `json:"size"`
	IncludeCommit bool   `json:"includeCommit"`
	RepoID        uint   `json:"repoID" binding:"required"`
}

type Tag struct {
	Name      string            `json:"name"`
	SHA       string            `json:"sha"`
	Message   string            `json:"message,omitempty"`
	CreatedBy *commit.Signature `json:"createdBy,omitempty"`
	Commit    *commit.Commit    `json:"commit,omitempty"`
}
