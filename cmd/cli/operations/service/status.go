package service

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamStatus = "status"
)

// RegisterStatus 注册 `gitfox status` 子命令
func RegisterStatus(app *kingpin.Application) {
	app.Command(ParamStatus, clii18n.T(i18n.CliStatusDescription)).Action(runStatus)
}

// runStatus 是 `gitfox status` 的实际执行入口。
// 逐项检查依赖工具安装状态和 GitFox 服务运行状态，以状态报告的形式呈现。
func runStatus(_ *kingpin.ParseContext) error {
	t := clii18n.T
	okStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	separatorLine := "────────────────────────────────────────────────────────"

	printBanner(t(i18n.CliStatusWelcome))

	fmt.Println(headerStyle.Render("📋 " + t(i18n.CliStatusTitle)))
	fmt.Println(separatorLine)

	// =========================================================
	// 区块 1. 依赖环境检查
	// =========================================================
	fmt.Println(headerStyle.Render("  " + t(i18n.CliStatusSectionDeps)))
	fmt.Println()

	// --- Git ---
	if ok, ver, _ := CheckGit(); ok {
		fmt.Println(okStyle.Render("  ✓ ") + fmt.Sprintf(t(i18n.CliStatusGitOk), ver))
	} else {
		fmt.Println(errStyle.Render("  ✗ ") + t(i18n.CliStatusGitMissing))
	}

	// --- Docker ---
	if ok, ver, _ := CheckDocker(); ok {
		fmt.Println(okStyle.Render("  ✓ ") + fmt.Sprintf(t(i18n.CliStatusDockerOk), ver))
	} else {
		fmt.Println(errStyle.Render("  ✗ ") + t(i18n.CliStatusDockerMissing))
	}

	// --- Docker Compose ---
	if ok, ver, _ := CheckDockerCompose(); ok {
		fmt.Println(okStyle.Render("  ✓ ") + fmt.Sprintf(t(i18n.CliStatusComposeOk), ver))
	} else {
		fmt.Println(errStyle.Render("  ✗ ") + t(i18n.CliStatusComposeMissing))
	}

	fmt.Println()
	fmt.Println(separatorLine)

	// =========================================================
	// 区块 2. GitFox 服务状态检查
	// =========================================================
	fmt.Println(headerStyle.Render("  " + t(i18n.CliStatusSectionService)))
	fmt.Println()

	registered := isServiceRegistered()
	_, port, healthURL := readHealthURL()

	if !registered {
		fmt.Println(errStyle.Render("  ✗ ") + t(i18n.CliStatusServiceNotRegistered))
	} else if healthURL != "" {
		// 服务已注册，探测健康接口
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(healthURL)

		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			fmt.Println(okStyle.Render("  ✓ ") + t(i18n.CliStatusServiceRunning))
			fmt.Println(okStyle.Render("  ✓ ") + fmt.Sprintf(t(i18n.CliStatusServicePort), port))
		} else {
			if resp != nil {
				resp.Body.Close()
			}
			fmt.Println(errStyle.Render("  ✗ ") + fmt.Sprintf(t(i18n.CliStatusServiceRegistered), port))
		}
	} else {
		// 注册了但没有 env 配置可读取端口
		fmt.Println(errStyle.Render("  ✗ ") + fmt.Sprintf(t(i18n.CliStatusServiceRegistered), "?"))
	}

	fmt.Println()
	fmt.Println(separatorLine)

	// =========================================================
	// 区块 3. 端口占用检查
	// =========================================================
	fmt.Println(headerStyle.Render("  " + t(i18n.CliStatusSectionPorts)))
	fmt.Println()

	// 从 prod.env 读取端口，未配置则用默认值
	envMap, _ := godotenv.Read("prod.env")
	httpPort := envMap["GITFOX_HTTP_PORT"]
	sshPort := envMap["GITFOX_SSH_PORT"]
	if httpPort == "" {
		httpPort = "3052"
	}
	if sshPort == "" {
		sshPort = "22"
	}

	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
	for _, p := range []string{httpPort, sshPort} {
		ps := CheckPort(p)
		switch {
		case ps.Unknown:
			fmt.Println(warnStyle.Render("  ? ") + fmt.Sprintf(t(i18n.CliStatusPortUnknown), p))
		case ps.InUse:
			label := fmt.Sprintf(t(i18n.CliStatusPortInUse), p, ps.PID, ps.Process)
			fmt.Println(okStyle.Render("  ✓ ") + label)
		default:
			fmt.Println(errStyle.Render("  ✗ ") + fmt.Sprintf(t(i18n.CliStatusPortFree), p))
		}
	}

	fmt.Println()
	fmt.Println(separatorLine)
	fmt.Println()

	return nil
}

// isServiceRegistered 检查当前系统是否已经注册了 GitFox 守护进程服务
func isServiceRegistered() bool {
	switch runtime.GOOS {
	case "linux":
		// 检查 systemd unit 文件是否存在
		if _, err := os.Stat("/etc/systemd/system/gitfox.service"); err == nil {
			return true
		}
		// 或者通过 systemctl 判断
		return exec.Command("systemctl", "is-enabled", "--quiet", "gitfox").Run() == nil
	case "darwin":
		if _, err := os.Stat("/Library/LaunchDaemons/com.gitfox.service.plist"); err == nil {
			return true
		}
		return false
	case "windows":
		// 尝试查询服务名是否存在
		return exec.Command("sc", "query", "gitfox").Run() == nil
	default:
		return false
	}
}
