package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
)

type ReqForCreate struct {
	RepoID uint   `json:"repoID" binding:"required"`
	Name   string `json:"name" binding:"required,max=100" example:"dev"`
	Source string `json:"source"`
}

// ReqFilter stores branch query parameters.
type ReqFilter struct {
	RepoID          uint   `json:"repoID" binding:"required"`
	Query           string `json:"query"`
	Prefix          string `json:"prefix"`
	IncludeArchived bool   `json:"includeArchived"`
	IncludeCommit   bool   `json:"includeCommit"`
	MaxDivergence   uint   `json:"maxDivergence"`
	Sort            string `json:"sort"`
	Order           string `json:"order"`
	Page            int    `json:"page"`
	Size            int    `json:"size"`
}

type Prefix struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type Branch struct {
	Name             string                   `json:"name"`
	SHA              string                   `json:"sha"`
	Commit           *commit.Commit           `json:"commit,omitempty"`
	IsDefault        bool                     `json:"is_default"`
	CommitDivergence *commit.CommitDivergence `json:"commitDivergence,omitempty"`
}
