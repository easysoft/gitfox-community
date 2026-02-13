package pipeline

import (
	"context"
	"io"
	"time"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	pkgS3 "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/s3"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/storage"
)

// CronScheduler is the interface used by the pipeline service to manage
// recurring cron jobs in the job scheduler. It is intentionally defined here
// (rather than importing the jobs package) to avoid an import cycle.
type CronScheduler interface {
	AddRecurringWithData(ctx context.Context, jobName, jobType, jobData, cronDef string, maxDur time.Duration) error
	RemoveRecurring(ctx context.Context, jobName string) error
}

var _ Interface = (*Service)(nil)

type Service struct {
	pipelineStore  pipeline.Interface
	s3Store        *pkgS3.Store           // S3 存储，配置开启时非 nil
	contentStorage storage.ContentStorage // 文件存储，始终可用
	logStream      livelog.LogStream      // 实时日志流
	cronScheduler  CronScheduler          // 定时任务调度器，可为 nil
}

func NewPipelineService(
	pipelineStore pipeline.Interface,
	s3Store *pkgS3.Store,
	contentStorage storage.ContentStorage,
	logStream livelog.LogStream,
	cronScheduler CronScheduler,
) *Service {
	return &Service{
		pipelineStore:  pipelineStore,
		s3Store:        s3Store,
		contentStorage: contentStorage,
		logStream:      logStream,
		cronScheduler:  cronScheduler,
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
	// SyncTriggerCronJobs 同步触发器的定时任务（按 triggerID 隔离，避免同 pipeline 多 trigger 互相覆盖）
	SyncTriggerCronJobs(ctx context.Context, triggerID int64, pipelineID int64, newCronExprs []string) error
	// ReconcileCronJobs 启动时对账：遍历所有 trigger，确保 schedule_jobs 表中的 cron job 与 trigger 表一致
	ReconcileCronJobs(ctx context.Context) error

	// --- 日志相关 ---
	// CreateLogs 创建执行日志
	CreateLogs(ctx context.Context, param *pipelineModel.S3Logs, r io.Reader) error
	// GetLogs 获取执行日志，返回 io.ReadCloser 支持流式读取，调用方负责 Close
	GetLogs(ctx context.Context, param *pipelineModel.S3Logs) (io.ReadCloser, error)

	// CreateLogStream 创建日志流
	CreateLogStream(ctx context.Context, stepID int64) error
	// DeleteLogStream 删除日志流
	DeleteLogStream(ctx context.Context, stepID int64) error
	// WriteLogStream 写入日志流
	WriteLogStream(ctx context.Context, stepID int64, line *livelog.Line) error
	// TailLogStream 监听日志流
	TailLogStream(ctx context.Context, stepID int64) (<-chan *livelog.Line, <-chan error)
	// InfoLogStream 获取日志流信息
	InfoLogStream(ctx context.Context) *livelog.LogStreamInfo
}
