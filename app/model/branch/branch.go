package branch

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/rule"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/sha"
)

type Branch struct {
	Name   string         `json:"name"`
	SHA    sha.SHA        `json:"sha"`
	Commit *commit.Commit `json:"commit,omitempty"`
}

// type BranchExtended struct {
// 	Branch
// 	IsDefault        bool               `json:"is_default"`
// 	CheckSummary     *CheckCountSummary `json:"check_summary,omitempty"`
// 	Rules            []RuleInfo         `json:"rules,omitempty"`
// 	PullRequests     []*PullReq         `json:"pull_requests,omitempty"`
// 	CommitDivergence *CommitDivergence  `json:"commit_divergence,omitempty"`
// }

type CreateBranchRequest struct {
	RepoID      uint   `json:"repoID" binding:"required"`
	Name        string `json:"name" binding:"required,max=100" example:"dev"`
	Target      string `json:"target"`
	CreatedBy   string `json:"createdBy" binding:"required"`
	DryRunRules bool   `json:"dryRunRules" binding:"required"`
	ByPassRules bool   `json:"byPassRules" binding:"required"`
}

type CreateBranchResponse struct {
	Branch
	rule.DryRunRulesOutput
}

type DeleteBranchResponse struct {
	rule.DryRunRulesOutput
}

type BranchMetadataOptions struct {
	IncludeChecks   bool `json:"include_checks"`
	IncludeRules    bool `json:"include_rules"`
	IncludePullReqs bool `json:"include_pullreqs"`
	MaxDivergence   int  `json:"max_divergence"`
}

// BranchFilter stores branch query parameters.
type BranchFilter struct {
	Query           string `json:"query"`
	Label           string `json:"label"`
	IncludeArchived bool   `json:"include_archived"`
	// Sort            enum.Sort             `json:"sort"`
	// Order           enum.Order            `json:"order"`
	Page          int  `json:"page"`
	Size          int  `json:"size"`
	IncludeCommit bool `json:"include_commit"`
	BranchMetadataOptions
}
