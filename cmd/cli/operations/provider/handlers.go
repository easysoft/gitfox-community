package provider

import (
	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/handler/codebase"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/router"
)

var HandlersWireSet = wire.NewSet(
	NewHandlers,
)

// NewHandlers 创建所有 HTTP 处理器的聚合实例
func NewHandlers(branchHandler *codebase.BranchHandler) *router.Handlers {
	return &router.Handlers{
		BranchHandler: branchHandler,
		// TODO: 添加更多 handler
		// PipelineHandler: pipelineHandler,
		// ArtifactHandler: artifactHandler,
		// SecurityHandler: securityHandler,
	}
}
