# Locker 服务

`app/service/locker` 是 GitFox 应用程序的特定领域锁定服务。它封装了通用的 `pkg/lock` 包，为仓库和拉取请求等应用程序实体提供方便、类型安全的锁定方法。

## 目的

此服务集中了关键领域操作的锁定逻辑，确保一致性并减少在共享资源（例如，更新默认分支、修改拉取请求）上进行并发操作时的竞争条件风险。

## 特性

- **领域特定 API**：提供像 `LockPR` 和 `LockDefaultBranch` 这样的方法，而不是通用的键锁定。
- **自动键管理**：处理锁键的构造（例如，按仓库 ID 命名空间）。
- **上下文感知日志记录**：自动将锁相关的元数据（键、命名空间、过期时间）添加到日志中。
- **安全性**：返回一个专用的 `unlockFn`，该函数在解锁阶段处理上下文超时，确保即使原始上下文被取消，锁也能被释放。

## 用法

### 初始化

`Locker` 需要来自 `pkg/lock` 的 `lock.MutexManager`。

```go
import (
    "github.com/easysoft/gitfox-community/app/service/locker"
    "github.com/easysoft/gitfox-community/pkg/lock"
)

func NewService(mtxManager lock.MutexManager) *MyService {
    return &MyService{
        locker: locker.NewLocker(mtxManager),
    }
}
```

### 锁定拉取请求

要锁定特定的拉取请求（或仓库的 pulls 命名空间）：

```go
func (s *MyService) UpdatePR(ctx context.Context, repoID, prNum int64) error {
    // 获取锁
    unlockFn, err := s.locker.LockPR(ctx, repoID, prNum, 10*time.Second)
    if err != nil {
        return err
    }
    // 推迟执行 locker 返回的解锁函数
    defer unlockFn()

    // ... 对 PR 执行操作 ...

    return nil
}
```

### 锁定默认分支更新

要在更新期间锁定仓库的默认分支：

```go
func (s *MyService) ChangeDefaultBranch(ctx context.Context, repoID int64, branchName string) error {
    // 获取锁
    unlockFn, err := s.locker.LockDefaultBranch(ctx, repoID, branchName, 5*time.Second)
    if err != nil {
        return err
    }
    defer unlockFn()

    // ... 更新默认分支 ...
    
    return nil
}
```

## 内部细节

- **命名空间**：使用 `"repo"` 作为底层锁的默认命名空间。
- **解锁行为**：返回的 `unlockFn` 创建一个新的带有超时的独立上下文，以确保即使父请求上下文被取消，`Unlock` 网络调用（如果适用）也能继续进行。