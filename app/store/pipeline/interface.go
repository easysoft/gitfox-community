package pipeline

//go:generate mockgen -source=$GOFILE -destination=mock/plugin_store.go -package=mock

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

// Interface 插件存储接口
type Interface interface {
	// Create 创建插件
	Create(ctx context.Context, plugin *pipelineModel.Plugin) error
	// UpdateByName 更新插件（根据 name）
	UpdateByName(ctx context.Context, plugin *pipelineModel.Plugin) error
	// GetByName 根据插件名称获取插件
	GetByName(ctx context.Context, name string) (*pipelineModel.Plugin, error)
	// GetAll 获取所有插件
	GetAll(ctx context.Context) ([]*pipelineModel.Plugin, error)
	// GetByGroupID 根据分组ID获取插件列表
	GetByGroupID(ctx context.Context, groupID int64) ([]*pipelineModel.Plugin, error)
	// GetAllWithGroup 获取所有插件及其分组信息（关联查询）
	GetAllWithGroup(ctx context.Context) ([]*pipelineModel.PluginWithGroup, error)
	// List 查询插件列表（分页）
	List(ctx context.Context, filter *pipelineModel.PluginFilter) (*database.PageResult[*pipelineModel.Plugin], error)
	// Delete 删除插件（硬删除）
	Delete(ctx context.Context, name string) error

	// GetAllGroup 获取所有插件分组
	GetAllGroup(ctx context.Context) ([]*pipelineModel.PluginGroup, error)
}
