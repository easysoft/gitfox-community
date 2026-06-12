package service

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamStart = "start"
)

// RegisterStart 注册 `gitfox start` 子命令到父应用中
func RegisterStart(app *kingpin.Application) {
	app.Command(ParamStart, clii18n.T(i18n.CliInstallStartDescription)).Action(runStart)
}

// runStart 是 `gitfox start` 的实际执行入口处理函数。
// 职责流：环境校验 -> 读取配置提取端口 -> 执行系统后台服务注入 -> HTTP 轮询进程存活探测。
func runStart(_ *kingpin.ParseContext) error {
	t := clii18n.T
	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00"))
	successStyle := lipgloss.NewStyle().Bold(true)

	// printBanner(t(i18n.CliStartWelcome))

	// 获取当前执行目录作为 GitFox 的核心工作路径
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
		return nil
	}

	// =========================================================
	// 阶段 0. 启动前置安全拦截 (必须具备 Root 提权环境)
	// =========================================================
	if err := CheckSudo(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	// =========================================================
	// 阶段 1. 启动前置依赖拦截 (依赖是否齐全)
	// =========================================================
	// 若 install 链路已在父进程做过依赖检查，则跳过本阶段，避免 git/docker/svn 警告重复打印。
	// flag 一次性消费：读完立刻 Unsetenv，避免泄漏给本进程后续 spawn 的子进程。
	depsPrechecked := os.Getenv(envDepsPrechecked) == "1"
	_ = os.Unsetenv(envDepsPrechecked)
	if !depsPrechecked {
		// fmt.Println(t(i18n.CliStartStepCheckDeps))
		if ok, _, err := CheckGit(); !ok || err != nil {
			fmt.Println(errStyle.Render(err.Error()))
			return nil
		}

		if ok, _, err := CheckDocker(); !ok || err != nil {
			fmt.Println(errStyle.Render(err.Error()))
			return nil
		}

		// Subversion 仅警告，不阻塞 start：用于支持关联 SVN 代码库功能。
		if ok, _, err := CheckSubversion(); !ok {
			fmt.Println(warnStyle.Render(t(i18n.CliStatusSvnMissing)))
		} else if err != nil {
			fmt.Println(warnStyle.Render(err.Error()))
		}
	}

	// =========================================================
	// 阶段 2. 读取环境配制检查 (识别是否允许提供服务)
	// =========================================================
	// 查找并校验工作目录下是否存在必要的 prod.env
	envPath := filepath.Join(workDir, ProdEnvFileName)
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println(errStyle.Render(t(i18n.CliStartErrEnvNotFound)))
		return nil
	}

	// 解析现有的 env 配置以抓取各项属性
	port := ""
	host := ""
	dsn := ""
	devopsURL := ""
	var sshPort string

	envMap, _ := godotenv.Read(envPath)
	if p, ok := envMap["GITFOX_HTTP_PORT"]; ok && p != "" {
		port = p
	}
	if h, ok := envMap["GITFOX_SERVER_HOST"]; ok && h != "" {
		host = h
	}
	if d, ok := envMap["GITFOX_DB_DATASOURCE"]; ok {
		dsn = d
	}
	if v, ok := envMap["GITFOX_DEVOPS_URL"]; ok {
		devopsURL = v
	}

	// 强制校验：如果这四个核心要素只要有一个为空，说明配置文件已被破坏，强制重新 init
	if port == "" || host == "" || dsn == "" || devopsURL == "" {
		fmt.Println(errStyle.Render(t(i18n.CliStartErrEnvIncomplete)))
		return nil
	}

	if err := CheckDB(dsn); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrCheckDb), err.Error())))
		return nil
	}

	if err := CheckZentao(devopsURL); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrCheckZentao), err.Error())))
		return nil
	}

	// =========================================================
	// 阶段 3. 智能服务注册决策引擎
	// =========================================================

	// 3.1 获取当前二进制绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
		return nil
	}
	exePath, _ = filepath.EvalSymlinks(exePath)

	// 3.2 先探测服务是否已经在跑且健康（避免重复启动）
	healthURL := BuildHealthURL(host, port)

	// 预声明在 goto 之前，避免跳过变量声明
	var existing ServiceInfo

	if CheckHealthOnce(healthURL, 5*time.Second) {
		fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStartAlreadyRunning), port))
		fmt.Println()

		wantRestart, restartErr := confirmPrompt(t(i18n.CliStartAlreadyRunningRestart))

		if restartErr != nil || !wantRestart {
			fmt.Println(t(i18n.CliStartAlreadyRunningSkip))
			fmt.Println()
			printStartSummary(workDir, host, port)
			return nil
		}

		// 用户确认重启
		// fmt.Println()
		fmt.Println(t(i18n.CliStartServiceRestarting))
		if err := RestartService(); err != nil {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
			return nil
		}
		goto healthCheck
	}

	// 3.3 端口冲突检测（服务未运行时才需要检查）
	sshPort = envMap["GITFOX_SSH_PORT"]
	if sshPort == "" {
		sshPort = "22"
	}
	for _, checkPortNum := range [2]string{port, sshPort} {
		ps := CheckPort(checkPortNum)
		if ps.Unknown {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrPortConflictUnknown), checkPortNum)))
			return nil
		}
		if ps.InUse {
			fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrPortConflict), checkPortNum, ps.PID, ps.Process)))
			return nil
		}
	}

	// 3.4 检查系统中是否已存在服务注册
	existing = InspectService()

	if existing.Registered {
		if existing.Corrupted {
			// 文件损坏，直接覆盖重建
			fmt.Println(warnStyle.Render(t(i18n.CliStartServiceCorrupted)))
			fmt.Println()
		} else if existing.ExePath == exePath && existing.WorkDir == workDir {
			// 同一二进制 + 同一工作目录 → 直接 restart，跳过注册
			if err := RestartService(); err != nil {
				fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
				return nil
			}
			// 直接跳到阶段 4 健康检查
			goto healthCheck
		} else {
			// 路径不同 → 提示用户对比并确认是否覆盖
			fmt.Println(warnStyle.Render(t(i18n.CliStartServiceExistsDiff)))
			fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsOldExe), existing.ExePath))
			fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsNewExe), exePath))
			if existing.WorkDir != workDir {
				fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsOldDir), existing.WorkDir))
				fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsNewDir), workDir))
			}
			fmt.Println()

			overwrite, overwriteErr := confirmPrompt(t(i18n.CliStartServiceOverwrite))

			if overwriteErr != nil || !overwrite {
				// 用户选择保留现有注册 → 用旧配置 restart
				fmt.Println(t(i18n.CliStartServiceRestarting))
				if err := RestartService(); err != nil {
					fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
					return nil
				}
				goto healthCheck
			}
			fmt.Println()
			// 用户确认覆盖 → 先 stop 旧服务，再走全新注册
			_ = StopService()
		}
	}

	// 执行全新注册
	if err := RegisterService(exePath, workDir); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

healthCheck:

	// =========================================================
	// 阶段 4. 结果验证：存活状态探针轮询 (Health Check)
	// =========================================================
	// 服务虽然被挂载进了后台，但不代表内部启动未抛出 Panic。
	// 通过短时长循环测试 /public/health 接口验证内部状态正常。
	healthURL = BuildHealthURL(host, port)
	if !WaitForServiceUp(healthURL, 15) {
		fmt.Println(errStyle.Render(t(i18n.CliStartErrHealthTimeout)))
		fmt.Println()
		printFailureSummary()
		return nil
	}

	doneStyle := lipgloss.NewStyle()
	fmt.Println(doneStyle.Render("✓ " + t(i18n.CliStartSuccess)))
	fmt.Println()
	printStartSummary(workDir, host, port)

	return nil
}

// printStartSummary 在终端打印出美观的后台服务运行信息仪表盘，方便运维管理查询
func printStartSummary(workDir, host, port string) {
	t := clii18n.T

	var logCmd string
	switch runtime.GOOS {
	case "windows":
		logCmd = windowsLogCmd() // → windows.go
	case "linux":
		logCmd = "sudo journalctl -u gitfox -f"
	}

	webConsole := fmt.Sprintf("http://%s:%s", host, port)
	appLogsDir := filepath.Join(workDir, "logs")

	headerStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.NewStyle().
		Width(20)
	valueStyle := lipgloss.NewStyle()

	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	fmt.Println(headerStyle.Render(t(i18n.CliStartSummaryTitle)))
	fmt.Println(strings.Repeat("─", 60))
	printRow(t(i18n.CliStartSummaryAddress), webConsole)
	printRow(t(i18n.CliStartSummaryAppLog), appLogsDir)
	printRow(t(i18n.CliStartSummaryLogCmd), logCmd)
	fmt.Println(strings.Repeat("─", 60))
}

// printFailureSummary 打印健康检查失败的排错指南
func printFailureSummary() {
	t := clii18n.T

	var logCmd string
	var statusCmd string
	switch runtime.GOOS {
	case "windows":
		logCmd = windowsLogCmd()       // → windows.go
		statusCmd = windowsStatusCmd() // → windows.go
	case "linux":
		logCmd = "sudo journalctl -u gitfox -n 50 --no-pager"
		statusCmd = "sudo systemctl status gitfox"
	}

	headerStyle := lipgloss.NewStyle().Bold(true)
	labelStyle := lipgloss.NewStyle().Width(20)
	valueStyle := lipgloss.NewStyle()

	printRow := func(label, value string) {
		fmt.Printf("  %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	fmt.Println(headerStyle.Render(t(i18n.CliStartFailureTitle)))
	fmt.Println(strings.Repeat("─", 60))
	printRow(t(i18n.CliStartFailureCheckStatus), statusCmd)
	printRow(t(i18n.CliStartFailureCheckLog), logCmd)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
}
