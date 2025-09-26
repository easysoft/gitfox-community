package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Level 日志级别类型
type Level int

// 日志级别枚举
const (
	LevelDebug Level = iota // 0 - 调试信息
	LevelInfo               // 1 - 普通信息
	LevelWarn               // 2 - 警告信息
	LevelError              // 3 - 错误信息
	LevelFatal              // 4 - 致命错误（会触发程序退出）
)

// 日志级别字符串映射
var levelNames = map[Level]string{
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelFatal: "FATAL",
}

// 全局日志实例
var (
	globalLogger *Logger
	once         sync.Once
)

// Config 日志配置
type Config struct {
	Level    string `yaml:"level"`    // 日志级别（debug/info/warn/error/fatal）
	Output   string `yaml:"output"`   // 输出方式（stdout/file）
	Path     string `yaml:"path"`     // 日志文件路径（output为file时有效）
	MaxSize  int    `yaml:"max_size"` // 单个日志文件最大大小（MB）
	MaxAge   int    `yaml:"max_age"`  // 日志文件保留天数
	Compress bool   `yaml:"compress"` // 是否压缩归档日志
}

// Logger 日志实例
type Logger struct {
	level  Level
	output io.Writer
	mu     sync.Mutex
}

// 初始化全局日志实例
func Init(cfg Config) {
	once.Do(func() {
		// 解析日志级别
		level := parseLevel(cfg.Level)

		// 初始化输出目的地
		var output io.Writer
		switch cfg.Output {
		case "file":
			// 确保日志目录存在
			if err := os.MkdirAll(filepath.Dir(cfg.Path), 0755); err != nil {
				panic(fmt.Sprintf("创建日志目录失败: %v", err))
			}
			// 使用 lumberjack 实现日志轮转
			output = &lumberjack.Logger{
				Filename: cfg.Path,
				MaxSize:  cfg.MaxSize,
				MaxAge:   cfg.MaxAge,
				Compress: cfg.Compress,
			}
		default:
			// 默认输出到控制台
			output = os.Stdout
		}

		globalLogger = &Logger{
			level:  level,
			output: output,
		}
	})
}

// 解析日志级别字符串为Level类型
func parseLevel(levelStr string) Level {
	switch levelStr {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	default:
		return LevelInfo // 默认info级别
	}
}

// 日志条目结构
type logEntry struct {
	Time    string                 `json:"time"`    // 时间戳
	Level   string                 `json:"level"`   // 日志级别
	Message string                 `json:"message"` // 日志消息
	Caller  string                 `json:"caller"`  // 调用位置
	Fields  map[string]interface{} `json:"fields"`  // 额外字段
}

// 通用日志写入方法
func (l *Logger) log(level Level, format string, v ...interface{}) {
	// 检查日志级别是否需要输出
	if level < l.level {
		return
	}

	// 获取调用位置（跳过当前log函数和公共方法）
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		caller = fmt.Sprintf("%s:%d", filepath.Base(file), line)
	}

	// 格式化消息
	message := fmt.Sprintf(format, v...)

	// 构建日志条目
	entry := logEntry{
		Time:    time.Now().Format("2006-01-02T15:04:05.000Z07:00"),
		Level:   levelNames[level],
		Message: message,
		Caller:  caller,
		Fields:  make(map[string]interface{}),
	}

	// 加锁保证并发安全
	l.mu.Lock()
	defer l.mu.Unlock()

	// 序列化为JSON并写入输出
	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "日志序列化失败: %v\n", err)
		return
	}
	data = append(data, '\n') // 添加换行符
	if _, err := l.output.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "日志写入失败: %v\n", err)
	}

	// 如果是fatal级别，输出后退出程序
	if level == LevelFatal {
		os.Exit(1)
	}
}

// 带上下文的日志方法（支持从context提取请求ID等信息）
func (l *Logger) logWithCtx(ctx context.Context, level Level, format string, v ...interface{}) {
	// 实际实现可以从ctx中提取请求ID、租户ID等信息添加到Fields
	// 简化示例：直接调用基础log方法
	l.log(level, format, v...)
}

// 以下是公共日志方法（按级别划分）

// Debug 输出调试日志
func Debug(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(LevelDebug, format, v...)
	}
}

// DebugCtx 带上下文的调试日志
func DebugCtx(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logWithCtx(ctx, LevelDebug, format, v...)
	}
}

// Info 输出信息日志
func Info(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(LevelInfo, format, v...)
	}
}

// InfoCtx 带上下文的信息日志
func InfoCtx(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logWithCtx(ctx, LevelInfo, format, v...)
	}
}

// Warn 输出警告日志
func Warn(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(LevelWarn, format, v...)
	}
}

// WarnCtx 带上下文的警告日志
func WarnCtx(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logWithCtx(ctx, LevelWarn, format, v...)
	}
}

// Error 输出错误日志
func Error(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(LevelError, format, v...)
	}
}

// ErrorCtx 带上下文的错误日志
func ErrorCtx(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logWithCtx(ctx, LevelError, format, v...)
	}
}

// Fatal 输出致命错误日志并退出
func Fatal(format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(LevelFatal, format, v...)
	}
}

// FatalCtx 带上下文的致命错误日志并退出
func FatalCtx(ctx context.Context, format string, v ...interface{}) {
	if globalLogger != nil {
		globalLogger.logWithCtx(ctx, LevelFatal, format, v...)
	}
}
