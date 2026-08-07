package database

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	applogger "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// gormLogger 是自定义的 GORM 日志器。
// 相比默认 logger，它在 SQL 输出行前附加 [trace_id=xxx] 前缀，
// 便于按请求链路在同一日志文件中排查问题。
type gormLogger struct {
	*log.Logger
	logger.Config
	traceStr, traceErrStr, traceWarnStr string
}

// newGormLogger 创建一个输出到 writer 的 GORM 日志器
func newGormLogger(writer io.Writer, config logger.Config) logger.Interface {
	return &gormLogger{
		Logger: log.New(writer, "\r\n", log.LstdFlags),
		Config: config,
		traceStr:     "%s%s\n[%.3fms] [rows:%v] %s",
		traceErrStr:  "%s%s %s\n[%.3fms] [rows:%v] %s",
		traceWarnStr: "%s%s %s\n[%.3fms] [rows:%v] %s",
	}
}

// LogMode 切换日志级别，返回新的日志器实例
func (l *gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

// Info 打印 info 级别日志
func (l *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Printf("%s\n[info] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Warn 打印 warn 级别日志
func (l *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Printf("%s\n[warn] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error 打印 error 级别日志
func (l *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Printf("%s\n[error] "+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace 打印 SQL 日志，在行首附加 trace_id 前缀（如有）
func (l *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	tracePrefix := getTracePrefix(ctx)

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		l.Printf(l.traceErrStr, tracePrefix, utils.FileWithLineNum(), err, float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		l.Printf(l.traceWarnStr, tracePrefix, utils.FileWithLineNum(), slowLog, float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		l.Printf(l.traceStr, tracePrefix, utils.FileWithLineNum(), float64(elapsed.Nanoseconds())/1e6, rowsDisplay(rows), sql)
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
	if traceID := applogger.GetTraceID(ctx); traceID != "" {
		return "[trace_id=" + traceID + "] "
	}
	return ""
}
