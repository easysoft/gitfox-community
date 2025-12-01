package repo

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
)

const (
	// maxGetContentFileSize specifies the maximum number of bytes a file content response contains.
	// If a file is any larger, the content is truncated.
	MaxGetContentFileSize = 1 << 22 // 4 MB
)

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
	Type         ContentType    `json:"type"`
	SHA          string         `json:"sha"`
	Name         string         `json:"name"`
	Path         string         `json:"path"`
	LatestCommit *commit.Commit `json:"latest_commit,omitempty"`
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

type FileContent struct {
	Encoding constant.ContentEncodingType `json:"encoding"`
	Data     string                       `json:"data"`
	Size     int64                        `json:"size"`
	DataSize int64                        `json:"data_size"`
}

func (c *FileContent) isContent() {}

type SymlinkContent struct {
	Target string `json:"target"`
	Size   int64  `json:"size"`
}

func (c *SymlinkContent) isContent() {}

type DirContent struct {
	Entries []ContentInfo `json:"entries"`
}

func (c *DirContent) isContent() {}

type SubmoduleContent struct {
	URL       string `json:"url"`
	CommitSHA string `json:"commit_sha"`
}

func (c *SubmoduleContent) isContent() {}

type PathsInput struct {
	IncludeDirectories bool   `json:"includeDirectories"`
	GitRef             string `json:"gitRef"`
}

type PathsOutput struct {
	Files       []string `json:"files"`
	Directories []string `json:"directories"`
}

type PathsDetailsInput struct {
	GitRef string   `json:"gitRef"`
	Paths  []string `json:"paths"`
}

type PathsDetailsOutput struct {
	Details []git.PathDetails `json:"details"`
}
