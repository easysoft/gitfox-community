package router

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/handler/codebase"
)

// Handlers 聚合所有的 HTTP 处理器
type Handlers struct {
	// Codebase 代码库相关处理器
	BranchHandler *codebase.BranchHandler

	// TODO: 添加更多 handler
	// PipelineHandler *pipeline.PipelineHandler
	// ArtifactHandler *artifact.ArtifactHandler
	// SecurityHandler *security.SecurityHandler
}
