package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/huh"
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
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)

	printBanner(t(i18n.CliStartWelcome))

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
	fmt.Println("🔍 " + t(i18n.CliStartStepCheckDeps))
	if ok, ver, err := CheckGit(); !ok || err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	} else {
		fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStartCheckGitOk), ver))
	}

	if ok, ver, err := CheckDocker(); !ok || err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	} else {
		fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStartCheckDockerOk), ver))
	}
	fmt.Println()

	// =========================================================
	// 阶段 2. 读取环境配制检查 (识别是否允许提供服务)
	// =========================================================
	// 查找并校验工作目录下是否存在必要的 prod.env
	envPath := filepath.Join(workDir, "prod.env")
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

	fmt.Println()
	fmt.Println("────────────────────────────────────────────────────────")

	// 强制校验：如果这四个核心要素只要有一个为空，说明配置文件已被破坏，强制重新 init
	if port == "" || host == "" || dsn == "" || devopsURL == "" {
		fmt.Println(errStyle.Render(t(i18n.CliStartErrEnvIncomplete)))
		return nil
	}

	fmt.Println("🔍 " + t(i18n.CliStartStepCheckDb))
	if err := CheckDB(dsn); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrCheckDb), err.Error())))
		return nil
	}
	fmt.Println(successStyle.Render("✓ ") + t(i18n.CliStartSuccessCheckDb))

	fmt.Println("🔍 " + t(i18n.CliStartStepCheckZentao))
	if err := CheckZentao(devopsURL); err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartErrCheckZentao), err.Error())))
		return nil
	}
	fmt.Println(successStyle.Render("✓ ") + t(i18n.CliStartSuccessCheckZentao))
	fmt.Println("────────────────────────────────────────────────────────")
	// =========================================================
	// 阶段 3. 智能服务注册决策引擎
	// =========================================================
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)

	// 3.1 获取当前二进制绝对路径
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
		return nil
	}
	exePath, _ = filepath.EvalSymlinks(exePath)

	// 3.2 先探测服务是否已经在跑且健康（避免重复启动）
	healthURL := fmt.Sprintf("http://%s:%s/public/health", host, port)
	preClient := http.Client{Timeout: 2 * time.Second}

	// 预声明在 goto 之前，避免跳过变量声明
	var existing ServiceInfo

	if resp, err := preClient.Get(healthURL); err == nil && resp.StatusCode == 200 {
		resp.Body.Close()
		fmt.Println(successStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStartAlreadyRunning), port))
		fmt.Println()

		var wantRestart bool
		restartErr := huh.NewConfirm().
			Title(t(i18n.CliStartAlreadyRunningRestart)).
			Affirmative(t(i18n.CliStartAlreadyRunningRestartYes)).
			Negative(t(i18n.CliStartAlreadyRunningRestartNo)).
			Value(&wantRestart).
			Run()

		if restartErr != nil || !wantRestart {
			fmt.Println(t(i18n.CliStartAlreadyRunningSkip))
			fmt.Println()
			printStartSummary(workDir, host, port)
			return nil
		}

		// 用户确认重启
		fmt.Println()
		fmt.Println("🔄 " + t(i18n.CliStartServiceRestarting))
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
			fmt.Println(errStyle.Render("❌ " + fmt.Sprintf(t(i18n.CliStartErrPortConflictUnknown), checkPortNum)))
			return nil
		}
		if ps.InUse {
			fmt.Println(errStyle.Render("❌ " + fmt.Sprintf(t(i18n.CliStartErrPortConflict), checkPortNum, ps.PID, ps.Process)))
			return nil
		}
	}

	// 3.4 检查系统中是否已存在服务注册
	existing = InspectService()

	if existing.Registered {
		if existing.Corrupted {
			// 文件损坏，直接覆盖重建
			fmt.Println(warnStyle.Render("⚠️  " + t(i18n.CliStartServiceCorrupted)))
			fmt.Println()
		} else if existing.ExePath == exePath && existing.WorkDir == workDir {
			// 同一二进制 + 同一工作目录 → 直接 restart，跳过注册
			fmt.Println("🔄 " + t(i18n.CliStartServiceExistsSame))
			if err := RestartService(); err != nil {
				fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliStartFailed), err)))
				return nil
			}
			// 直接跳到阶段 4 健康检查
			goto healthCheck
		} else {
			// 路径不同 → 提示用户对比并确认是否覆盖
			fmt.Println(warnStyle.Render("⚠️  " + t(i18n.CliStartServiceExistsDiff)))
			fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsOldExe), existing.ExePath))
			fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsNewExe), exePath))
			if existing.WorkDir != workDir {
				fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsOldDir), existing.WorkDir))
				fmt.Println(fmt.Sprintf(t(i18n.CliStartServiceExistsNewDir), workDir))
			}
			fmt.Println()

			var overwrite bool
			overwriteErr := huh.NewConfirm().
				Title(t(i18n.CliStartServiceOverwrite)).
				Affirmative(t(i18n.CliStartServiceOverwriteYes)).
				Negative(t(i18n.CliStartServiceOverwriteNo)).
				Value(&overwrite).
				Run()

			if overwriteErr != nil || !overwrite {
				// 用户选择保留现有注册 → 用旧配置 restart
				fmt.Println("🔄 " + t(i18n.CliStartServiceRestarting))
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
	} else {
		// 首次注册 → 弹出确认框
		var confirm bool
		confirmErr := huh.NewConfirm().
			Title(t(i18n.CliStartConfirmRegister)).
			Affirmative(t(i18n.CliStartConfirmYes)).
			Negative(t(i18n.CliStartConfirmNo)).
			Value(&confirm).
			Run()

		if confirmErr != nil || !confirm {
			fmt.Println(t(i18n.CliStartConfirmCancel))
			return nil
		}
		fmt.Println()
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
	fmt.Println()
	fmt.Println("⏳ " + fmt.Sprintf(t(i18n.CliStartStepHealthCheck), port))
	healthURL = fmt.Sprintf("http://%s:%s/public/health", host, port)

	client := http.Client{Timeout: 2 * time.Second}
	isHealthy := false
	for i := 0; i < 15; i++ {
		resp, err := client.Get(healthURL)
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			isHealthy = true
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}

	if !isHealthy {
		fmt.Println(errStyle.Render("❌ " + t(i18n.CliStartErrHealthTimeout)))
		fmt.Println()
		printFailureSummary()
		return nil
	}

	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF82"))
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
		logCmd = t(i18n.CliStartSummaryLogCmdWindows)
	case "linux":
		logCmd = "sudo journalctl -u gitfox -f"
	case "darwin":
		logCmd = t(i18n.CliStartSummaryLogCmdMac)
	}

	webConsole := fmt.Sprintf("http://%s:%s", host, port)
	appLogsDir := filepath.Join(workDir, "logs")

	fmt.Println("  ────────────────────────────────────────────────────────")
	fmt.Println("  " + t(i18n.CliStartSummaryTitle))
	fmt.Println("  ────────────────────────────────────────────────────────")
	fmt.Printf("  🌐 %-16s : %s\n", t(i18n.CliStartSummaryAddress), webConsole)
	fmt.Printf("  📁 %-16s : %s\n", t(i18n.CliStartSummaryAppLog), appLogsDir)
	fmt.Println()
	fmt.Println("  " + t(i18n.CliStartSummaryManageTitle))
	fmt.Printf("  📖 %-16s : %s\n", t(i18n.CliStartSummaryLogCmd), logCmd)
	fmt.Println("  ────────────────────────────────────────────────────────")
	fmt.Println()
}

// printFailureSummary 打印健康检查失败的排错指南
func printFailureSummary() {
	t := clii18n.T

	var logCmd string
	var statusCmd string
	switch runtime.GOOS {
	case "windows":
		logCmd = t(i18n.CliStartSummaryLogCmdWindows)
		statusCmd = "sc query gitfox"
	case "linux":
		logCmd = "sudo journalctl -u gitfox -n 50 --no-pager"
		statusCmd = "sudo systemctl status gitfox"
	case "darwin":
		logCmd = t(i18n.CliStartSummaryLogCmdMac)
		statusCmd = "sudo launchctl list | grep gitfox"
	}

	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)

	fmt.Println(errStyle.Render("  ────────────────────────────────────────────────────────"))
	fmt.Println(errStyle.Render("  " + t(i18n.CliStartFailureTitle)))
	fmt.Println(errStyle.Render("  ────────────────────────────────────────────────────────"))
	fmt.Printf("  🔍 %-16s : %s\n", t(i18n.CliStartFailureCheckStatus), statusCmd)
	fmt.Printf("  📖 %-16s : %s\n", t(i18n.CliStartFailureCheckLog), logCmd)
	fmt.Println(errStyle.Render("  ────────────────────────────────────────────────────────"))
	fmt.Println()
}
