package service

import (
	"fmt"

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
	errStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4500"))
	warnStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF8C00"))
	headerStyle := lipgloss.NewStyle().Bold(true)

	// ========== 阶段 0：环境依赖检查 ==========
	_ = Check(headerStyle, okStyle, errStyle, warnStyle)

	// ========== 阶段 1：GitFox 服务状态检查（端到端聚合） ==========
	fmt.Println()
	fmt.Println(headerStyle.Render(t(i18n.CliStatusSectionService)))

	hr := CheckGitfoxHealth()
	switch hr.Status {
	case GitfoxHealthOK:
		fmt.Println(okStyle.Render(t(i18n.CliStatusServiceRunning)))
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusServicePort), hr.Port)))
	case GitfoxHealthNotRegistered:
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceNotRegistered)))
	case GitfoxHealthCorrupted:
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceCorrupted)))
	case GitfoxHealthStopped:
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceStopped)))
	case GitfoxHealthConfigMissing:
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceConfigMissing)))
	default:
		// 兜底：未知状态当未运行
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceNotRunning)))
	}
	return nil
}
