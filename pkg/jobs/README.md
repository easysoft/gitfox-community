# pkg/jobs - 任务执行器核心包

## 概述

`pkg/jobs` 包提供了一个轻量级、安全的任务执行框架，专注于任务的**执行机制**和**调度时机控制**。它是 GitFox 任务系统的底层核心，负责：

- 任务处理器（Handler）的注册与管理
- 任务的安全执行与异常捕获
- 智能定时器调度
- 任务状态变更的发布

## 核心组件

### 1. Executor - 任务执行器

#### 职责
- 管理所有任务类型的处理器（Handler）
- 提供任务执行的统一入口
- 保证执行过程的安全性（panic 恢复）

#### 实现原理

```go
type Executor struct {
    handlerMap      map[string]Handler  // 任务类型 -> 处理器映射
    handlerComplete bool                // 注册是否已完成
    publisher       pubsub.Publisher    // 事件发布器
}
```

**关键设计**：

1. **处理器注册阶段控制**
   - 使用 `handlerComplete` 标志区分注册阶段和运行阶段
   - 一旦调用 `FinishRegistration()`，禁止再注册新的处理器
   - 防止运行时动态注册导致的并发问题

2. **类型安全**
   - 每个任务类型只能注册一个处理器
   - 注册时检查 jobType 非空、Handler 非 nil
   - 防止重复注册同一类型

3. **执行安全**
   ```go
   func (e *Executor) Exec(ctx context.Context, jobName, jobType string,
                           input string, reportFn ProgressReporter) (result string, err error) {
       defer func() {
           if r := recover(); r != nil {
               // 捕获 panic，转换为 error 返回
               err = errors.Newf("panic while processing job=%s type=%s: %v\n%s",
                   jobName, jobType, r, debug.Stack())
           }
       }()

       exec, ok := e.handlerMap[jobType]
       if !ok {
           return "", errNoHandlerDefined
       }

       return exec.Handle(ctx, input, reportFn)
   }
   ```

#### 安全保障机制

| 安全措施 | 保护目标 | 实现方式 |
|---------|---------|---------|
| Panic 恢复 | 防止单个任务崩溃导致整个系统崩溃 | `defer recover()` + 堆栈追踪 |
| 类型检查 | 防止执行未注册的任务 | 执行前检查 `handlerMap` |
| 注册锁定 | 防止运行时注册导致数据竞争 | `handlerComplete` 标志 |
| 上下文传递 | 支持超时和取消 | 所有 Handler 接收 `context.Context` |

### 2. Handler - 任务处理器接口

```go
type Handler interface {
    Handle(ctx context.Context, input string, reportFn ProgressReporter) (result string, err error)
}
```

#### 设计原则

1. **上下文感知**
   - 接收 `context.Context` 参数
   - 必须响应 `ctx.Done()` 信号，支持优雅取消
   - 超时控制由调用方（Scheduler）统一管理

2. **进度报告**
   - 可选使用 `ProgressReporter` 回调
   - 支持长时间运行的任务实时报告进度（0-100）
   - 允许中途更新任务结果

3. **输入/输出序列化**
   - `input` 和 `result` 均为字符串
   - 业务数据需自行序列化（通常使用 JSON）
   - 保持接口简洁通用

#### 实现建议

```go
type MyJobHandler struct {
    // dependencies
}

func (h *MyJobHandler) Handle(ctx context.Context, input string,
                               reportFn ProgressReporter) (string, error) {
    // 1. 解析输入
    var params MyJobParams
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }

    // 2. 执行任务，定期检查上下文
    for i, step := range params.Steps {
        select {
        case <-ctx.Done():
            return "", ctx.Err()  // 优雅退出
        default:
        }

        // 执行步骤
        if err := processStep(step); err != nil {
            return "", err
        }

        // 报告进度
        progress := (i + 1) * 100 / len(params.Steps)
        reportFn(progress, fmt.Sprintf("Completed step %d/%d", i+1, len(params.Steps)))
    }

    // 3. 返回结果
    result := MyJobResult{Status: "success"}
    data, _ := json.Marshal(result)
    return string(data), nil
}
```

### 3. SchedulerTimer - 智能定时器

#### 设计目标

实现一个**自适应的、可抢占的**调度定时器，平衡实时性与资源效率。

#### 核心概念

**Edgy（边缘触发）模式**：
- 当定时器处于 `edgy=true` 状态时，表示调度器处于"空闲但随时准备工作"状态
- 此时若有新任务插入（调用 `RescheduleEarlier`），定时器会**立即触发**
- 避免新任务等待当前定时器周期结束

#### 工作原理

```
状态机：

[启动] → [等待最早任务时间]
           ↓
        [定时器触发] → [处理任务]
           ↑              ↓
           └──[更多任务]─┘

[无任务] → [edgy=true，最长等待30分钟]
              ↓
           [新任务插入] → [立即触发]
```

#### 关键方法

1. **ResetAt(next time.Time, edgy bool)**
   - 重置定时器到指定时间
   - 自动限制在 [1ns, 30min] 范围内
   - 设置 edgy 状态

2. **RescheduleEarlier(next time.Time)**
   - 仅当新时间**更早**时才重置定时器
   - 如果 `edgy=true`，立即触发
   - 零时间（`time.Time{}`）仅在 edgy 时触发

3. **Stop()**
   - 安全停止定时器
   - **清空通道残留事件**，防止误触发

#### 时间约束

```go
const timerMaxDur = 30 * time.Minute  // 最大等待时间，防止永久休眠
const timerMinDur = time.Nanosecond   // 最小间隔，实现立即触发
```

#### 使用场景示例

```go
timer := NewSchedulerTimer()

// 场景1: 常规调度
nextJob := time.Now().Add(10 * time.Minute)
timer.ResetAt(nextJob, false)  // 10分钟后触发

// 场景2: 无待处理任务
timer.ResetAt(time.Now().Add(30*time.Minute), true)  // edgy模式，最长等30分钟

// 场景3: 插入紧急任务
urgentJob := time.Now().Add(1 * time.Minute)
timer.RescheduleEarlier(urgentJob)  // 如果当前是edgy，立即触发

// 场景4: 等待定时器触发
select {
case now := <-timer.Ch():
    // 处理到期任务
case <-ctx.Done():
    timer.Stop()
}
```

## 执行流程图

```
┌─────────────────────────────────────────────────────────┐
│                   Scheduler (上层)                       │
│  - 从数据库获取待执行任务                                │
│  - 管理并发数量                                          │
│  - 使用分布式锁协调多实例                                │
└────────────────────┬────────────────────────────────────┘
                     │
                     │ Exec(ctx, jobName, jobType, input, reportFn)
                     ↓
┌─────────────────────────────────────────────────────────┐
│              Executor (执行器 - 本包)                    │
│                                                          │
│  1. 查找 Handler                                         │
│  2. Panic 恢复保护                                       │
│  3. 调用 exec.Handle(...)                               │
│                                                          │
└────────────────────┬────────────────────────────────────┘
                     │
                     │ Handle(ctx, input, reportFn)
                     ↓
┌─────────────────────────────────────────────────────────┐
│         Handler 实现 (业务层注册)                        │
│                                                          │
│  - 解析 input (JSON)                                    │
│  - 执行具体业务逻辑                                      │
│  - 定期检查 ctx.Done()                                  │
│  - 调用 reportFn() 报告进度                             │
│  - 返回 result (JSON) 或 error                          │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 安全性设计

### 1. 异常隔离

**问题**：单个任务的 panic 不应影响整个系统

**解决方案**：
```go
defer func() {
    if r := recover(); r != nil {
        // 捕获 panic，记录完整堆栈
        err = errors.Newf("panic: %v\n%s", r, debug.Stack())
    }
}()
```

**效果**：
- 任务 panic 转换为 error 返回
- 调度器可以继续执行其他任务
- 完整堆栈信息帮助调试

### 2. 上下文传播

**问题**：如何优雅地终止长时间运行的任务

**解决方案**：
- 所有 Handler 接收 `context.Context`
- 调度器在超时前取消上下文
- Handler 定期检查 `ctx.Done()`

**最佳实践**：
```go
for {
    select {
    case <-ctx.Done():
        return "", ctx.Err()  // 立即退出
    default:
    }

    // 执行一小段工作
}
```

### 3. 资源清理

**定时器清理**：
```go
func (t *schedulerTimer) Stop() {
    t.timer.Stop()

    // 关键：清空通道残留事件
    select {
    case <-t.timer.C:
    default:
    }
}
```

**为什么重要**：
- Go 的 `Timer.Stop()` 不保证清空通道
- 残留事件会导致下次 `Reset` 后立即触发
- 必须手动 drain channel

### 4. 并发安全

**注册阶段保护**：
- `handlerMap` 仅在注册阶段修改
- `FinishRegistration()` 后变为只读
- 无需锁，零运行时开销

**执行阶段**：
- `handlerMap` 只读访问，天然线程安全
- 多个 goroutine 可并发调用 `Exec()`

## 性能特性

### 1. 零分配执行路径

```go
// 执行任务不分配额外内存（除了业务逻辑本身）
func (e *Executor) Exec(...) (string, error) {
    exec := e.handlerMap[jobType]  // map 查找，无内存分配
    return exec.Handle(...)         // 接口调用，无装箱
}
```

### 2. 智能调度减少 CPU 消耗

- 无任务时最长休眠 30 分钟
- 避免无意义的轮询数据库
- 新任务插入时立即唤醒（edgy 机制）

### 3. 定时器复用

```go
// 复用同一个 Timer 对象
t.timer.Reset(dur)  // 而非每次创建新 Timer
```

## 测试建议

### 单元测试模板

```go
func TestMyHandler(t *testing.T) {
    handler := &MyJobHandler{}

    // 准备输入
    input, _ := json.Marshal(MyJobParams{Value: 42})

    // 执行
    ctx := context.Background()
    result, err := handler.Handle(ctx, string(input), nil)

    // 验证
    assert.NoError(t, err)
    assert.Contains(t, result, "success")
}

func TestHandlerTimeout(t *testing.T) {
    handler := &SlowJobHandler{}

    // 设置超时上下文
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    // 执行应超时
    _, err := handler.Handle(ctx, "", nil)
    assert.Equal(t, context.DeadlineExceeded, err)
}
```

### 集成测试要点

1. **测试 Executor 注册逻辑**
   - 重复注册应返回错误
   - 注册完成后禁止新注册

2. **测试 Panic 恢复**
   - Handler panic 不应导致测试崩溃
   - error 应包含堆栈信息

3. **测试定时器行为**
   - Edgy 模式立即触发
   - 早期任务抢占晚期任务
   - 停止清空通道

## 总结

| 特性 | 实现方式 | 好处 |
|-----|---------|-----|
| 安全执行 | Panic 恢复 + 堆栈追踪 | 单个任务失败不影响系统 |
| 优雅终止 | Context 传播 | 支持超时和取消 |
| 智能调度 | Edgy 定时器 | 平衡实时性和资源消耗 |
| 类型安全 | Handler 注册检查 | 编译期和运行期双重保障 |
| 零开销 | 无锁只读 map | 高并发执行无瓶颈 |

**设计哲学**：

> 提供**机制**而非**策略**：本包只关心"如何执行任务"，而"什么时候执行哪些任务"由上层 Scheduler 决定。

**下一步**：查看 [app/service/jobs/README.md](../../app/service/jobs/README.md) 了解如何使用 Scheduler 调度任务。
