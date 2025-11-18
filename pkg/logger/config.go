package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

const (
	LogTypeAPIServer = "apiserver"
	LogTypeRunner    = "runner"
	LogTypeSSH       = "ssh"
	LogTypeScan      = "scan"
	LogTypeSystem    = "system"
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

// filepath : zap config file path
// section  : Config options
func NewZapLogger(config config.LoggingConfig, logType string) *zap.Logger {
	// 获取日志路径
	logPath := getLogPath(config, logType)

	// 确保日志目录存在（只在初始化时执行一次）
	if err := ensureLogDir(logPath); err != nil {
		panic("failed to create log directory: " + err.Error())
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

			// Lumberjack 策略（当前启用）
			// - 按文件大小轮转 ：当文件达到 max-size MB 时创建新文件
			// - 保留文件数量 ：最多保留 max-count 个备份文件
			// - 按时间清理 ：保留 max-age 天的日志文件
			// - 文件命名 ：使用序号后缀（如 app.log , app.log.1 , app.log.2 ）
			// - 压缩 ：不压缩旧文件（ Compress: false ）
			// - 时区 ：使用本地时间（ LocalTime: true ）
			zapcore.NewMultiWriteSyncer(
				func() (writer []zapcore.WriteSyncer) {
					writer = append(writer,
						zapcore.AddSync(
							&lumberjack.Logger{
								Filename:   logPath,
								MaxSize:    config.MaxSize,
								MaxBackups: config.MaxCount,
								MaxAge:     config.MaxAge,
								Compress:   false,
								LocalTime:  true,
							}),
					)
					if config.Stdout {
						writer = append(writer, zapcore.AddSync(os.Stdout))
					}
					return writer
				}()...,
			),

			// Rotatelogs 策略（已禁用）
			// - 按时间轮转 ：每天创建新文件
			// - 保留文件数量 ：最多保留 max-count 个备份文件
			// - 按时间清理 ：保留 max-age 天的日志文件
			// - 文件命名 ：使用日期后缀（如 app.log.2023-01-01 ）
			// - 压缩 ：不压缩旧文件（ Compress: false ）
			// - 时区 ：使用本地时间（ LocalTime: true ）

			// zapcore.NewMultiWriteSyncer(
			// 	func() []zapcore.WriteSyncer {
			// 		hook, err := rotatelogs.New(
			// 			logSec.GetString("file-name")+".%Y-%m-%d",
			// 			rotatelogs.WithLinkName(logSec.GetString("file-name")),
			// 			rotatelogs.WithMaxAge(time.Hour*24*time.Duration(logSec.GetInt("max-age"))),
			// 			rotatelogs.WithRotationTime(time.Hour*24),
			// 			rotatelogs.WithRotationSize(logSec.GetInt64("max-size")*1024*1024),
			// 			//rotatelogs.WithRotationCount(logSec.GetUint("max-count")),
			// 		)
			// 		if err != nil {
			// 			panic(hook)
			// 		}
			// 		writes := []zapcore.WriteSyncer{zapcore.AddSync(hook)}
			// 		if logSec.GetBool("stdout") {
			// 			writes = append(writes, zapcore.AddSync(os.Stdout))
			// 		}
			// 		return writes
			// 	}()...,
			// ),
			zapcore.Level(config.Level),
		),

		zap.AddStacktrace(zapcore.Level(config.StackTraceLevel)),

		zap.WithCaller(config.Caller),
	)

}
