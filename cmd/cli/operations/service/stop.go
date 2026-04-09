package service

import (
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

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
		return stopWindows()
	case "linux":
		return stopLinux()
	case "darwin":
		return stopMac()
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

func stopWindows() error {
	if _, err := exec.LookPath("nssm"); err != nil {
		return errors.New(clii18n.T(i18n.CliStartErrNssmMissing))
	}
	if err := exec.Command("nssm", "stop", "gitfox").Run(); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliStopFailed), err)
	}
	return nil
}

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

func stopMac() error {
	if _, err := exec.LookPath("launchctl"); err != nil {
		return errors.New(clii18n.T(i18n.CliStartErrLaunchctlMissing))
	}

	plistPath := "/Library/LaunchDaemons/com.gitfox.service.plist"
	if err := exec.Command("launchctl", "unload", "-w", plistPath).Run(); err != nil {
		return fmt.Errorf(clii18n.T(i18n.CliStopFailed), err)
	}
	return nil
}

// runStop 是 `gitfox stop` 的实际执行入口处理函数。
func runStop(_ *kingpin.ParseContext) error {
	t := clii18n.T
	errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4500")).Bold(true)
	// successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000")).Bold(true)

	printBanner(t(i18n.CliStopWelcome))

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
		//fmt.Println(fmt.Sprintf(t(i18n.CliStopHealthChecking), port))

		client := http.Client{Timeout: 2 * time.Second}
		stillUp := false
		for i := 0; i < 5; i++ {
			resp, err := client.Get(healthURL)
			if err != nil {
				// 连接被拒绝 = 服务确实已停止
				break
			}
			if resp != nil {
				resp.Body.Close()
			}
			if i == 4 {
				stillUp = true
			}
			time.Sleep(1 * time.Second)
		}

		if stillUp {
			warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)
			fmt.Println(warnStyle.Render(t(i18n.CliStopHealthStillUp)))
		}
		// else {
		// 	fmt.Println(successStyle.Render("✓ ") + t(i18n.CliStopHealthConfirmed))
		// }
	}

	fmt.Println()
	doneStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#000000"))
	fmt.Println(doneStyle.Render("✓ " + t(i18n.CliStopSuccess)))
	fmt.Println()

	return nil
}

// readPortFromEnv 尝试从当前目录的 prod.env 中读取端口号
func readPortFromEnv() string {
	envMap, err := godotenv.Read(filepath.Join(".", "prod.env"))
	if err != nil {
		return ""
	}
	return envMap["GITFOX_HTTP_PORT"]
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
