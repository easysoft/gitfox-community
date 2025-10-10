好的，这是为您整理好的 `README.md` 文档：

# GitFox Logger

GitFox Logger 是一个基于 Zap 的高性能异步日志库，专为高并发场景设计。它提供了同步和异步两种日志记录模式，支持结构化日志和语法糖日志记录方式。

## 特性

- 🚀 **高性能异步日志**：基于通道的异步写入，避免 I/O 阻塞
- 🔧 **灵活配置**：支持 YAML 配置文件，可配置日志级别、格式、轮转策略等
- 📝 **多种日志方式**：支持结构化日志和语法糖日志
- 🔄 **日志轮转**：支持按大小、时间、数量进行日志文件轮转
- 🎨 **多种输出格式**：支持 JSON 和 Console 两种编码格式
- 🛡️ **并发安全**：使用锁和通道保证线程安全
- ♻️ **内存优化**：通过对象池减少 GC 压力
- 🏷️ **命名空间**：支持 Logger 命名和字段继承
- 🔗 **链式调用**：支持方法链式调用

## 快速开始

### 1. 创建配置文件

创建一个 YAML 配置文件（例如 `config.yaml`）：

```yaml
app:
  level: -1                               # 日志级别：-1(debug), 0(info), 1(warn), 2(error), 3(dpanic), 4(panic), 5(fatal)
  stdout: true                            # 是否输出到控制台
  caller: true                            # 是否显示调用者信息
  encoder: console                        # 编码格式：json 或 console
  max-age: 30                             # 日志文件最大保存天数
  max-count: 10                           # 日志文件最大保存数量
  max-size: 100                           # 日志文件切割大小（MB）
  file-name: ./logs/app.log               # 日志文件路径
  level-color: true                       # 是否启用级别颜色（仅对控制台输出有效）
  time-format: "2006-01-02 15:04:05.999"  # 时间格式字符串
  stack-trace-level: 2                    # 堆栈跟踪级别
```

### 2. 基本使用

```go
package main

import (
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// 创建同步日志器
	log := logger.NewZapMateLogger("./config.yaml", "app")
	defer log.Sync()

	// 结构化日志
	log.Info("用户登录",
		zap.String("username", "john"),
		zap.Int("user_id", 123),
		zap.String("ip", "192.168.1.1"))

	// 创建异步日志器（推荐用于高并发场景）
	asyncLog := log.SetAsyncer(1000) // 设置通道缓冲区大小为1000
	defer asyncLog.Flush()

	asyncLog.Info("异步日志消息",
		zap.String("module", "auth"),
		zap.Duration("latency", time.Millisecond*100))
}
```

### 3. Sugar Logger（语法糖日志）

Sugar Logger 提供更便捷的日志记录方式：

```go
// 获取 Sugar Logger
sugar := log.Sugar()
defer sugar.Flush()

// 简单日志
sugar.Info("这是一条信息日志")
sugar.Error("发生了错误")

// 格式化日志
sugar.Infof("用户 %s 登录成功", "john")
sugar.Errorf("连接数据库失败: %v", err)

// 结构化日志（键值对）
sugar.Infow("用户操作",
    "action", "login",
    "username", "john",
    "timestamp", time.Now(),
    "success", true)

// 转换回结构化日志器
structured := sugar.Desugar()
```

## 高级用法

### 1. Logger 命名和字段继承

```go
// 创建命名 Logger
userLogger := log.Named("user-service")
authLogger := log.Named("auth-service")

// 添加持久化字段
logWithFields := userLogger.With(
    zap.String("service", "user"),
    zap.String("version", "v1.0"),
    zap.String("instance_id", "inst-123"),
)

// 所有后续日志都会包含这些字段
logWithFields.Info("用户创建成功", zap.String("user_id", "user-456"))
logWithFields.Error("用户创建失败", zap.Error(err))

// Sugar 模式的命名和字段
sugarWithContext := log.Sugar().Named("payment").With(
    "service", "payment",
    "version", "v2.0",
)
sugarWithContext.Infow("支付处理", "order_id", "order-789", "amount", 99.99)
```

### 2. 不同日志级别的使用

```go
// 调试信息（开发环境）
log.Debug("调试信息", zap.Any("debug_data", debugInfo))

// 一般信息
log.Info("服务启动", zap.Int("port", 8080))

// 警告信息
log.Warn("连接池使用率过高", zap.Float64("usage", 0.85))

// 错误信息
log.Error("数据库连接失败",
    zap.Error(err),
    zap.String("database", "mysql"),
    zap.Int("retry_count", 3))

// 开发环境 Panic（不会终止程序）
log.DPanic("开发环境检测到严重问题", zap.String("issue", "memory_leak"))

// Panic（会终止程序）
log.Panic("无法恢复的错误", zap.Error(criticalErr))

// Fatal（会终止程序）
log.Fatal("致命错误", zap.Error(fatalErr))
```

### 3. 异步日志最佳实践

```go
// 创建异步日志器
asyncLog := logger.NewZapMateLogger("./config.yaml", "app").SetAsyncer(2000)

// 在高并发场景中使用
func handleRequest(w http.ResponseWriter, r *http.Request) {
    start := time.Now()

    // 异步记录请求开始
    asyncLog.Info("请求开始",
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
        zap.String("remote_addr", r.RemoteAddr))

    // 处理业务逻辑...

    // 异步记录请求结束
    asyncLog.Info("请求完成",
        zap.String("method", r.Method),
        zap.String("path", r.URL.Path),
        zap.Duration("duration", time.Since(start)))
}

// 程序退出前确保所有日志写入
func gracefulShutdown() {
    asyncLog.Flush() // 等待所有异步日志写入完成
}
```

### 4. 错误处理和恢复

```go
func businessLogic() {
    defer func() {
        if r := recover(); r != nil {
            log.Error("程序 panic 恢复",
                zap.Any("panic", r),
                zap.Stack("stack"))
        }
    }()

    // 业务逻辑...
    if err := someOperation(); err != nil {
        log.Error("操作失败",
            zap.Error(err),
            zap.String("operation", "someOperation"),
            zap.String("context", "business_logic"))
        return
    }
}
```

## 配置参数详解

| 参数                | 类型     | 默认值                     | 说明                                                                     |
| ------------------- | -------- | -------------------------- | ------------------------------------------------------------------------ |
| `level`             | `int`    | `0`                        | 日志级别：-1(debug), 0(info), 1(warn), 2(error), 3(dpanic), 4(panic), 5(fatal) |
| `stdout`            | `bool`   | `false`                    | 是否同时输出到控制台                                                     |
| `caller`            | `bool`   | `false`                    | 是否显示调用者文件和行号信息                                             |
| `encoder`           | `string` | `console`                  | 编码格式：`json` 或 `console`                                               |
| `max-age`           | `int`    | `0`                        | 日志文件最大保存天数，0 表示不删除                                       |
| `max-count`         | `int`    | `0`                        | 日志文件最大保存数量，0 表示不限制                                       |
| `max-size`          | `int`    | `100`                      | 日志文件切割大小（MB）                                                   |
| `file-name`         | `string` | `./logs/app.log`           | 日志文件保存路径                                                         |
| `level-color`       | `bool`   | `false`                    | 是否启用日志级别颜色（仅对 console 编码器和控制台输出有效）              |
| `time-format`       | `string` | `2006-01-02 15:04:05.999`  | 时间格式，使用 Go 的时间格式                                             |
| `stack-trace-level` | `int`    | `2`                        | 打印堆栈跟踪的最低级别                                                   |

## API 参考

### ZapMateLogger 方法

```go
// 创建 Logger
func NewZapMateLogger(filepath, section string) *ZapMateLogger

// 设置异步模式
func (log *ZapMateLogger) SetAsyncer(chanLen uint) *ZapMateLogger

// 添加命名空间
func (log *ZapMateLogger) Named(s string) *ZapMateLogger

// 添加字段
func (log *ZapMateLogger) With(fields ...zap.Field) *ZapMateLogger

// 转换为 Sugar Logger
func (log *ZapMateLogger) Sugar() *SugaredMateLogger

// 日志记录方法
func (log *ZapMateLogger) Debug(msg string, fields ...zap.Field)
func (log *ZapMateLogger) Info(msg string, fields ...zap.Field)
func (log *ZapMateLogger) Warn(msg string, fields ...zap.Field)
func (log *ZapMateLogger) Error(msg string, fields ...zap.Field)
func (log *ZapMateLogger) DPanic(msg string, fields ...zap.Field)
func (log *ZapMateLogger) Panic(msg string, fields ...zap.Field)
func (log *ZapMateLogger) Fatal(msg string, fields ...zap.Field)

// 刷新缓冲区
func (log *ZapMateLogger) Flush() error
```

### SugaredMateLogger 方法

```go
// 转换回结构化 Logger
func (s *SugaredMateLogger) Desugar() *ZapMateLogger

// 添加命名空间
func (s *SugaredMateLogger) Named(name string) *SugaredMateLogger

// 添加字段
func (s *SugaredMateLogger) With(args ...interface{}) *SugaredMateLogger

// 简单日志记录
func (s *SugaredMateLogger) Debug(args ...interface{})
func (s *SugaredMateLogger) Info(args ...interface{})
func (s *SugaredMateLogger) Warn(args ...interface{})
func (s *SugaredMateLogger) Error(args ...interface{})

// 格式化日志记录
func (s *SugaredMateLogger) Debugf(template string, args ...interface{})
func (s *SugaredMateLogger) Infof(template string, args ...interface{})
func (s *SugaredMateLogger) Warnf(template string, args ...interface{})
func (s *SugaredMateLogger) Errorf(template string, args ...interface{})

// 键值对日志记录
func (s *SugaredMateLogger) Debugw(msg string, keysAndValues ...interface{})
func (s *SugaredMateLogger) Infow(msg string, keysAndValues ...interface{})
func (s *SugaredMateLogger) Warnw(msg string, keysAndValues ...interface{})
func (s *SugaredMateLogger) Errorw(msg string, keysAndValues ...interface{})

// 刷新缓冲区
func (s *SugaredMateLogger) Flush() error
```

## 异步日志的工作原理

1.  **生产者**：主线程将日志条目放入 entryChan 通道
2.  **消费者**：异步协程从通道中取出日志条目并写入
3.  **对象复用**：使用 sync.Pool 复用 logEntry 对象，减少 GC 压力
4.  **同步控制**：使用 sync.WaitGroup 确保日志写入完成

```
主线程                    异步协程
  |                         |
  | 创建日志条目              |
  |                         |
  | 放入通道 ──────────────→ | 从通道取出
  |                         |
  | 继续执行业务逻辑          | 写入日志文件
  |                         |
  | Flush() ──────────────→ | 等待写入完成
```

## 性能优化建议

### 1. 异步 vs 同步选择

-   **高并发场景**：使用异步模式，设置合适的缓冲区大小（建议 1000-5000）
-   **低并发场景**：使用同步模式，减少内存占用
-   **关键日志**：使用同步模式，确保日志立即写入

### 2. 缓冲区大小设置

```go
// 根据 QPS 设置缓冲区大小
// QPS < 1000: chanLen = 1000
// QPS 1000-10000: chanLen = 2000-5000
// QPS > 10000: chanLen = 5000-10000

asyncLog := log.SetAsyncer(2000) // 适合中等并发场景
```

### 3. 字段复用

```go
// 好的做法：复用 Logger 实例
serviceLogger := log.Named("user-service").With(
    zap.String("service", "user"),
    zap.String("version", "v1.0"),
)

// 避免：每次都创建新的字段
// log.Info("message", zap.String("service", "user")) // 不推荐
```

### 4. 日志级别控制

```go
// 生产环境建议设置为 info 级别
// 开发环境可以设置为 debug 级别

// 配置文件中：
// level: 0  # info 级别，过滤掉 debug 日志
```

## 最佳实践

### 1. 结构化日志字段命名

```go
// 推荐的字段命名规范
log.Info("用户操作",
    zap.String("user_id", "123"),           // 使用下划线分隔
    zap.String("action", "login"),          // 动作类型
    zap.String("resource", "user"),         // 资源类型
    zap.Duration("duration", elapsed),      // 耗时
    zap.String("client_ip", clientIP),      // 客户端信息
    zap.Int("status_code", 200),           // 状态码
    zap.Error(err),                        // 错误信息
)
```

### 2. 错误日志记录

```go
// 记录错误时包含足够的上下文信息
if err := userService.CreateUser(user); err != nil {
    log.Error("创建用户失败",
        zap.Error(err),                    // 原始错误
        zap.String("user_id", user.ID),    // 用户ID
        zap.String("username", user.Name), // 用户名
        zap.String("operation", "create_user"), // 操作类型
        zap.String("service", "user-service"),  // 服务名
    )
    return err
}
```

### 3. 性能敏感场景

```go
// 在性能敏感的场景中，先检查日志级别
if log.Core().Enabled(zap.DebugLevel) {
    expensiveData := generateExpensiveDebugData()
    log.Debug("调试信息", zap.Any("data", expensiveData))
}
```

### 4. 程序退出处理

```go
func main() {
    log := logger.NewZapMateLogger("./config.yaml", "app").SetAsyncer(1000)

    // 设置优雅退出
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-c
        log.Info("收到退出信号，正在关闭...")
        log.Flush() // 确保所有日志写入
        os.Exit(0)
    }()

    // 业务逻辑...
}
```

## 故障排查

### 1. 日志丢失问题

-   检查异步缓冲区是否溢出
-   确保程序退出前调用 `Flush()`
-   检查磁盘空间是否充足

### 2. 性能问题

-   减少日志输出频率
-   调整缓冲区大小
-   使用合适的日志级别
-   避免在日志中使用复杂的数据结构

### 3. 配置问题

-   检查配置文件路径和格式
-   验证日志文件路径权限
-   确认时间格式设置正确

## 示例项目

查看 `pkg/demo/logger/` 目录下的完整示例：

```bash
cd pkg/demo/logger
go run main.go
```

## 运行测试

运行 GitFox Logger 的测试用例：

```bash
# 运行所有测试
go test ./pkg/logger

# 运行测试并显示详细输出
go test -v ./pkg/logger

# 运行测试并显示覆盖率
go test -cover ./pkg/logger

# 运行基准测试
go test -bench=. ./pkg/logger

# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./pkg/logger
go tool cover -html=coverage.out -o coverage.html
```

## 依赖

-   [go.uber.org/zap](https://github.com/uber-go/zap) - 高性能日志库
-   [gopkg.in/natefinch/lumberjack.v2](https://github.com/natefinch/lumberjack) - 日志轮转
-   [github.com/spf13/viper](https://github.com/spf13/viper) - 配置管理

## 许可证

本项目采用 MIT 许可证。