package service

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	clii18n "github.com/easysoft/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamUninstall = "uninstall"
)

// RegisterUninstall 注册 `gitfox uninstall` 子命令到父应用中
func RegisterUninstall(app *kingpin.Application) {
	app.Command(ParamUninstall, clii18n.T(i18n.CliUninstallDescription)).Action(runUninstall)
}

// runUninstall 是 `gitfox uninstall` 的实际执行入口。
// 流程：sudo 校验 -> 确认对话框 -> stop -> disable -> rm -> daemon-reload -> 健康检查
func runUninstall(_ *kingpin.ParseContext) error {
	t := clii18n.T
	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00"))
	// successStyle := lipgloss.NewStyle().Bold(true)

	// printBanner(t(i18n.CliUninstallWelcome))

	// ========== 阶段 0. 权限校验 ==========
	if err := CheckSudo(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	// ========== 阶段 1. 二次确认 ==========
	confirm, confirmErr := confirmPrompt(t(i18n.CliUninstallConfirm))

	if confirmErr != nil || !confirm {
		fmt.Println(t(i18n.CliUninstallConfirmCancel))
		return nil
	}

	workDir, err := WorkDir()
	if err != nil {
		fmt.Println(errStyle.Render(fmt.Sprintf(t(i18n.CliUninstallFailed), err)))
		return nil
	}

	hc, hcErr := readHealthURL(workDir)
	if hcErr != nil {
		// 配置缺失不阻断 uninstall：仍要执行 stop + 系统注册移除。
		fmt.Println(warnStyle.Render(hcErr.Error()))
	}

	// ========== 阶段 2. 停止服务 ==========
	fmt.Println(t(i18n.CliStopStep))
	// 停止允许失败（服务可能本就没在运行）
	_ = StopService()

	// ========== 阶段 3. 从系统中彻底移除 ==========
	if err := uninstallService(); err != nil {
		fmt.Println(errStyle.Render(err.Error()))
		return nil
	}

	// ========== 阶段 4. 反向健康检查端口 ==========
	if hc.Port != "" {
		ps := CheckPort(hc.Port)
		if ps.InUse {
			fmt.Println(warnStyle.Render(t(i18n.CliStopHealthStillUp)))
			return nil
		}
	}

	doneStyle := lipgloss.NewStyle()
	fmt.Println(doneStyle.Render("✓ " + t(i18n.CliUninstallSuccess)))

	return nil
}

// uninstallService 根据操作系统执行完整的反向清理流程
func uninstallService() error {
	switch runtime.GOOS {
	case "windows":
		return uninstallWindows() // → windows.go
	case "linux":
		return uninstallLinux()
	default:
		return errors.New(clii18n.T(i18n.CliStartErrUnsupportedOs))
	}
}

func uninstallLinux() error {
	t := clii18n.T

	if _, err := exec.LookPath("systemctl"); err != nil {
		return errors.New(t(i18n.CliSystemctlMissing))
	}

	servicePath := "/etc/systemd/system/gitfox.service"

	// 1. disable: 拔除开机自启钩子
	//fmt.Println("  " + t(i18n.CliUninstallStepDisable))
	_ = exec.Command("systemctl", "disable", "gitfox").Run()

	// 2. rm: 粉碎服务定义文件
	//fmt.Println("  " + t(i18n.CliUninstallStepRemove))
	if err := os.Remove(servicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf(t(i18n.CliUninstallFailed), err)
	}

	// 3. daemon-reload: 通知 systemd 刷新缓存
	//fmt.Println("  " + t(i18n.CliUninstallStepReload))
	_ = exec.Command("systemctl", "daemon-reload").Run()

	return nil
}
