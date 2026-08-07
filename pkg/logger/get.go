package logger

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// ContextLogger 是 logger 在 Context 中的 key
const ContextLogger = "logger"

// contextKey 用于在 context.Context 中存储 logger 的 key 类型
type contextKey string

const LoggerContextKey = contextKey(ContextLogger)

// traceIDContextKey 用于在 context.Context 中存储 trace_id 的 key
const traceIDContextKey = contextKey("trace_id")

var (
	defaultLogger       *SugaredMateLogger
	defaultLoggerOnce   sync.Once
	defaultLogConfig    *config.LoggingConfig
	defaultLogConfigMu  sync.RWMutex
)

// InitDefaultLogger 使用指定的配置初始化默认 logger。
// 应在应用启动时尽早调用（在首次调用 GetLogger 之前）。
// 如果未调用此函数，则回退到仅输出到 stderr 的 zap.NewProduction()。
func InitDefaultLogger(cfg config.LoggingConfig) {
	defaultLogConfigMu.Lock()
	defer defaultLogConfigMu.Unlock()
	defaultLogConfig = &cfg
}

// getDefaultLogger 获取或创建默认的 logger（不带 trace_id）
// 如果已通过 InitDefaultLogger 设置了配置，则使用 lumberjack 轮转写入文件；
// 否则回退到仅输出 stderr 的 zap.NewProduction()。
func getDefaultLogger() *SugaredMateLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogConfigMu.RLock()
		cfg := defaultLogConfig
		defaultLogConfigMu.RUnlock()

		if cfg != nil {
			// 使用专用的 default 日志类型，与 system.log 分离避免 lumberjack 轮转冲突
			zapLogger := NewZapLogger(*cfg, LogTypeDefault)
			defaultLogger = &SugaredMateLogger{
				base:          &ZapMateLogger{Logger: zapLogger, lock: new(sync.Mutex), entryChan: make(chan *logEntry, 1), wg: new(sync.WaitGroup)},
				SugaredLogger: zapLogger.Sugar(),
			}
		} else {
			// 未设置配置时的后备方案
			zapLogger, _ := zap.NewProduction()
			defaultLogger = &SugaredMateLogger{
				base:          &ZapMateLogger{Logger: zapLogger},
				SugaredLogger: zapLogger.Sugar(),
			}
		}
	})
	return defaultLogger
}

// NewNopLogger 创建一个无操作的 logger，用于测试
func NewNopLogger() *SugaredMateLogger {
	nopLogger := zap.NewNop()
	return &SugaredMateLogger{
		base: &ZapMateLogger{
			lock:      new(sync.Mutex),
			isAsync:   false,
			Logger:    nopLogger,
			entryChan: make(chan *logEntry, 1),
			wg:        new(sync.WaitGroup),
		},
		SugaredLogger: nopLogger.Sugar(),
	}
}

// GetLogger 从 context.Context 中获取 logger
// 如果上下文中不存在 logger，返回一个不带 trace_id 的默认 logger
func GetLogger(ctx context.Context) *SugaredMateLogger {
	if log := ctx.Value(LoggerContextKey); log != nil {
		if logger, ok := log.(*SugaredMateLogger); ok {
			return logger
		}
	}
	// 如果上下文中没有 logger，返回默认 logger
	return getDefaultLogger()
}

// WithLogger 将 logger 添加到 context.Context 中
func WithLogger(ctx context.Context, log *SugaredMateLogger) context.Context {
	return context.WithValue(ctx, LoggerContextKey, log)
}

// GetLoggerFromGin 从 gin.Context 中获取 logger
// 如果 gin.Context 中不存在 logger，返回一个不带 trace_id 的默认 logger
func GetLoggerFromGin(c *gin.Context) *SugaredMateLogger {
	if val, exists := c.Get(ContextLogger); exists {
		if log, ok := val.(*SugaredMateLogger); ok {
			return log
		}
	}
	return getDefaultLogger()
}

// GetContextWithLogger 从 gin.Context 中提取所有上下文信息（包括 logger、user_id 等）
// 并返回一个包含这些信息的新 context.Context，用于 Controller 层传递给 Service 层
func GetContextWithLogger(c *gin.Context) context.Context {
	ctx := c.Request.Context()

	// 获取并添加 logger
	log := GetLoggerFromGin(c)
	ctx = WithLogger(ctx, log)

	// 将 gin.Context 中的其他键值对也传递到 context.Context 中
	// 例如 user_id, trace_id 等
	if val, exists := c.Get("user_id"); exists {
		ctx = context.WithValue(ctx, contextKey("user_id"), val)
	}

	if val, exists := c.Get("trace_id"); exists {
		ctx = context.WithValue(ctx, traceIDContextKey, val)
	}

	return ctx
}

// WithTraceID 将 trace_id 添加到 context.Context 中
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDContextKey, traceID)
}

// GetTraceID 从 context.Context 中获取 trace_id，未设置时返回空字符串
func GetTraceID(ctx context.Context) string {
	if val := ctx.Value(traceIDContextKey); val != nil {
		if traceID, ok := val.(string); ok {
			return traceID
		}
	}
	return ""
}

// FromContext 从 context.Context 中获取 logger，与 drone/runner-go/logger.FromContext 兼容。
// 如果上下文中不存在项目 logger，返回默认 logger。
func FromContext(ctx context.Context) *SugaredMateLogger {
	return GetLogger(ctx)
}

// WithContext 将 logger 写入 context.Context 并返回新 context，与 drone/runner-go/logger.WithContext 兼容。
func WithContext(ctx context.Context, log *SugaredMateLogger) context.Context {
	return WithLogger(ctx, log)
}
