package repo

import "github.com/easysoft/gitfox/types"

// GetContentReq 获取仓库内容请求参
type GetContentReq struct {
	RepoId              int64  `json:"repoId"`
	GitRef              string `json:"gitRef"`
	RepoPath            string `json:"repoPath"`
	IncludeLatestCommit bool   `json:"includeLatestCommit"`
	FlattenDirectories  bool   `json:"flattenDirectories"`
}

// GetContentOutput 获取仓库内容响应结果
type GetContentOutput struct {
	ContentInfo
	Content Content `json:"content"`
}

type ContentInfo struct {
	Type         ContentType   `json:"type"`
	SHA          string        `json:"sha"`
	Name         string        `json:"name"`
	Path         string        `json:"path"`
	LatestCommit *types.Commit `json:"latest_commit,omitempty"`
}

type ContentType string

const (
	ContentTypeFile      ContentType = "file"
	ContentTypeDir       ContentType = "dir"
	ContentTypeSymlink   ContentType = "symlink"
	ContentTypeSubmodule ContentType = "submodule"
)

// Content restricts the possible types of content returned by the api.
// 内容接口，限制API返回的可能类型的内容
type Content interface {
	isContent()
}
