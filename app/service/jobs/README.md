# app/service/jobs - 任务调度服务

## 概述

`app/service/jobs` 包提供了一个**分布式、持久化、高可用**的任务调度服务。它基于 `pkg/jobs` 包提供的执行器，添加了：

- 任务的持久化存储（数据库）
- 分布式多实例协调（分布式锁）
- 任务生命周期管理（调度、执行、重试、过期）
- 定时任务支持（Cron 表达式）
- 任务状态实时同步（PubSub）

**适用场景**：
- 异步任务执行（如：发送邮件、生成报告）
- 定时任务调度（如：数据清理、健康检查）
- 长时间运行任务（如：代码扫描、构建任务）
- 分布式任务协调（多实例环境）

## 架构设计

### 整体架构

```
┌────────────────────────────────────────────────────────────┐
│                     API / Service Layer                     │
│  - RunJob(jobName, jobType, data, maxDuration)             │
│  - AddRecurringWithData(jobName, jobType, data, cron)      │
│  - CancelJob(jobName)                                       │
└────────────────────┬───────────────────────────────────────┘
                     │
                     ↓
┌────────────────────────────────────────────────────────────┐
│                      Scheduler (调度器)                     │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  主循环 (Run)                                        │  │
│  │  - 定时器触发 → 获取全局锁 → 加载就绪任务            │  │
│  │  - 检查并发数 → 启动任务 goroutine                   │  │
│  │  - 监听取消信号 (PubSub)                             │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐  │
│  │  任务执行流程                                        │  │
│  │  1. preExec:  更新状态为 running，设置超时时间       │  │
│  │  2. doExec:   调用 Executor.Exec() 执行              │  │
│  │  3. postExec: 更新状态，处理重试/重新调度            │  │
│  └─────────────────────────────────────────────────────┘  │
│                                                             │
└────────────────────┬───────────────────────────────────────┘
                     │
      ┌──────────────┼──────────────┐
      ↓              ↓               ↓
┌──────────┐  ┌────────────┐  ┌──────────┐
│ Database │  │   Lock     │  │  PubSub  │
│  (Jobs)  │  │(分布式锁)   │  │(任务取消) │
└──────────┘  └────────────┘  └──────────┘
```

### 核心组件

#### 1. Scheduler - 调度器

```go
type Scheduler struct {
    // 依赖项
    jobStore      jobsStore.Interface       // 任务持久化存储
    executor      *pkgJobs.Executor         // 任务执行器
    mxManager     lock.MutexManager         // 分布式锁管理器
    pubsubService pubsub.PubSub             // 发布订阅服务
    log           *logger.SugaredMateLogger // 独立的 Job 日志

    // 配置
    instanceID    string         // 实例 ID（区分多实例）
    maxRunning    int            // 最大并发任务数
    retentionTime time.Duration  // 任务保留时间

    // 运行时状态
    signal       chan time.Time                 // 调度信号通道
    done         chan struct{}                  // 停止信号
    wgRunning    sync.WaitGroup                 // 运行中任务计数
    cancelJobMx  sync.Mutex                     // 取消任务锁
    cancelJobMap map[string]context.CancelFunc  // 任务取消函数映射
}
```

## 核心流程

### 1. 任务生命周期

```
┌──────────┐
│ Scheduled│ ← 创建任务
└────┬─────┘
     │
     │ 调度器选中 (时间到达 + 有空闲槽位)
     ↓
┌──────────┐
│ Running  │ ← preExec() 设置 RunDeadline, RunBy
└────┬─────┘
     │
     │ 执行完成
     ↓
   ┌─┴──────────────────────────┐
   │                             │
   ↓                             ↓
┌──────────┐              ┌──────────┐
│ Finished │              │  Failed  │
└──────────┘              └────┬─────┘
                               │
                 ┌─────────────┼─────────────┐
                 │                            │
                 │ ConsecutiveFailures        │ ConsecutiveFailures
                 │ <= MaxRetries              │ > MaxRetries
                 ↓                            ↓
          ┌──────────┐                 ┌──────────┐
          │Scheduled │ ← 15秒后重试     │  Failed  │ ← 最终失败
          │(retry)   │                 │ (final)  │
          └──────────┘                 └──────────┘

特殊情况：
┌──────────┐
│ Running  │
└────┬─────┘
     │ CancelJob() 或 超时
     ↓
┌──────────┐
│ Canceled │
└──────────┘

定时任务：
┌──────────┐
│ Finished │ (IsRecurring=true)
└────┬─────┘
     │ postExec() 计算下次执行时间
     ↓
┌──────────┐
│Scheduled │ ← Scheduled = cronExpr.Next(now)
└──────────┘
```

### 2. 调度循环详解

```go
func (s *Scheduler) Run(ctx context.Context) error {
    // 1. 初始化
    consumer := s.pubsubService.Subscribe(ctx, PubSubTopicCancelJob, s.handleCancelJob)
    s.registerSystemJobs()  // 注册系统任务（purge, overdue）
    s.createSystemJobs(ctx) // 创建定时任务
    s.executor.FinishRegistration()

    timer := pkgJobs.NewSchedulerTimer()

    // 2. 主循环
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()  // 优雅退出

        case newTime := <-s.signal:
            // 外部信号：新任务插入
            timer.RescheduleEarlier(newTime)

        case now := <-timer.Ch():
            // 定时器触发：处理就绪任务
            count, nextExec, gotAllJobs, err := s.processReadyJobs(ctx, now)

            // 重置定时器
            timer.ResetAt(nextExec, !gotAllJobs)
        }
    }
}
```

**关键点**：

1. **全局锁保护**
   - 每次 `processReadyJobs` 获取全局锁
   - 防止多实例同时处理同一任务
   - 锁的粒度：整个任务加载和启动过程

2. **并发控制**
   ```go
   availableSlots := maxRunning - countRunning
   jobs := jobStore.ListReady(ctx, now, availableSlots+1)  // 多取一个检测是否还有更多

   if len(jobs) > availableSlots {
       jobs = jobs[:availableSlots]
       gotAllJobs = false  // 标记有更多任务，设置 edgy 模式
   }
   ```

3. **任务启动**
   ```go
   for _, job := range jobs {
       s.preExec(job)                         // 更新状态为 running
       s.jobStore.UpdateExecution(ctx, job)   // 持久化
       s.runJob(ctx, job, jobLog)             // 异步执行
   }
   ```

### 3. 任务执行流程

```go
func (s *Scheduler) runJob(ctx context.Context, j *jobModel.ScheduleJob, jobLog *logger.SugaredMateLogger) {
    s.wgRunning.Add(1)
    go func(ctx context.Context, jobName, jobType, jobData string, jobRunDeadline *datetime.Local) {
        defer s.wgRunning.Done()

        // 1. 执行任务
        execResult, execFailure := s.doExec(ctx, jobName, jobType, jobData, jobRunDeadline)

        // 2. 获取全局锁（使用独立的 context，不受任务超时影响）
        backgroundCtx := context.Background()
        mx, _ := globalLock(backgroundCtx, s.mxManager)
        defer mx.Unlock(backgroundCtx)

        // 3. 重新加载任务（数据库可能已被其他实例修改）
        job, _ := s.jobStore.GetByName(backgroundCtx, jobName)

        // 4. 更新任务状态
        postExec(job, execResult, execFailure)
        s.jobStore.UpdateExecution(backgroundCtx, job)

        // 5. 发布状态变更事件
        publishStateChange(backgroundCtx, s.pubsubService, job)

        // 6. 触发下一轮调度
        if job.State == StateScheduled {
            s.scheduleProcessing(time.Time(*job.Scheduled))
        } else {
            s.scheduleIfHaveMoreJobs()  // 释放槽位，检查是否有待执行任务
        }
    }(ctx, j.Name, j.Type, j.Data, j.RunDeadline, jobLog)
}
```

**为什么使用 `context.Background()`？**

任务执行超时后，`ctx` 已经被取消，但我们仍需要：
- 更新任务状态到数据库
- 释放全局锁
- 发布状态变更事件

使用独立的 context 确保清理工作不受任务超时影响。

## 调用方式

### 1. 注册任务处理器

**步骤 1：实现 Handler 接口**

```go
// app/service/myservice/job_handler.go
package myservice

import (
    "context"
    "encoding/json"

    pkgJobs "github.com/easysoft/gitfox-community/pkg/jobs"
)

type MyJobHandler struct {
    myService *MyService  // 注入依赖
}

type MyJobInput struct {
    UserID string `json:"user_id"`
    Action string `json:"action"`
}

type MyJobResult struct {
    Status  string `json:"status"`
    Message string `json:"message"`
}

func (h *MyJobHandler) Handle(ctx context.Context, input string,
                               reportFn pkgJobs.ProgressReporter) (string, error) {
    // 解析输入
    var params MyJobInput
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", err
    }

    // 执行业务逻辑
    if err := h.myService.PerformAction(ctx, params.UserID, params.Action); err != nil {
        return "", err
    }

    // 报告进度
    reportFn(100, "Action completed")

    // 返回结果
    result := MyJobResult{Status: "success", Message: "Action performed"}
    data, _ := json.Marshal(result)
    return string(data), nil
}
```

**步骤 2：注册到 Scheduler**

```go
// cmd/cli/operations/server/app.go
func (app *App) Start(ctx context.Context) error {
    // 注册自定义任务处理器
    myHandler := &myservice.MyJobHandler{myService: app.MyService}
    app.JobScheduler.executor.Register("my-job-type", myHandler)

    // 启动调度器
    go func() {
        if err := app.JobScheduler.Run(ctx); err != nil {
            log.Error("Job scheduler failed", "error", err)
        }
    }()

    // ...
}
```

### 2. 创建一次性任务

```go
// 在 Controller 或 Service 中
func (c *MyController) TriggerAsyncTask(ctx *gin.Context) {
    // 准备输入数据
    input := MyJobInput{
        UserID: ctx.GetString("user_id"),
        Action: "send_email",
    }
    inputJSON, _ := json.Marshal(input)

    // 创建任务
    err := c.scheduler.RunJob(
        ctx,
        "send-email-12345",         // 任务名称（唯一）
        "my-job-type",              // 任务类型
        string(inputJSON),          // 输入数据
        300,                        // 最大执行时间（秒）
    )

    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    ctx.JSON(200, gin.H{"message": "Task created"})
}
```

### 3. 创建定时任务

```go
// 系统启动时注册定时任务
func (s *Scheduler) registerMyRecurringJob(ctx context.Context) error {
    input := MyJobInput{Action: "cleanup"}
    inputJSON, _ := json.Marshal(input)

    return s.AddRecurringWithData(
        ctx,
        "cleanup-job",              // 任务名称
        "my-job-type",              // 任务类型
        string(inputJSON),          // 输入数据
        "0 2 * * *",                // Cron 表达式（每天凌晨 2 点）
        10*time.Minute,             // 最大执行时间
    )
}
```

### 4. 取消正在运行的任务

```go
err := scheduler.CancelJob(ctx, "send-email-12345")
if err != nil {
    return err
}
```

**注意事项**：
- 不能取消定时任务（`IsRecurring=true`）
- 只能取消 `scheduled` 或 `running` 状态的任务
- 取消会通过 PubSub 广播到所有实例

### 5. 查询任务进度

```go
// 查询单个任务
progress, err := scheduler.GetJobProgress(ctx, "send-email-12345")
// 返回：{"state": "running", "progress": 50, "result": "...", "failure": ""}

// 查询任务组
progresses, err := scheduler.GetJobProgressForGroup(ctx, "email-batch-001")
// 返回：[{...}, {...}, ...]
```

## 安全边界问题

### 1. 分布式锁的必要性

**问题**：多实例环境下，如何避免同一任务被重复执行？

**解决方案**：
```go
func (s *Scheduler) processReadyJobs(ctx context.Context, now time.Time) (...) {
    mx, err := globalLock(ctx, s.mxManager)  // 获取全局锁 "jobs"
    if err != nil {
        return 0, time.Time{}, false, err
    }
    defer mx.Unlock(ctx)

    // 加锁期间的操作：
    // 1. 查询就绪任务
    // 2. 更新状态为 running
    // 3. 启动 goroutine 执行
}
```

**锁的粒度**：
- ✅ 保护：加载任务 → 更新状态 → 启动执行
- ❌ 不保护：任务具体执行过程（会导致串行化）

**风险点**：
- 如果锁未正确释放，整个集群停止调度
- 如果锁超时时间过短，可能导致双重执行

**推荐配置**：
```go
// pkg/lock/mutex.go
const DefaultTimeout = 30 * time.Second  // 锁超时时间
```

### 2. 任务超时与优雅退出

**问题**：如何防止任务无限期运行？

**解决方案**：

```go
func (s *Scheduler) doExec(...) (execResult, execError string) {
    // 设置超时上下文
    jobCtx, done := context.WithDeadline(ctx, execDeadline)
    defer done()

    // 注册取消函数
    s.cancelJobMap[jobName] = done

    // 执行任务
    execResult, err := s.executor.Exec(jobCtx, ...)

    // 清理取消函数
    delete(s.cancelJobMap, jobName)

    return
}
```

**Handler 的责任**：
```go
func (h *MyHandler) Handle(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
    for _, item := range items {
        // 必须定期检查上下文
        select {
        case <-ctx.Done():
            return "", ctx.Err()  // 优雅退出
        default:
        }

        // 处理 item
    }
}
```

**常见错误**：
```go
// ❌ 错误：忽略上下文，导致无法取消
func (h *BadHandler) Handle(ctx context.Context, ...) (string, error) {
    time.Sleep(10 * time.Hour)  // 无法被取消
    return "done", nil
}

// ✅ 正确：响应上下文
func (h *GoodHandler) Handle(ctx context.Context, ...) (string, error) {
    select {
    case <-time.After(10 * time.Hour):
        return "done", nil
    case <-ctx.Done():
        return "", ctx.Err()
    }
}
```

### 3. 过期任务回收

**问题**：实例异常退出，导致任务永久处于 `running` 状态

**解决方案**：`jobOverdue` 系统任务

```go
// 每 20 分钟运行一次
const JobCronOverdue = "*/20 * * * *"

func (j *jobOverdue) Handle(ctx context.Context, ...) (string, error) {
    // 查询超过 RunDeadline 的任务
    overdueJobs := j.store.ListDeadlineExceeded(ctx, now)

    for _, job := range overdueJobs {
        // 标记为失败
        postExec(job, "", "deadline exceeded")
        j.store.UpdateExecution(ctx, job)
    }

    return fmt.Sprintf("reclaimed %d jobs", len(overdueJobs)), nil
}
```

**为什么需要？**
- 服务器崩溃：任务状态未更新
- 网络分区：实例间状态不一致
- 数据库事务失败：状态更新回滚

### 4. 任务数据隔离

**问题**：如何防止不同业务的任务互相干扰？

**解决方案**：

1. **使用唯一的任务名称**
   ```go
   // ✅ 推荐：包含业务标识
   jobName := fmt.Sprintf("repo-%d:scan-%s", repoID, scanID)

   // ❌ 不推荐：通用名称
   jobName := "scan-job"
   ```

2. **使用 GroupID 分组**
   ```go
   // 批量创建任务
   jobs := []*jobModel.ScheduleJob{
       {Name: "task-1", GroupID: "batch-123", ...},
       {Name: "task-2", GroupID: "batch-123", ...},
   }
   scheduler.RunJobs(ctx, "batch-123", jobs)

   // 批量删除
   scheduler.PurgeJobsByGroupID(ctx, "batch-123")
   ```

3. **任务类型隔离**
   - 不同业务使用不同的 `jobType`
   - 每个类型对应独立的 Handler
   - Handler 内部可以进一步验证权限

### 5. 任务数据安全

**敏感数据存储**：

```go
// ❌ 不安全：密码明文存储在 job.Data
input := map[string]string{
    "user": "admin",
    "password": "secret123",  // 明文存储在数据库
}

// ✅ 推荐：使用 token 或引用
input := map[string]string{
    "user": "admin",
    "credential_id": "cred-abc123",  // 从密钥管理系统获取
}
```

**输入验证**：

```go
func (h *MyHandler) Handle(ctx context.Context, input string, ...) (string, error) {
    var params MyJobInput
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        return "", errors.New("invalid input format")
    }

    // 验证业务逻辑
    if params.UserID == "" {
        return "", errors.New("user_id is required")
    }

    // 验证权限（如果需要）
    if !h.authz.CanPerformAction(ctx, params.UserID, params.Action) {
        return "", errors.New("permission denied")
    }

    // ...
}
```

### 6. 并发限制

**问题**：如何避免任务过载？

**解决方案**：

1. **全局并发限制**
   ```go
   scheduler, _ := NewScheduler(
       jobStore,
       mxManager,
       pubsubService,
       log,
       "instance-1",
       10,  // maxRunning：最多同时运行 10 个任务
       24*time.Hour,
   )
   ```

2. **业务级限流**（在 Handler 内部）
   ```go
   type RateLimitedHandler struct {
       limiter *rate.Limiter
   }

   func (h *RateLimitedHandler) Handle(ctx context.Context, ...) (string, error) {
       if err := h.limiter.Wait(ctx); err != nil {
           return "", err
       }
       // 执行任务
   }
   ```

## 常见问题 Q&A

### Q1: 如何选择 MaxDurationSeconds？

**A**: 根据任务类型设置：

| 任务类型 | 推荐时长 | 理由 |
|---------|---------|-----|
| 快速 API 调用 | 30-60 秒 | 网络请求通常很快 |
| 邮件发送 | 120-300 秒 | SMTP 可能较慢 |
| 代码扫描 | 600-1800 秒 | 取决于代码量 |
| 大文件处理 | 1800-3600 秒 | I/O 密集型 |
| 数据清理 | 300-600 秒 | 批量操作 |

**原则**：
- 宁可设置长一点，避免误杀正常任务
- 配合 `reportFn()` 实时反馈进度
- 超时任务会被 `jobOverdue` 回收

### Q2: 定时任务执行失败会怎样？

**A**: 定时任务失败后仍会重新调度

```go
func postExec(job *jobModel.ScheduleJob, resultData, resultErr string) {
    if job.IsRecurring {
        // 无论成功或失败，都计算下次执行时间
        exp, _ := cronexpr.Parse(job.RecurringCron)
        nextTime := exp.Next(now)
        job.Scheduled = &nextTime
        job.State = StateScheduled

        if resultErr != "" {
            job.ConsecutiveFailures++  // 记录连续失败次数
        } else {
            job.ConsecutiveFailures = 0  // 成功后重置
        }
    }
}
```

**监控建议**：
- 定期检查 `ConsecutiveFailures` 字段
- 设置告警阈值（如：连续失败 > 3 次）

### Q3: MaxRetries 如何设置？

**A**: 取决于任务特性

```go
// 幂等任务（可以安全重试）
scheduler.RunJob(ctx, jobName, jobType, data, 300)  // MaxRetries 默认 0

// 临时性失败可能性高的任务
job := &jobModel.ScheduleJob{
    MaxRetries: 3,  // 最多重试 3 次
    ...
}

// 关键任务
job := &jobModel.ScheduleJob{
    MaxRetries: 5,  // 多次重试
    ...
}
```

**重试延迟**：
```go
const retryDelay = 15 * time.Second  // 固定延迟
```

**注意**：
- 定时任务（`IsRecurring=true`）忽略 `MaxRetries`
- 重试不会改变 `jobName`，注意幂等性

### Q4: 如何实现任务依赖？

**A**: 当前版本不直接支持，需要手动实现

**方案 1：在 Handler 内部检查**
```go
func (h *PipelineHandler) Handle(ctx context.Context, input string, ...) (string, error) {
    var params PipelineInput
    json.Unmarshal([]byte(input), &params)

    // 检查前置任务
    for _, depJobName := range params.Dependencies {
        depJob, _ := h.jobStore.GetByName(ctx, depJobName)
        if depJob.State != StateFinished {
            return "", errors.New("dependency not finished: " + depJobName)
        }
    }

    // 执行主逻辑
}
```

**方案 2：使用回调链**
```go
func (h *StepAHandler) Handle(ctx context.Context, ...) (string, error) {
    // 执行 Step A
    doStepA()

    // 触发 Step B
    h.scheduler.RunJob(ctx, "step-b-job", "step-b-type", data, 300)

    return "step A done", nil
}
```

### Q5: 多实例部署时如何确保任务不重复执行？

**A**: 依赖分布式锁 + 数据库行级锁

**保障机制**：
1. **调度阶段**：`globalLock("jobs")` 保护
2. **状态更新**：数据库事务原子性
3. **实例 ID**：`RunBy` 字段记录执行实例

**验证方法**：
```sql
-- 检查是否有重复执行
SELECT name, COUNT(*) FROM schedule_jobs
WHERE state = 'running' AND is_recurring = 0
GROUP BY name
HAVING COUNT(*) > 1;
```

### Q6: 如何处理长时间运行的任务？

**A**: 使用进度报告 + 合理的超时时间

```go
func (h *LongRunningHandler) Handle(ctx context.Context, input string,
                                     reportFn pkgJobs.ProgressReporter) (string, error) {
    items := loadItems()
    total := len(items)

    for i, item := range items {
        // 检查上下文
        select {
        case <-ctx.Done():
            return "", ctx.Err()
        default:
        }

        // 处理单个 item
        processItem(item)

        // 每 10 个 item 报告一次进度
        if i%10 == 0 {
            progress := (i + 1) * 100 / total
            reportFn(progress, fmt.Sprintf("Processed %d/%d", i+1, total))
        }
    }

    reportFn(100, "All items processed")
    return "success", nil
}
```

**客户端轮询进度**：
```go
ticker := time.NewTicker(5 * time.Second)
defer ticker.Stop()

for {
    select {
    case <-ticker.C:
        progress, _ := scheduler.GetJobProgress(ctx, jobName)
        fmt.Printf("Progress: %d%%, State: %s\n", progress["progress"], progress["state"])

        if progress["state"] == "finished" || progress["state"] == "failed" {
            return
        }
    }
}
```

### Q7: 任务日志在哪里查看？

**A**: 独立的 Job 日志文件

**配置**：
```toml
# config.toml
[logging]
log-dir = "/var/log/gitfox"

[logging.job]
filename = "job.log"
```

**日志输出**：
- 所有任务执行日志写入 `job.log`
- 包含结构化字段：`job.Name`, `job.Type`, `duration`, `job.Result`, `job.Failure`
- 独立于 API 日志和系统日志

**查询示例**：
```bash
# 查看特定任务的日志
grep "job.Name=send-email-12345" /var/log/gitfox/job.log

# 查看所有失败任务
grep "job.Failure=" /var/log/gitfox/job.log | grep -v 'job.Failure=""'
```

### Q8: 如何清理历史任务？

**A**: 自动清理 + 手动清理

**自动清理**（系统任务）：
```go
const JobCronPurge = "15 */4 * * *"  // 每 4 小时清理一次

// 删除超过 retentionTime 的已完成/失败任务
```

**手动清理**：
```go
// 删除特定任务组
scheduler.PurgeJobsByGroupID(ctx, "batch-123")

// 直接操作数据库（谨慎使用）
DELETE FROM schedule_jobs WHERE state IN ('finished', 'failed') AND edited_date < NOW() - INTERVAL 30 DAY;
```

### Q9: 如何监控任务系统的健康状态？

**A**: 监控以下指标

| 指标 | SQL 查询 | 告警阈值 |
|-----|---------|---------|
| 运行中任务数 | `SELECT COUNT(*) FROM schedule_jobs WHERE state='running'` | > maxRunning |
| 调度延迟 | `SELECT COUNT(*) FROM schedule_jobs WHERE state='scheduled' AND scheduled < NOW()` | > 10 |
| 连续失败任务 | `SELECT COUNT(*) FROM schedule_jobs WHERE consecutive_failures > 3` | > 0 |
| 过期未完成 | `SELECT COUNT(*) FROM schedule_jobs WHERE state='running' AND run_deadline < NOW()` | > 0 |

**Prometheus 指标示例**：
```go
var (
    jobsRunning = prometheus.NewGauge(...)
    jobsScheduled = prometheus.NewGauge(...)
    jobsFailed = prometheus.NewCounter(...)
)

// 定期更新
func updateMetrics(ctx context.Context) {
    running, _ := jobStore.CountRunning(ctx)
    jobsRunning.Set(float64(running))
}
```

## 最佳实践

### 1. Handler 实现规范

```go
type BestPracticeHandler struct {
    logger *logger.SugaredMateLogger
    // ...
}

func (h *BestPracticeHandler) Handle(ctx context.Context, input string,
                                      reportFn pkgJobs.ProgressReporter) (string, error) {
    // 1. 日志记录（开始）
    h.logger.Infow("Job started", "input", input)

    // 2. 输入验证
    var params MyParams
    if err := json.Unmarshal([]byte(input), &params); err != nil {
        h.logger.Errorw("Invalid input", "error", err)
        return "", errors.New("invalid input: " + err.Error())
    }

    // 3. 上下文检查（长循环中）
    for i, item := range params.Items {
        select {
        case <-ctx.Done():
            h.logger.Warnw("Job canceled", "processed", i, "total", len(params.Items))
            return "", ctx.Err()
        default:
        }

        // 4. 业务逻辑
        if err := h.processItem(ctx, item); err != nil {
            // 记录但继续（部分失败）
            h.logger.Warnw("Item failed", "item", item, "error", err)
            continue
        }

        // 5. 进度报告
        if i%10 == 0 {
            progress := (i + 1) * 100 / len(params.Items)
            reportFn(progress, fmt.Sprintf("%d/%d", i+1, len(params.Items)))
        }
    }

    // 6. 构造结果
    result := MyResult{
        Processed: len(params.Items),
        Timestamp: time.Now(),
    }
    resultJSON, _ := json.Marshal(result)

    // 7. 日志记录（完成）
    h.logger.Infow("Job completed", "result", string(resultJSON))

    return string(resultJSON), nil
}
```

### 2. 任务命名规范

```go
// 推荐格式：<业务>-<资源ID>-<操作>-<唯一标识>
func generateJobName(business, resourceID, operation string) string {
    uuid := generateUUID()
    return fmt.Sprintf("%s-%s-%s-%s", business, resourceID, operation, uuid)
}

// 示例
"repo-123-scan-a1b2c3d4"
"pipeline-456-build-e5f6g7h8"
"space-789-cleanup-i9j0k1l2"
```

### 3. 错误处理策略

```go
func (h *MyHandler) Handle(ctx context.Context, input string, ...) (string, error) {
    // 区分临时性错误和永久性错误

    if err := h.callExternalAPI(); err != nil {
        // 临时性错误：网络超时、限流
        if isTemporaryError(err) {
            return "", err  // 允许重试
        }

        // 永久性错误：认证失败、资源不存在
        if isPermanentError(err) {
            // 记录详细日志
            h.logger.Errorw("Permanent error", "error", err)
            // 返回明确的错误信息，不触发重试
            return "", errors.New("permanent failure: " + err.Error())
        }
    }

    return "success", nil
}

func isTemporaryError(err error) bool {
    // 检查错误类型
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true
    }
    // 检查 HTTP 状态码
    if httpErr, ok := err.(*HTTPError); ok {
        return httpErr.StatusCode == 429 || httpErr.StatusCode >= 500
    }
    return false
}
```

## 总结

| 特性 | 实现方式 | 优势 |
|-----|---------|-----|
| 分布式协调 | Redis 分布式锁 | 多实例安全调度 |
| 任务持久化 | MySQL/PostgreSQL | 数据不丢失，可审计 |
| 定时调度 | Cron 表达式 | 灵活的时间规则 |
| 并发控制 | `maxRunning` 限制 | 防止过载 |
| 故障恢复 | `jobOverdue` 回收 | 自动清理僵尸任务 |
| 实时同步 | PubSub 事件 | 跨实例取消任务 |
| 独立日志 | Job 专用 logger | 便于问题排查 |

**架构优势**：
- 高可用：多实例无单点故障
- 可扩展：水平扩展，增加实例即可
- 可观测：独立日志 + 数据库记录
- 安全性：分布式锁 + 超时控制 + panic 恢复
