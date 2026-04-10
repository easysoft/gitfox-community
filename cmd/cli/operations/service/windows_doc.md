# GitFox Windows 服务架构演进文档

本文档记录了 GitFox CLI 在 Windows 平台上将应用注册为系统后台服务的架构演进过程。
经历了三个主要阶段，每个阶段解决了前一阶段遗留的核心问题。

---

## 阶段一：NSSM 外部工具方案（已废弃）

### 方案描述

最初的 Windows 服务方案依赖第三方工具 [NSSM (Non-Sucking Service Manager)](https://nssm.cc/)。
NSSM 的工作原理是将任意可执行文件包装为 Windows 系统服务：

```
用户执行 gitfox start
       ↓
CLI 调用 exec.Command("nssm", "install", "gitfox", "gitfox.exe", "server", "prod")
       ↓
NSSM 在 Windows SCM 中注册一个服务条目
       ↓
SCM 启动 nssm.exe（宿主进程），nssm.exe 再拉起 gitfox.exe（子进程）
```

### 存在的问题

| 问题 | 说明 |
|---|---|
| **三方依赖** | 要求用户额外下载 `nssm.exe` 并加入 PATH，增大了部署复杂度和沟通成本 |
| **进程嵌套** | `SCM → nssm.exe → gitfox.exe` 三层进程链，停止信号传递不可靠 |
| **优雅关机失败** | NSSM 在超时后直接 `TerminateProcess` 杀子进程，数据库连接/文件锁无法安全释放 |
| **安全审计风险** | 引入未签名的第三方 exe 到生产环境 |

### 废弃原因

> 不让用户安装第三方软件，降低部署门槛。

---

## 阶段二：Go 原生 SCM API + 重复启动代码（已废弃）

### 方案描述

为消除 NSSM 依赖，改用 Go 的 `golang.org/x/sys/windows/svc` 包直接与 Windows 服务控制管理器 (SCM) 交互。
GitFox 自身同时扮演"服务管理器"和"服务执行体"两个角色：

```
用户执行 gitfox start
       ↓
CLI 通过 mgr.CreateService() 在 SCM 注册 gitfox.exe 本身
       ↓
SCM 直接启动 gitfox.exe service run（无中间进程）
       ↓
gitfox.exe 内部实现 svc.Handler 接口，响应 SCM 的 Stop/Shutdown 命令
```

这一阶段解决了三方依赖和进程嵌套问题，但引入了新的**代码质量问题**：

### 存在的问题

`windows_svc.go` 中的 `Execute()` 函数完整复制了 `server.go` 中 `run()` 的全部业务启动逻辑
（约 130+ 行），包括：

- 数据库 Migrate
- HTTP 路由初始化
- SSH 服务器启动
- CI Poller 拉起
- Job Scheduler 启动
- 优雅关机流程

这意味着每新增一个核心组件，都需要同步修改**两处**代码（Linux 前台 + Windows 服务），
极易遗漏导致平台行为不一致。

### 废弃原因

> 不允许启动代码重复，也不允许重写启动逻辑。核心业务启动代码必须只有一份。

---

## 阶段三：Context 托管模型（当前方案）

### 设计目标

通过引入**上下文托管模型 (Context-based Hosting Model)**，实现：

1. **核心启动逻辑只有一份** → `server.RunServerWithContext()`
2. **不依赖任何第三方软件** → 使用 Go 原生 Windows SCM API
3. **完美优雅关机** → 所有组件通过同一个 `context.Context` 协调生命周期

### 架构设计

```
                    ┌─────────────────────────┐
                    │   server.go             │
                    │   RunServerWithContext() │
                    │                         │
                    │  • DB Migrate           │
                    │  • HTTP Server          │
                    │  • SSH Server           │
                    │  • CI Poller            │
                    │  • Job Scheduler        │
                    │  • Graceful Shutdown    │
                    └───────────┬─────────────┘
                                │
                    ctx.Done() = 优雅退出
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                   │
     ┌────────▼────────┐ ┌─────▼──────┐  ┌────────▼────────┐
     │  Linux 前台      │ │ Linux      │  │ Windows SCM     │
     │  (command.run)   │ │ Systemd    │  │ (Execute)       │
     │                  │ │ (systemctl)│  │                  │
     │ signal.Notify    │ │ → 调用     │  │ svc.ChangeReq   │
     │ SIGINT/SIGTERM   │ │   command  │  │ Stop/Shutdown   │
     │      ↓           │ │   .run()   │  │      ↓           │
     │ cancel()         │ │            │  │ cancel()         │
     └─────────────────┘ └────────────┘  └─────────────────┘
```

### 核心代码解析

#### 1. 通用启动引擎 (`server.go`)

```go
// RunServerWithContext 是唯一的业务启动入口。
// 所有平台（Linux 前台、Linux Systemd、Windows SCM）都调用同一个函数。
func RunServerWithContext(ctx context.Context, mode string, enableCI bool, onReady func()) error {
    app, cleanup, err := NewApp(mode)
    // ... 初始化 DB Migrate、HTTP、SSH、Poller、Scheduler ...

    // 当网络端口就绪时，通知外部宿主
    if onReady != nil {
        onReady()
    }

    // 阻塞等待：外部 cancel 或内部致命错误
    select {
    case <-ctx.Done():
        // 优雅关机
    case err := <-serverErrors:
        return err
    }

    // 按序释放资源：HTTP → SSH → Poller → EventSystem → Scheduler
    srv.Shutdown(shutdownCtx)
    // ...
    return nil
}
```

**核心设计点**：

- `ctx` 是生命周期总闸。无论是 `Ctrl+C`、`systemctl stop` 还是 Windows SCM 的 Stop 命令，
  最终都转化为 `cancel()` 调用，统一触发 `ctx.Done()`。
- `onReady` 回调在 HTTP/SSH 端口挂载成功后触发，
  供 Windows SCM 在此时向系统报告 `svc.Running` 状态。

#### 2. Linux 前台入口 (`server.go` → `command.run`)

```go
func (c *command) run(*kingpin.ParseContext) error {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 拦截 Ctrl+C / Systemd 停止信号
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-quit
        cancel() // 信号 → context 取消 → RunServerWithContext 开始优雅关机
    }()

    return RunServerWithContext(ctx, c.mode, c.enableCI, nil)
}
```

#### 3. Windows 服务执行体 (`windows_svc.go` → `Execute`)

```go
func (gs *gitfoxService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
    const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
    s <- svc.Status{State: svc.StartPending}

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 监听 SCM 通道：收到 Stop/Shutdown 时触发 cancel()
    go func() {
        for cr := range r {
            switch cr.Cmd {
            case svc.Stop, svc.Shutdown:
                cancel()  // SCM 停止命令 → context 取消 → 优雅关机
                return
            case svc.Interrogate:
                s <- cr.CurrentStatus
            }
        }
    }()

    // 调用与 Linux 完全相同的启动引擎
    // onReady 回调：业务就绪后才向 SCM 报告 Running
    err := server.RunServerWithContext(ctx, "prod", true, func() {
        s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
    })

    if err != nil {
        s <- svc.Status{State: svc.Stopped}
        return true, 1
    }

    s <- svc.Status{State: svc.StopPending}
    return false, 0
}
```

**关键点**：
- `Execute` 内部**没有任何业务启动代码**（没有 Migrate、没有 HTTP 初始化）
- 它仅做两件事：① 监听 SCM 命令 → `cancel()`；② 调用通用引擎
- `onReady` 回调确保只有在 HTTP/SSH 真正就绪后才向 SCM 报告 `svc.Running`，
  避免 SCM 误判超时

### 服务注册流程 (`registerWindows`)

当用户执行 `gitfox start` 或 `gitfox install` 时：

```
1. mgr.Connect()                          → 连接 Windows SCM
2. 防御性清理旧服务（如果存在）                → 避免 "service already exists" 冲突
3. mgr.CreateService("gitfox",             → 注册新服务
     exePath,                                 启动命令：gitfox.exe service run --workdir C:\path
     svcConfig)
4. s.Start()                               → 触发 SCM 启动服务
```

SCM 启动服务时的调用链：

```
SCM 启动进程 → gitfox.exe service run --workdir C:\gitfox
                  ↓
          RegisterServiceRun 注册的隐藏子命令
                  ↓
          runWindowsService(workDir)
                  ↓
          os.Chdir(workDir)  // 切换工作目录
                  ↓
          svc.Run("gitfox", &gitfoxService{})  // 进入 SCM 事件循环
                  ↓
          gitfoxService.Execute()  // 调用 RunServerWithContext
```

---

## 文件组织

| 文件 | 构建标签 | 职责 |
|---|---|---|
| `server/server.go` | 无 | 通用启动引擎 `RunServerWithContext`，所有平台共享 |
| `service/windows.go` | 无 | Windows 通用辅助（端口检测、日志命令提示、常量定义） |
| `service/windows_svc.go` | `//go:build windows` | Windows SCM API 实现（注册/停止/卸载/执行/状态查询/管理员检测） |
| `service/windows_svc_stub.go` | `//go:build !windows` | 非 Windows 平台的桩函数（确保跨平台编译） |
| `service/register.go` | 无 | 服务注册/重启的跨平台调度层 |
| `service/start.go` | 无 | `gitfox start` 命令的核心决策引擎 |
| `service/stop.go` | 无 | `gitfox stop` 命令 + `readHealthURL` |
| `service/uninstall.go` | 无 | `gitfox uninstall` 命令 |
| `service/check.go` | 无 | 权限/依赖/端口/数据库/禅道检测 |

---

## 优雅关机的信号传递路径对比

### Linux (Systemd)

```
systemctl stop gitfox
    → SIGTERM 信号
        → signal.Notify 捕获
            → cancel()
                → ctx.Done()
                    → HTTP Shutdown → SSH Shutdown → Poller Stop → Scheduler Stop
```

### Windows (SCM)

```
sc stop gitfox / 服务管理器点击停止
    → SCM 发送 svc.Stop 到 ChangeRequest 通道
        → Execute 的监听协程收到
            → cancel()
                → ctx.Done()
                    → HTTP Shutdown → SSH Shutdown → Poller Stop → Scheduler Stop
                        → Execute 返回 svc.StopPending
```

两条路径最终汇聚到同一个 `RunServerWithContext` 的 `select { case <-ctx.Done() }` 分支，
执行完全相同的优雅关机流程。**这就是"只改一处，全平台生效"的核心保证。**

---

## 管理员权限检测

Windows 服务操作（注册/停止/卸载）需要管理员权限。检测逻辑：

```go
// windows_svc.go (仅 Windows 编译)
func checkSudoWindows() error {
    token := windows.GetCurrentProcessToken()
    if !token.IsElevated() {
        return errors.New(clii18n.T(i18n.CliStartErrRequiresAdmin))
    }
    return nil
}
```

在 `install`、`start`、`stop`、`uninstall` 命令入口处统一调用 `CheckSudo()`，
非管理员用户会立即收到友好的国际化提示，而不是等到 SCM 操作失败后才收到英文系统原始错误。
