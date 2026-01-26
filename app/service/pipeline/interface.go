package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
)

type PluginService interface {
	Migrate(ctx context.Context, url string) error
	// GetGroupedPlugins 获取按分组归类的插件列表
	GetGroupedPlugins(ctx context.Context) ([]*pipelineModel.GroupedPlugins, error)
	// GetPluginByName 根据插件名称获取插件
	GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error)
}
