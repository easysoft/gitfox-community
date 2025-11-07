package tag

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
)

// ReqForCreate used for tag creation apis.
type ReqForCreate struct {
	Name        string `json:"name"`
	Target      string `json:"target"`
	Message     string `json:"message"`
	BypassRules bool   `json:"bypassRules"`
}

type ReqFilter struct {
	Query         string `json:"query"`
	Sort          string `json:"sort"`
	Order         string `json:"order"`
	Page          int    `json:"page"`
	Size          int    `json:"size"`
	IncludeCommit bool   `json:"includeCommit"`
}

type Tag struct {
	Name        string            `json:"name"`
	SHA         string            `json:"sha"`
	IsAnnotated bool              `json:"is_annotated"`
	Title       string            `json:"title,omitempty"`
	Message     string            `json:"message,omitempty"`
	Tagger      *commit.Signature `json:"tagger,omitempty"`
	Commit      *commit.Commit    `json:"commit,omitempty"`
}
