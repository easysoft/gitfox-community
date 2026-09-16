package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	gormlogger "gorm.io/gorm/logger"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

// gormLogger 是自定义的 GORM 日志器。
// 相比默认 logger，它在 SQL 输出行前附加 [trace_id=xxx] 前缀，
// 便于按请求链路在同一日志文件中排查问题。
type gormLogger struct {
	*log.Logger
	gormlogger.Config
	traceStr, traceErrStr, traceWarnStr string
}

// newGormLogger 创建一个输出到 writer 的 GORM 日志器
func newGormLogger(writer io.Writer, config gormlogger.Config) gormlogger.Interface {
	return &gormLogger{
		Logger:       log.New(writer, "\r\n", log.LstdFlags),
		Config:       config,
		traceStr:     "%s[%.3fms] [rows:%v] %s",
		traceErrStr:  "%s%s\n[%.3fms] [rows:%v] %s",
		traceWarnStr: "%s%s\n[%.3fms] [rows:%v] %s",
	}
}

// LogMode 切换日志级别，返回新的日志器实例
func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 打印 info 级别日志
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.Printf("[info] "+msg, data...)
	}
}

// Warn 打印 warn 级别日志
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.Printf("[warn] "+msg, data...)
	}
}

// Error 打印 error 级别日志
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.Printf("[error] "+msg, data...)
	}
}

// Trace 打印 SQL 日志，在行首附加 trace_id 前缀（如有）
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	tracePrefix := getTracePrefix(ctx)

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && (!errors.Is(err, gormlogger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		l.Printf(l.traceErrStr, tracePrefix, err, float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, tracePrefix, slowLog, float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
	case l.LogLevel == gormlogger.Info:
		sql, rows := fc()
		l.Printf(l.traceStr, tracePrefix, float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
	}
}

// rowsDisplay 将 rows=-1（无行数信息）显示为 "-"
func rowsDisplay(rows int64) any {
	if rows == -1 {
		return "-"
	}
	return rows
}

// getTracePrefix 从 context 中提取 trace_id 并拼成日志前缀；无 trace_id 时返回空字符串
func getTracePrefix(ctx context.Context) string {
	if traceID := logger.GetTraceID(ctx); traceID != "" {
		return "[trace_id=" + traceID + "] "
	}
	return ""
}
