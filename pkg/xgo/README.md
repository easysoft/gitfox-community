
# XGo - 安全的 Goroutine 管理包

XGo 是一个为 GitFox 项目设计的 Goroutine 管理包，提供了安全的并发执行、panic 恢复、重试机制和清理功能。

## 特性

- 🛡️ **安全执行**: 自动捕获和恢复 panic，防止程序崩溃
- 🔄 **重试机制**: 支持多种退避策略的自动重试
- 🧹 **清理功能**: 支持资源清理和错误通知
- 📊 **高性能**: 轻量级设计，平均执行时间 < 1μs
- 🔍 **可观测性**: 自动记录错误和堆栈信息

## 快速开始

### 基本使用

```go
package main

import (
    "fmt"
    "time"
    "github.com/easysoft/gitfox-community/pkg/xgo"
)

func main() {
    // 基本的安全 goroutine 执行
    xgo.Go(func() {
        fmt.Println("Hello from safe goroutine!")
    }, func(err error) {
        fmt.Printf("Error occurred: %v\n", err)
    })
    
    // 等待 goroutine 完成
    time.Sleep(100 * time.Millisecond)
}
```

### 带清理函数的执行

```go
func main() {
    var resource *SomeResource
    
    xgo.GoWithCleaner(func() {
        resource = acquireResource()
        // 执行一些可能 panic 的操作
        doSomethingRisky(resource)
    }, func(err error) {
        fmt.Printf("Task failed: %v\n", err)
    }, func() {
        // 清理资源
        if resource != nil {
            resource.Close()
        }
    })
}
```

### 重试机制

```go
func main() {
    // 带重试的网络请求
    xgo.GoWithRetries(func() error {
        return makeNetworkRequest()
    }, func(err error) {
        fmt.Printf("All retries failed: %v\n", err)
    }, 
        xgo.WithTimes(5),                     // 重试 5 次
        xgo.WithBackoff(xgo.LinearBackoff),  // 线性退避
    )
}
```

## API 文档

### 核心函数

#### `Go(fn func(), notifyFunc func(error))`

启动一个安全的 goroutine，自动处理 panic。

**参数:**

- `fn`: 要执行的函数
- `notifyFunc`: panic 时的通知函数（可为 nil）

#### `GoWithCleaner(fn func(), notifyFunc func(error), cleaner ...func())`

启动带清理函数的安全 goroutine。

**参数:**

- `fn`: 要执行的函数
- `notifyFunc`: panic 时的通知函数（可为 nil）
- `cleaner`: 清理函数列表，无论成功或失败都会执行

#### `GoWithRetries(fn func() error, notifyFunc func(error), opts ...RetryOption)`

启动带重试机制的 goroutine。

**参数:**

- `fn`: 返回 error 的函数，nil 表示成功
- `notifyFunc`: 最终失败时的通知函数
- `opts`: 重试选项

### 重试选项

#### `WithTimes(times int) RetryOption`

设置重试次数（默认: 5）

#### `WithBackoff(mode RetryMode) RetryOption`

设置退避策略：

- `None`: 无退避
- `LinearBackoff`: 线性退避（固定间隔）
- `RandomBackoff`: 随机退避
- `ExponentialBackoff`: 指数退避（暂未实现）

### 工具函数

#### `Recover(notifyFunc func(error), cleanups ...func())`

手动 panic 恢复函数，可在 defer 中使用。

## 使用示例

### 1. 数据库操作

```go
func processUserData(userID int) {
    var db *DBConnection
    xgo.GoWithCleaner(func() {
        db = connectDB()
        user := db.GetUser(userID)
        processUser(user)
    }, func(err error) {
        log.Printf("Failed to process user %d: %v", userID, err)
    }, func() {
        // 确保数据库连接被关闭
        if db != nil {
            db.Close()
        }
    })
}
```

### 2. 网络请求重试

```go
func fetchDataWithRetry(url string) {
    xgo.GoWithRetries(func() error {
        resp, err := http.Get(url)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode != 200 {
            return fmt.Errorf("HTTP %d", resp.StatusCode)
        }
        
        return processResponse(resp)
    }, func(err error) {
        log.Printf("Failed to fetch %s after retries: %v", url, err)
    }, 
        xgo.WithTimes(3),
        xgo.WithBackoff(xgo.RandomBackoff),
    )
}
```

### 3. 批量任务处理

```go
func processBatch(items []Item) {
    for _, item := range items {
        item := item // 避免闭包捕获问题
        xgo.Go(func() {
            if err := processItem(item); err != nil {
                panic(fmt.Sprintf("Failed to process item %d: %v", item.ID, err))
            }
        }, func(err error) {
            log.Printf("Item processing failed: %v", err)
            // 记录失败的项目以便后续处理
            recordFailedItem(item)
        })
    }
}
```

## 性能报告

基于 Intel(R) Core(TM) i3-8100T CPU @ 3.10GHz 的基准测试结果：

### 基准测试结果

| 函数            | 执行次数  | 平均耗时    | 性能等级 |
|-----------------|-----------|-------------|----------|
| `Go`            | 1,902,727 | 628.2 ns/op | 🚀 优秀  |
| `GoWithRetries`  | 1,456,723 | 950.3 ns/op | ⚡ 良好  |

### 性能分析

#### `Go` 函数性能

- **平均耗时**: 628.2 纳秒
- **吞吐量**: ~1.59M ops/sec
- **内存开销**: 极低，仅 goroutine 创建成本
- **适用场景**: 高频异步任务

#### `GoWithRetries` 函数性能

- **平均耗时**: 950.3 纳秒
- **吞吐量**: ~1.05M ops/sec
- **额外开销**: 重试逻辑增加 ~51% 耗时
- **适用场景**: 需要可靠性的任务

### 性能对比

```
基准测试环境:
- OS: Linux (amd64)
- CPU: Intel i3-8100T @ 3.10GHz
- Go版本: 1.21+
- 测试时间: 4.859s
```

### 内存使用

XGo 包设计为零额外内存分配（除了 goroutine 本身），所有配置通过选项模式传递，避免了不必要的堆分配。

### 扩展性

- **并发安全**: 所有函数都是并发安全的
- **无状态设计**: 函数间无共享状态
- **线性扩展**: 性能随 CPU 核心数线性扩展

## 最佳实践

### 1. 错误处理

```go
// ✅ 推荐：提供错误处理函数
xgo.Go(riskyTask, func(err error) {
    log.Printf("Task failed: %v", err)
    metrics.IncrementErrorCount()
})

// ❌ 不推荐：忽略错误
xgo.Go(riskyTask, nil)
```

### 2. 资源管理

```go
// ✅ 推荐：使用清理函数
xgo.GoWithCleaner(func() {
    file := openFile()
    processFile(file)
}, handleError, func() {
    if file != nil {
        file.Close()
    }
})
```

### 3. 重试策略

```go
// ✅ 推荐：根据场景选择合适的重试策略
// 网络请求：随机退避
xgo.GoWithRetries(networkCall, handleError, 
    xgo.WithTimes(3), 
    xgo.WithBackoff(xgo.RandomBackoff))

// 数据库操作：线性退避
xgo.GoWithRetries(dbOperation, handleError,
    xgo.WithTimes(5),
    xgo.WithBackoff(xgo.LinearBackoff))
```

## 注意事项

1. **Goroutine 生命周期**: XGo 不管理 goroutine 的生命周期，请确保适当的同步
2. **错误传播**: panic 会被捕获并转换为 error，不会向上传播
3. **清理函数**: 清理函数总是会执行，无论任务成功或失败
4. **重试逻辑**: 只有返回 error 才会重试，panic 不会触发重试

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../../LICENSE) 文件了解详情。
