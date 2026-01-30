package logger

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ContextLogger 是 logger 在 Context 中的 key
const ContextLogger = "logger"

// contextKey 用于在 context.Context 中存储 logger 的 key 类型
type contextKey string

const LoggerContextKey = contextKey(ContextLogger)

var (
	defaultLogger     *SugaredMateLogger
	defaultLoggerOnce sync.Once
)

// getDefaultLogger 获取或创建默认的 logger（不带 trace_id）
func getDefaultLogger() *SugaredMateLogger {
	defaultLoggerOnce.Do(func() {
		zapLogger, _ := zap.NewProduction()
		defaultLogger = &SugaredMateLogger{
			base:          &ZapMateLogger{Logger: zapLogger},
			SugaredLogger: zapLogger.Sugar(),
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
		ctx = context.WithValue(ctx, contextKey("trace_id"), val)
	}

	return ctx
}
