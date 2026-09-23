// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pipeline

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/pipeline_store.go -package=mock

import (
	"context"

	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	"github.com/easysoft/gitfox-community/pkg/database"
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

	// --- 流水线相关 ---
	// CreatePipeline 创建流水线
	CreatePipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) (*pipelineModel.Pipeline, error)
	// UpdatePipeline 更新流水线
	UpdatePipeline(ctx context.Context, pipeline *pipelineModel.Pipeline) error
	// GetPipelineByID 根据流水线ID获取流水线（deleted=0）
	GetPipelineByID(ctx context.Context, id int64) (*pipelineModel.Pipeline, error)
	// GetPipelineByRepoIDAndName 根据仓库ID和流水线名称获取流水线（deleted=0）
	GetPipelineByRepoIDAndName(ctx context.Context, repoID int64, name string) (*pipelineModel.Pipeline, error)

	// --- 流水线内容相关 ---
	// CreatePipelineContent 创建流水线内容
	CreatePipelineContent(ctx context.Context, content *pipelineModel.PipelineContent) (*pipelineModel.PipelineContent, error)
	// UpdatePipelineContent 更新流水线内容
	UpdatePipelineContent(ctx context.Context, content *pipelineModel.PipelineContent) error
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

	// --- 触发相关 ---
	// CreateTrigger 创建触发
	CreateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error
	// UpdateTrigger 更新触发
	UpdateTrigger(ctx context.Context, trigger *pipelineModel.Trigger) error
	// GetTriggerByID 根据触发ID获取触发
	GetTriggerByID(ctx context.Context, id int64) (*pipelineModel.Trigger, error)
	// GetTriggerByPipelineID 根据流水线ID获取触发
	GetTriggerByPipelineID(ctx context.Context, pipelineID int64) (*pipelineModel.Trigger, error)
	// GetTriggers 获取触发列表
	GetTriggers(ctx context.Context, filter *pipelineModel.TriggerFilter) ([]*pipelineModel.Trigger, error)
	// DeleteTrigger 删除触发（硬删除）
	DeleteTrigger(ctx context.Context, id int64) error
}
