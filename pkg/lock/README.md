# Lock 包

`pkg/lock` 提供了一个通用的分布式锁接口。它旨在抽象底层的锁提供者（例如 Redis、InMemory），并为整个应用程序中的锁获取和释放提供一致的 API。

## 特性

- **通用接口**：`MutexManager` 和 `Mutex` 接口允许可插拔的后端。
- **InMemory 提供者**：内置内存实现，适用于开发和测试。
- **可配置**：通过函数选项进行广泛的配置（过期时间、重试次数、重试延迟、命名空间等）。
- **上下文支持**：`Lock` 和 `Unlock` 操作支持 `context.Context`，用于取消和超时控制。
- **自定义错误**：用于精确错误处理的类型化错误（例如 `ErrorKindLockHeld`，`ErrorKindCannotLock`）。

## 用法

### 初始化

使用配置初始化 `MutexManager`。目前，提供了 `InMemory` 提供者。

```go
import (
    "time"
    "github.com/easysoft/gitfox-community/pkg/lock"
)

func main() {
    config := lock.Config{
        App:        "gitfox",
        Provider:   lock.MemoryProvider,
        Expiry:     10 * time.Second,
        Tries:      3,
        RetryDelay: 500 * time.Millisecond,
    }
    
    // 创建一个 InMemory MutexManager
    manager := lock.NewInMemory(config)
}
```

### 获取锁

使用 `NewMutex` 方法创建一个互斥锁实例，然后使用 `Lock` 来获取它。

```go
func criticalSection(manager lock.MutexManager) error {
    ctx := context.Background()
    key := "resource-123"

    // 使用选项创建互斥锁
    mutex, err := manager.NewMutex(key, 
        lock.WithNamespace("my-namespace"),
        lock.WithExpiry(5 * time.Second),
    )
    if err != nil {
        return err
    }

    // 获取锁
    if err := mutex.Lock(ctx); err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    
    // 确保锁被释放
    defer func() {
        if err := mutex.Unlock(ctx); err != nil {
            fmt.Printf("failed to unlock: %v\n", err)
        }
    }()

    // ... 执行临界区操作 ...
    
    return nil
}
```

## 配置选项

您可以使用传递给 `NewMutex` 的 `lock.Option` 函数来自定义互斥锁的行为：

- `WithNamespace(ns string)`: 设置锁键的命名空间。
- `WithExpiry(d time.Duration)`: 设置锁的过期时间。
- `WithTries(t int)`: 设置重试尝试的次数。
- `WithRetryDelay(d time.Duration)`: 设置重试之间的固定延迟。
- `WithRetryDelayFunc(f DelayFunc)`: 设置自定义的重试延迟函数。
- `WithValue(v string)`: 手动设置锁的令牌值（用于锁所有权转移）。

## 错误处理

该包提供了特定的错误类型以帮助识别失败原因：

- `ErrorKindLockHeld`: 锁已被持有。
- `ErrorKindCannotLock`: 超时或无法获取锁。
- `ErrorKindMaxRetriesExceeded`: 重试次数耗尽。
- `ErrorKindContext`: 上下文取消或截止时间已过。

