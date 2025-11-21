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
	RepoID          uint   `json:"repoID" form:"repoID" binding:"required"`
	Query           string `json:"query" form:"query"`
	Prefix          string `json:"prefix" form:"prefix"`
	IncludeArchived bool   `json:"includeArchived" form:"includeArchived"`
	IncludeCommit   bool   `json:"includeCommit" form:"includeCommit"`
	MaxDivergence   uint   `json:"maxDivergence" form:"maxDivergence"`
	Sort            string `json:"sort" form:"sort" default:"name"`
	Order           string `json:"order" form:"order" default:"asc"`
	Page            int    `json:"page" form:"page" default:"1"`
	Size            int    `json:"size" form:"size" default:"20"`
}

type Prefix struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type Branch struct {
	Name             string                   `json:"name"`
	SHA              string                   `json:"sha"`
	Commit           *commit.Commit           `json:"commit,omitempty"`
	IsDefault        bool                     `json:"isDefault"`
	CommitDivergence *commit.CommitDivergence `json:"commitDivergence,omitempty"`
}
