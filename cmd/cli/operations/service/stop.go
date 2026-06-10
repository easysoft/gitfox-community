package service

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamStop = "stop"
)

// RegisterStop 注册 `gitfox stop` 子命令到父应用中
func RegisterStop(app *kingpin.Application) {
	app.Command(ParamStop, clii18n.T(i18n.CliStopDescription)).Action(runStop)
}

// StopService 根据当前操作系统停止 GitFox 守护进程服务。
// 返回 error 表示停止失败（包括服务未注册的情况）。
func StopService() error {
	switch runtime.GOOS {
	case "windows":
		return stopWindows() // → windows.go
	case "linux":
		return stopLinux()
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

// stopWindows 已迁移至 windows.go

func stopLinux() error {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New(clii18n.T(i18n.CliStartErrSystemctlMissing))
	}

	// 检查服务是否存在
	if err := exec.Command("systemctl", "is-active", "--quiet", "gitfox").Run(); err != nil {
		return errors.New(clii18n.T(i18n.CliStopErrNotRunning))
	}

	if err := exec.Command("systemctl", "stop", "gitfox").Run(); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliStopFailed), err)
	}
	return nil
}

// runStop 是 `gitfox stop` 的实际执行入口处理函数。
func runStop(_ *kingpin.ParseContext) error {
	t := clii18n.T
	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00"))
	// successStyle := lipgloss.NewStyle().Bold(true)

	// printBanner(t(i18n.CliStopWelcome))

	// ========== 阶段 0. 权限校验 ==========
	if err := CheckSudo(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	_, port, healthURL := readHealthURL()

	// ========== 阶段 1. 执行停止 ==========
	fmt.Println(t(i18n.CliStopStep))
	if err := StopService(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	// ========== 阶段 2. 反向健康检查：确认端口不再响应 ==========
	if port != "" {
		if CheckPortDown(healthURL, 5) {
			fmt.Println(warnStyle.Render(t(i18n.CliStopHealthStillUp)))
		}
	}

	doneStyle := lipgloss.NewStyle()
	fmt.Println(doneStyle.Render("✓ " + t(i18n.CliStopSuccess)))

	return nil
}

// readHealthURL 从 prod.env 读取 host 和 port，拼装健康检查 URL。
// 优先使用 GITFOX_SERVER_HOST，兜底 127.0.0.1。
func readHealthURL() (host, port, url string) {
	envMap, err := godotenv.Read(filepath.Join(".", "prod.env"))
	if err != nil {
		return "127.0.0.1", "", ""
	}
	port = envMap["GITFOX_HTTP_PORT"]
	host = envMap["GITFOX_SERVER_HOST"]
	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		return host, port, ""
	}
	url = fmt.Sprintf("http://%s:%s/public/health", host, port)
	return host, port, url
}
