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
	Query         string `json:"query" form:"query"`
	Sort          string `json:"sort" form:"sort" default:"name"`
	Order         string `json:"order" form:"order" default:"asc"`
	Page          int    `json:"page" form:"page" default:"1"`
	PageSize      int    `json:"pageSize" form:"pageSize" default:"20"`
	IncludeCommit bool   `json:"includeCommit" form:"includeCommit"`
	RepoID        uint   `json:"repoID" form:"repoID" binding:"required"`
}

type Tag struct {
	Name      string            `json:"name"`
	SHA       string            `json:"sha"`
	Message   string            `json:"message,omitempty"`
	CreatedBy *commit.Signature `json:"createdBy,omitempty"`
	Commit    *commit.Commit    `json:"commit,omitempty"`
}
