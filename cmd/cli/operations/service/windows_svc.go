//go:build windows

package service

// windows_svc.go 使用 Go 原生的 Windows Service API 管理 GitFox 系统服务。
//
// 核心依赖：
//   - golang.org/x/sys/windows/svc     — 服务运行时事件循环（Handler 接口）
//   - golang.org/x/sys/windows/svc/mgr — 服务控制管理器（SCM）的 CRUD 操作
//
// 本文件中所有函数仅在 GOOS=windows 下编译，由 windows.go / register.go 等文件路由调用。

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/server"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/Microsoft/go-winio"
	"github.com/quicklyon/kingpin/v2"
	"golang.org/x/sys/windows"
)

// Windows 服务名称常量
const windowsServiceName = "gitfox"

// checkSudoWindows 通过 Windows Token API 检测当前进程是否以管理员权限运行。
// 如果不是管理员，返回国际化的友好错误提示，避免用户在后续 SCM 操作时
// 收到晦涩的 "Access is denied" 英文系统错误。
func checkSudoWindows() error {
	token := windows.GetCurrentProcessToken()
	if !token.IsElevated() {
		return errors.New(clii18n.T(i18n.CliStartErrRequiresAdmin))
	}
	return nil
}

// ─────────────────────────────────────────────
// 服务注册 / 启动
// ─────────────────────────────────────────────

// registerWindows 通过 Windows SCM 原生 API 将 GitFox 注册为系统服务并立即启动。
//
// 流程：
//  1. 连接到 Windows 服务控制管理器（SCM）
//  2. 防御性清理已有的同名服务（允许失败）
//  3. 使用 mgr.CreateService 注册服务，传入可执行文件路径与启动参数
//  4. 通过 s.Start() 触发服务启动
//
// Windows 常见问题排查：
//   - Access is denied → 必须以"管理员身份运行"终端执行此命令
//   - 服务启动后立即崩溃 → 检查 prod.env 中的 GITFOX_SERVER_HOST 是否为本机实际可绑定 IP
func registerWindows(exePath, workDir string) error {
	t := clii18n.T
	fmt.Println(t(i18n.CliStartStepWindowsService))

	// 1. 连接 SCM
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsConnectScm), err)
	}
	defer m.Disconnect()

	// 2. 防御性清理旧实例（忽略错误，服务可能本就不存在）
	if oldSvc, err := m.OpenService(windowsServiceName); err == nil {
		_ = controlAndWaitForStop(oldSvc)
		_ = oldSvc.Delete()
		oldSvc.Close()

		// 等待 SCM 释放服务注册（避免 "service marked for deletion" 错误）
		time.Sleep(1 * time.Second)
	}

	// 3. 注册新服务
	// 启动参数为 "service run"，指示 GitFox 进入 Windows 服务模式
	svcConfig := mgr.Config{
		DisplayName:      "GitFox Service",
		Description:      "GitFox 代码托管平台后台服务",
		StartType:        mgr.StartAutomatic,
		ServiceStartName: "", // 使用 LocalSystem 帐户
	}

	s, err := m.CreateService(
		windowsServiceName,
		exePath,
		svcConfig,
		"service", "run", // 启动参数
		"--workdir", workDir,
	)
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsInstallService), err)
	}
	defer s.Close()

	// 4. 启动服务
	if err := s.Start(); err != nil {
		return fmt.Errorf(t(i18n.CliStartFailed), err)
	}

	return nil
}

// restartWindows 通过 SCM 重启已注册的 GitFox Windows 服务。
func restartWindows() error {
	t := clii18n.T
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsConnectScm), err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsOpenService), err)
	}
	defer s.Close()

	// 先停止（忽略错误，可能本就未运行）
	_ = controlAndWaitForStop(s)

	// 再启动
	return s.Start()
}

// ─────────────────────────────────────────────
// 服务停止
// ─────────────────────────────────────────────

// stopWindows 通过 SCM 停止正在运行的 GitFox Windows 服务。
func stopWindows() error {
	t := clii18n.T
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsConnectScm), err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsOpenService), err)
	}
	defer s.Close()

	return controlAndWaitForStop(s)
}

// controlAndWaitForStop 发送停止信号并等待服务实际停止，超时 30 秒。
func controlAndWaitForStop(s *mgr.Service) error {
	t := clii18n.T
	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf(t(i18n.CliStopErrWindowsStopCmd), err)
	}

	timeout := time.Now().Add(30 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return errors.New(t(i18n.CliStopErrWindowsTimeout))
		}
		time.Sleep(500 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf(t(i18n.CliStopErrWindowsQueryStatus), err)
		}
	}
	return nil
}

// ─────────────────────────────────────────────
// 服务卸载
// ─────────────────────────────────────────────

// uninstallWindows 通过 SCM 彻底移除 GitFox Windows 服务注册。
func uninstallWindows() error {
	t := clii18n.T

	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf(t(i18n.CliStartErrWindowsConnectScm), err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return fmt.Errorf(t(i18n.CliUninstallFailed), err)
	}
	defer s.Close()

	// 先停止（允许失败，服务可能本来就没在运行）
	_ = controlAndWaitForStop(s)

	if err := s.Delete(); err != nil {
		return fmt.Errorf(t(i18n.CliUninstallFailed), err)
	}
	return nil
}

// ─────────────────────────────────────────────
// 服务查询 / 状态检测
// ─────────────────────────────────────────────

// inspectWindows 通过 SCM 查询已注册的 GitFox Windows 服务元数据。
func inspectWindows() ServiceInfo {
	m, err := mgr.Connect()
	if err != nil {
		return ServiceInfo{Registered: false}
	}
	defer m.Disconnect()

	s, err := m.OpenService(windowsServiceName)
	if err != nil {
		return ServiceInfo{Registered: false}
	}
	defer s.Close()

	info := ServiceInfo{Registered: true}

	cfg, err := s.Config()
	if err != nil {
		info.Corrupted = true
		return info
	}

	// BinaryPathName 格式: "C:\path\gitfox.exe service run --workdir C:\path\workdir"
	// 提取可执行文件路径（第一个字段）
	info.ExePath = extractExePathFromBinaryPath(cfg.BinaryPathName)

	// 从启动参数中提取 --workdir
	info.WorkDir = extractWorkDirFromBinaryPath(cfg.BinaryPathName)

	if info.ExePath == "" {
		info.Corrupted = true
	}
	return info
}

// ─────────────────────────────────────────────
// 服务运行入口（由 SCM 调用）
// ─────────────────────────────────────────────

// RegisterServiceRun 注册 `service run` 内部子命令。
// 该命令由 Windows SCM 启动进程时调用，不面向用户。
func RegisterServiceRun(app *kingpin.Application) {
	var workDir string
	cmd := app.Command("service", "Windows 服务内部管理命令").Hidden()
	runCmd := cmd.Command("run", "以 Windows 服务模式运行 (由 SCM 调用)").Action(func(_ *kingpin.ParseContext) error {
		return runWindowsService(workDir)
	})
	runCmd.Flag("workdir", "工作目录").StringVar(&workDir)
}

// runWindowsService 是 Windows 服务模式的实际入口。
// 它切换工作目录后调用 svc.Run() 进入 SCM 事件驱动循环。
func runWindowsService(workDir string) error {
	t := clii18n.T
	// 切换工作目录（确保 prod.env 等相对路径正确解析）
	if workDir != "" {
		if err := os.Chdir(workDir); err != nil {
			return fmt.Errorf(t(i18n.CliStartErrWindowsChdir), err)
		}
	}

	return svc.Run(windowsServiceName, &gitfoxService{})
}

// gitfoxService 实现 svc.Handler 接口，是 Windows 服务的核心执行体。
type gitfoxService struct{}

// Execute 是 Windows SCM 调用的服务主循环。
//
// 生命周期：
//  1. 报告 StartPending 状态
//  2. 初始化 App（数据库、路由、SSH 等）
//  3. 启动 HTTP / SSH / CI Poller / Job Scheduler
//  4. 报告 Running 状态
//  5. 监听 SCM 的 Stop / Shutdown 命令
//  6. 执行优雅关闭
func (gs *gitfoxService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	// 接受 Stop 和 Shutdown 命令
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听 SCM 通道：当接收到 Stop 或 Shutdown 指令时，立刻触发 Context.Cancel() 平滑切断后路
	go func() {
	loop:
		for {
			select {
			case cr := <-r:
				switch cr.Cmd {
				case svc.Stop, svc.Shutdown:
					cancel()
					break loop
				case svc.Interrogate:
					s <- cr.CurrentStatus
				}
			case <-ctx.Done():
				break loop
			}
		}
	}()

	// 核心组件：调用与控制台相同的启动引擎
	// 使用 onReady 钩子在真正启动完成后再向 SCM 报告 svc.Running
	err := server.RunServerWithContext(ctx, "prod", true, func() {
		s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	})

	// 如果服务由于致命错误退出
	if err != nil {
		s <- svc.Status{State: svc.Stopped}
		return true, 1
	}

	// 正常平滑退出
	s <- svc.Status{State: svc.StopPending}
	return false, 0
}

// ─────────────────────────────────────────────
// 辅助函数
// ─────────────────────────────────────────────

// extractExePathFromBinaryPath 从 SCM BinaryPathName 中提取可执行文件路径。
// BinaryPathName 格式: "C:\path\gitfox.exe service run --workdir C:\workdir"
// 或带引号: "\"C:\path\gitfox.exe\" service run --workdir C:\workdir"
func extractExePathFromBinaryPath(binPath string) string {
	if len(binPath) == 0 {
		return ""
	}
	// 处理带引号的路径
	if binPath[0] == '"' {
		end := 1
		for end < len(binPath) && binPath[end] != '"' {
			end++
		}
		if end < len(binPath) {
			return binPath[1:end]
		}
		return binPath[1:]
	}
	// 不带引号：取第一个空格前的部分
	for i, c := range binPath {
		if c == ' ' {
			return binPath[:i]
		}
	}
	return binPath
}

// extractWorkDirFromBinaryPath 从 SCM BinaryPathName 中提取 --workdir 参数值。
func extractWorkDirFromBinaryPath(binPath string) string {
	const flag = "--workdir"
	pos := strings.Index(binPath, flag)
	if pos < 0 {
		return ""
	}
	rest := binPath[pos+len(flag):]
	// 跳过前导空格
	rest = strings.TrimLeft(rest, " ")
	if len(rest) == 0 {
		return ""
	}
	// 读取值（到下一个空格或末尾）
	if spaceIdx := strings.Index(rest, " "); spaceIdx >= 0 {
		return rest[:spaceIdx]
	}
	return rest
}

// newDockerHTTPClientNpipe 基于 Windows named pipe 路径构造 docker daemon 通信的 http.Client。
// 仅在 GOOS=windows 下编译，由 check.go 的 newDockerHTTPClient 在 scheme=="npipe" 时分发至此。
func newDockerHTTPClientNpipe(u *url.URL) (*http.Client, string, error) {
	path := u.Path
	if path == "" {
		path = `\\.\pipe\docker_engine`
	}
	if _, err := os.Stat(path); err != nil {
		return nil, "", err
	}

	client := &http.Client{
		Timeout: dockerProbeTimeout,
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return winio.DialPipeContext(ctx, path)
			},
		},
	}
	return client, "http://docker", nil
}
