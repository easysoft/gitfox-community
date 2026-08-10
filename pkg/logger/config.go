package logger

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

const (
	LogTypeAPIServer = "apiserver"
	LogTypeRunner    = "runner"
	LogTypeSSH       = "ssh"
	LogTypeScan      = "scan"
	LogTypeSystem    = "system"
	LogTypeJob       = "job"
	LogTypeSQL       = "sql"
	LogTypeDefault   = "default"
)

func getLogPath(c config.LoggingConfig, logType string) string {
	switch logType {
	case LogTypeAPIServer:
		return filepath.Join(c.LogDir, c.LogAPIServerConf.Filename)
	case LogTypeRunner:
		return filepath.Join(c.LogDir, c.LogRunnerConf.Filename)
	case LogTypeSSH:
		return filepath.Join(c.LogDir, c.LogSSHConf.Filename)
	case LogTypeScan:
		return filepath.Join(c.LogDir, c.LogScanConf.Filename)
	case LogTypeSystem:
		return filepath.Join(c.LogDir, c.LogSystemConf.Filename)
	case LogTypeJob:
		return filepath.Join(c.LogDir, c.LogJobConf.Filename)
	case LogTypeSQL:
		if c.LogSQLConf.Filename == "" {
			return filepath.Join(c.LogDir, "sql.log")
		}
		return filepath.Join(c.LogDir, c.LogSQLConf.Filename)
	case LogTypeDefault:
		if c.LogDefaultConf.Filename == "" {
			return filepath.Join(c.LogDir, "default.log")
		}
		return filepath.Join(c.LogDir, c.LogDefaultConf.Filename)
	default:
		// 如果传入未知类型，默认使用 system 日志
		return filepath.Join(c.LogDir, c.LogSystemConf.Filename)
	}
}

// ensureLogDir 确保日志目录存在，只在初始化时调用一次
func ensureLogDir(logPath string) error {
	if logPath == "" {
		return nil
	}
	logDir := filepath.Dir(logPath)
	return os.MkdirAll(logDir, 0755)
}

var (
	rotatelogsRegistry = make(map[string]*rotatelogs.RotateLogs)
	rotatelogsMu       sync.Mutex
)

// sharedRotatelogs 返回指定日志路径对应的 rotatelogs 实例（如 system.log.2026-08-07）。
func sharedRotatelogs(config config.LoggingConfig, logPath string) (*rotatelogs.RotateLogs, error) {
	rotatelogsMu.Lock()
	defer rotatelogsMu.Unlock()
	if rl, ok := rotatelogsRegistry[logPath]; ok {
		return rl, nil
	}

	opts := []rotatelogs.Option{
		rotatelogs.WithRotationTime(time.Hour * 24),
	}
	// 按天保留（MaxAge 天），MaxAge 与 MaxCount 不能同时设置，优先 MaxAge
	if config.MaxAge > 0 {
		opts = append(opts, rotatelogs.WithMaxAge(time.Hour*24*time.Duration(config.MaxAge)))
	} else if config.MaxCount > 0 {
		opts = append(opts, rotatelogs.WithRotationCount(uint(config.MaxCount)))
	}
	// 单日文件超过 maxSize 时额外轮转（生成 .1 .2 等序号后缀）
	if config.MaxSize > 0 {
		opts = append(opts, rotatelogs.WithRotationSize(int64(config.MaxSize)*1024*1024))
	}

	rl, err := rotatelogs.New(logPath+".%Y-%m-%d", opts...)
	if err != nil {
		return nil, err
	}
	rotatelogsRegistry[logPath] = rl
	return rl, nil
}

// filepath : zap config file path
// section  : Config options
func NewZapLogger(config config.LoggingConfig, logType string) *zap.Logger {
	// 获取日志路径
	logPath := getLogPath(config, logType)

	// 确保日志目录存在（只在初始化时执行一次）
	if err := ensureLogDir(logPath); err != nil {
		panic("failed to create log directory: " + err.Error())
	}

	rotLogWriter, err := sharedRotatelogs(config, logPath)
	if err != nil {
		panic("failed to create log writer: " + err.Error())
	}

	return zap.New(
		zapcore.NewCore(
			func() zapcore.Encoder {
				encoderConfig := zapcore.EncoderConfig{
					MessageKey:     "msg",                                          // 日志消息字段名
					LevelKey:       "level",                                        // 日志级别字段名
					TimeKey:        "time",                                         // 时间字段名
					NameKey:        "logger",                                       // 日志器名称字段名
					CallerKey:      "file",                                         // 调用者文件字段名
					FunctionKey:    "func",                                         // 函数名字段名
					StacktraceKey:  "stacktrace",                                   // 堆栈跟踪字段名
					LineEnding:     zapcore.DefaultLineEnding,                      // 行结束符
					EncodeLevel:    zapcore.CapitalLevelEncoder,                    // 日志级别编码 ：  大写级别（INFO、ERROR等）
					EncodeTime:     zapcore.TimeEncoderOfLayout(config.TimeFormat), // 时间编码 ： 自定义时间格式
					EncodeDuration: zapcore.SecondsDurationEncoder,                 // 持续时间编码 ：  以秒为单位
					EncodeCaller:   zapcore.ShortCallerEncoder,                     // 调用者编码 ：短文件名格式（如 main.go:123 ）
				}
				if config.LevelColor {
					encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
				}
				if config.Encoder == "json" {
					return zapcore.NewJSONEncoder(encoderConfig)
				}
				return zapcore.NewConsoleEncoder(encoderConfig)
			}(),

			// Rotatelogs 策略
			// - 按天轮转 ：每天创建新文件，文件名包含日期（如 system.log.2026-08-07）
			// - 按时间清理 ：保留 max-age 天的日志文件
			// - 单日超过 max-size MB 时额外轮转（序号后缀）
			zapcore.NewMultiWriteSyncer(
				func() (writer []zapcore.WriteSyncer) {
					writer = append(writer, zapcore.AddSync(rotLogWriter))
					if config.Stdout {
						writer = append(writer, zapcore.AddSync(os.Stdout))
					}
					return writer
				}()...,
			),
			zapcore.Level(config.Level),
		),

		zap.AddStacktrace(zapcore.Level(config.StackTraceLevel)),

		zap.WithCaller(config.Caller),
	)

}

// NewSQLFileWriter 创建用于 GORM SQL 日志输出的文件写入器
// 按日期轮转，文件名包含日期（如 sql.log.2026-08-07）
// 调用方无需关闭该 writer（rotatelogs 会在每次写入后自动处理轮转）
func NewSQLFileWriter(config config.LoggingConfig) io.Writer {
	logPath := getLogPath(config, LogTypeSQL)
	if logPath == "" {
		return io.Discard
	}
	if err := ensureLogDir(logPath); err != nil {
		return io.Discard
	}
	rl, err := sharedRotatelogs(config, logPath)
	if err != nil {
		return io.Discard
	}
	return rl
}
