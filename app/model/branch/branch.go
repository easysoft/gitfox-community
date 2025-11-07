package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/rule"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
)

type ReqForCreate struct {
	RepoID      uint   `json:"repoID" binding:"required"`
	Name        string `json:"name" binding:"required,max=100" example:"dev"`
	Target      string `json:"target"`
	CreatedBy   string `json:"createdBy" binding:"required"`
	DryRunRules bool   `json:"dryRunRules" binding:"required"`
	ByPassRules bool   `json:"byPassRules" binding:"required"`
}

type RespForCreate struct {
	Branch
	rule.DryRunRulesOutput
}

type RespForDelete struct {
	rule.DryRunRulesOutput
}

// ReqFilter stores branch query parameters.
type ReqFilter struct {
	RepoID          uint   `json:"repoID"`
	Query           string `json:"query"`
	Label           string `json:"label"`
	IncludeArchived bool   `json:"include_archived"`
	IncludeCommit   bool   `json:"include_commit"`
	Sort            string `json:"sort"`
	Order           string `json:"order"`
	Page            int    `json:"page"`
	Size            int    `json:"size"`
	// BranchMetadataOptions
}

type ReqFilterExt struct {
	IncludeChecks   bool `json:"include_checks"`
	IncludeRules    bool `json:"include_rules"`
	IncludePullReqs bool `json:"include_pullreqs"`
	MaxDivergence   int  `json:"max_divergence"`
}

type Label struct {
	Name  string `json:"name"`
	Total int    `json:"total"`
}

type Branch struct {
	Name      string         `json:"name"`
	SHA       sha.SHA        `json:"sha"`
	Commit    *commit.Commit `json:"commit,omitempty"`
	IsDefault bool           `json:"is_default"`
}

// type BranchExt struct {
// 	Branch
// 	IsDefault        bool               `json:"is_default"`
// 	CheckSummary     *CheckCountSummary `json:"check_summary,omitempty"`
// 	Rules            []RuleInfo         `json:"rules,omitempty"`
// 	PullRequests     []*PullReq         `json:"pull_requests,omitempty"`
// 	CommitDivergence *CommitDivergence  `json:"commit_divergence,omitempty"`
// }
