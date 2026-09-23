// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package service

import (
	"fmt"
	"os"

	"github.com/easysoft/gitfox-community/pkg/i18n"
	clii18n "github.com/easysoft/gitfox-community/pkg/i18n/cli"
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
	missing := Check(headerStyle, okStyle, errStyle, warnStyle)

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

	// 阻塞性依赖（Git / Docker）缺失时以非零退出码结束，便于脚本判断
	if missing {
		os.Exit(1)
	}
	return nil
}
