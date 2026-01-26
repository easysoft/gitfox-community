package pipeline

//go:generate mockgen -source=$GOFILE -destination=mock/pipeline_store.go -package=mock

import (
	"context"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var _ Interface = (*Store)(nil)

// Store 插件存储实现
type Store struct {
	db *gorm.DB
}

// NewPipelineStore 创建流水线存储实例
func NewPipelineStore(db *gorm.DB) *Store {
	return &Store{
		db: db,
	}
}

type Interface interface {

	// --- 流水线内容相关 ---
	// GetPipelineContentByPipelineID 根据流水线ID获取流水线内容
	GetPipelineContentByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.PipelineContent, error)

	// --- 插件相关 ---
	// CreatePlugin 创建插件
	CreatePlugin(ctx context.Context, plugin *pipelineModel.Plugin) error
	// UpdatePluginByName 更新插件（根据 name）
	UpdatePluginByName(ctx context.Context, plugin *pipelineModel.Plugin) error
	// GetPluginByName 根据插件名称获取插件
	GetPluginByName(ctx context.Context, name string) (*pipelineModel.Plugin, error)
	// GetAllPlugin 获取所有插件
	GetAllPlugin(ctx context.Context) ([]*pipelineModel.Plugin, error)
	// ListPlugin 查询插件列表（分页）
	ListPlugin(ctx context.Context, filter *pipelineModel.PluginFilter) (*database.PageResult[*pipelineModel.Plugin], error)
	// DeletePlugin 删除插件（硬删除）
	DeletePlugin(ctx context.Context, name string) error
	// GetAllWithGroup 获取所有插件及其分组信息（关联查询）
	GetAllWithGroup(ctx context.Context) ([]*pipelineModel.PluginWithGroup, error)

	// --- 插件分组相关 ---
	// GetAllPluginGroup 获取所有插件分组
	GetAllPluginGroup(ctx context.Context) ([]*pipelineModel.PluginGroup, error)
}
