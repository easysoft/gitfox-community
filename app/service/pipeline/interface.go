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

	// GetPipelineByID 根据流水线ID获取流水线（deleted=0）
	GetPipelineByID(ctx context.Context, id int64) (*pipelineModel.Pipeline, error)

	// GetPipelineContent 根据流水线ID获取流水线内容
	GetPipelineContent(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineFile, error)

	// GetPipelineVariables 根据流水线ID获取流水线变量，返回 key-value 映射
	// 如果 runtime=true，取 key 和 默认值 defaultValue；如果 runtime=false，取 key 和 固定值 value
	GetPipelineVariables(ctx context.Context, pipelineID int64) (map[string]string, error)

	// Migrate migrates the pipeline plugin to the latest version.
	Migrate(ctx context.Context, url string) error
	// GetGroupedPlugins 获取按分组归类的插件列表
	GetGroupedPlugins(ctx context.Context) ([]*pipelineModel.GroupedPlugins, error)
	// GetPluginByName 根据插件名称获取插件
	GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error)
	// Resolve resolves the pipeline plugin configuration.
	Resolve(ctx context.Context, kind, typ, pluginName string) (*pipelineModel.Yaml, error)

	// --- 触发相关 ---
	// CreateTrigger 创建触发
	CreateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error
	// UpdateTrigger 更新触发
	UpdateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error
	// GetTriggerByID 根据触发ID获取触发
	GetTriggerByID(ctx context.Context, id int64) (*pipelineModel.Trigger, error)
	// GetTriggers 获取触发列表
	GetTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) ([]*pipelineModel.Trigger, error)
	// DeleteTrigger 删除触发
	DeleteTrigger(ctx context.Context, id int64) error
}
