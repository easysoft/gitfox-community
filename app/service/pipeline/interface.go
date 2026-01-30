package pipeline

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
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

	// GetPipelineContent 根据流水线ID获取流水线内容
	GetPipelineContent(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineFile, error)

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
	// ListTriggers 查询触发列表
	ListTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) (*database.PageResult[*pipelineModel.Trigger], error)
	// DeleteTrigger 删除触发
	DeleteTrigger(ctx context.Context, id int64) error
}
