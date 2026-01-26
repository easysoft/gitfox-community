package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
)

var _ Interface = (*Service)(nil)

type Service struct {
	pipelineStore pipeline.Interface
}

func NewPipelineService(pipelineStore pipeline.Interface) *Service {
	return &Service{
		pipelineStore: pipelineStore,
	}
}

type Interface interface {

	// pipelineController 是流水线控制器
	GetPipelineController(ctx context.Context, projectID int64) (*pipelineModel.PipelineFile, error)

	// Migrate migrates the pipeline plugin to the latest version.
	Migrate(ctx context.Context, url string) error
	// GetGroupedPlugins 获取按分组归类的插件列表
	GetGroupedPlugins(ctx context.Context) ([]*pipelineModel.GroupedPlugins, error)
	// GetPluginByName 根据插件名称获取插件
	GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error)
}
