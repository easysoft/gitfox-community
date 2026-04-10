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
	okStyle := lipgloss.NewStyle().Bold(true)
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))

	printBanner(t(i18n.CliStatusWelcome))
	//fmt.Println(headerStyle.Render(t(i18n.CliStatusTitle)))

	// =========================================================
	// 区块 1. 依赖环境检查
	// =========================================================
	fmt.Println(headerStyle.Render(t(i18n.CliStatusSectionDeps)))

	// --- Git ---
	if ok, ver, _ := CheckGit(); ok {
		fmt.Println(okStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStatusGitOk), ver))
	} else {
		fmt.Println(errStyle.Render("✗ ") + t(i18n.CliStatusGitMissing))
		fmt.Println(lipgloss.NewStyle().Render("      Install: https://git-scm.com/downloads"))
	}

	// --- Docker ---
	if ok, ver, err := CheckDocker(); ok && err == nil {
		fmt.Println(okStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStatusDockerOk), ver))
	} else if ok && err != nil {
		// 已安装但 daemon 未运行
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
		fmt.Println(warnStyle.Render("✗ ") + err.Error())
	} else {
		fmt.Println(errStyle.Render("✗ ") + t(i18n.CliStatusDockerMissing))
		fmt.Println(lipgloss.NewStyle().Render("      Install: https://docs.docker.com/get-docker/"))
	}

	// --- Docker Compose ---
	if ok, ver, _ := CheckDockerCompose(); ok {
		fmt.Println(okStyle.Render("✓ ") + fmt.Sprintf(t(i18n.CliStatusComposeOk), ver))
	} else {
		fmt.Println(errStyle.Render("✗ ") + t(i18n.CliStatusComposeMissing))
		fmt.Println(lipgloss.NewStyle().Render("      Install: https://docs.docker.com/compose/install/"))
	}

	fmt.Println()
	// fmt.Println(separatorLine)

	// =========================================================
	// 区块 2. GitFox 服务状态检查
	// =========================================================
	fmt.Println(headerStyle.Render("  " + t(i18n.CliStatusSectionService)))

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
	case "windows":
		// isServiceRegisteredWindows 已迁移至 windows.go
		return isServiceRegisteredWindows()
	default:
		return false
	}
}
