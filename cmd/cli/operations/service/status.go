package service

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/huh/spinner"
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

	// =========================================================
	// 区块 1. 依赖环境检查（带 spinner loading）
	// =========================================================
	var gitOk bool
	var gitVer string
	var gitErr error
	var dockerOk bool
	var dockerVer string
	var dockerErr error
	var svnOk bool
	var svnVer string
	var svnErr error

	_ = spinner.New().
		Title(t(i18n.CliStatusChecking)).
		Action(func() {
			gitOk, gitVer, gitErr = CheckGit()
			dockerOk, dockerVer, dockerErr = CheckDocker()
			svnOk, svnVer, svnErr = CheckSubversion()
		}).Run()

	fmt.Println(headerStyle.Render(t(i18n.CliStatusSectionDeps)))

	// --- Git ---
	if gitOk && gitErr == nil {
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusGitOk), gitVer)))
	} else if gitOk && gitErr != nil {
		// 已安装但版本过低或解析失败
		fmt.Println(warnStyle.Render(gitErr.Error()))
	} else {
		fmt.Println(errStyle.Render(t(i18n.CliStatusGitMissing)))
	}

	// --- Docker ---
	if dockerOk && dockerErr == nil {
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusDockerOk), dockerVer)))
	} else if dockerOk && dockerErr != nil {
		// 已安装但 daemon 未运行
		fmt.Println(warnStyle.Render(dockerErr.Error()))
	} else {
		fmt.Println(errStyle.Render(t(i18n.CliStatusDockerMissing)))
	}

	// --- Subversion (非阻塞，仅警告) ---
	if svnOk && svnErr == nil {
		fmt.Println(okStyle.Render(fmt.Sprintf(t(i18n.CliStatusSvnOk), svnVer)))
	} else if svnOk && svnErr != nil {
		// 已安装但版本过低或执行失败：影响关联 SVN 代码库功能
		fmt.Println(warnStyle.Render(svnErr.Error()))
	} else {
		fmt.Println(warnStyle.Render(t(i18n.CliStatusSvnMissing)))
	}

	// =========================================================
	// 区块 2. GitFox 服务状态检查（端到端聚合）
	// =========================================================
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
		// 把具体读取错误带出，便于排查（prod.env 路径、权限等）
		if hr.Err != nil {
			fmt.Println(warnStyle.Render(hr.Err.Error()))
		} else {
			fmt.Println(warnStyle.Render(t(i18n.CliStatusServiceNotReady)))
		}
	case GitfoxHealthAddressMissing, GitfoxHealthUnreachable:
		fmt.Println(warnStyle.Render(t(i18n.CliStatusServiceNotReady)))
	default:
		// 兜底：未知状态当未运行
		fmt.Println(errStyle.Render(t(i18n.CliStatusServiceNotRunning)))
	}
	return nil
}
