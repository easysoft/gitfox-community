package service

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/quicklyon/kingpin/v2"
)

// RegisterVersion 注册 version 子命令
func RegisterVersion(app *kingpin.Application) {
	cmd := app.Command("version", clii18n.T(i18n.CliVersionWelcome))
	cmd.Action(runVersion)
}

func runVersion(_ *kingpin.ParseContext) error {
	t := clii18n.T

	printBanner(t(i18n.CliVersionWelcome))

	labelStyle := lipgloss.NewStyle().
		Width(20)

	valueStyle := lipgloss.NewStyle()

	printRow := func(label, value string) {
		fmt.Printf("%s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	// ── 构建信息区块 ──────────────────────────────────────────
	// fmt.Println(separatorLine)
	// fmt.Println()

	version := common.Version
	if version == "" {
		version = "0.0.1"
	}
	buildDate := common.BuildDate
	if buildDate == "" {
		buildDate = "unknown"
	}
	gitBranch := common.GitBranch
	if gitBranch == "" {
		gitBranch = "unknown"
	}
	gitCommit := common.GitCommitHash
	if gitCommit == "" {
		gitCommit = "unknown"
	}

	printRow(t(i18n.CliVersionVersion), version)
	printRow(t(i18n.CliVersionBuildDate), buildDate)
	printRow(t(i18n.CliVersionGitBranch), gitBranch)
	printRow(t(i18n.CliVersionGitCommit), gitCommit)

	// fmt.Println()
	// fmt.Println(separatorLine)

	// // ── 服务状态区块 ──────────────────────────────────────────
	// fmt.Println()

	// // 注册状态
	// registered := isServiceRegistered()
	// regLabel := labelStyle.Render(t(i18n.CliVersionServiceRegistered))
	// if registered {
	// 	fmt.Println("  " + regLabel + "  " + okStyle.Render("✓ "+t(i18n.CliVersionServiceStatusRegistered)))
	// } else {
	// 	fmt.Println("  " + regLabel + "  " + errStyle.Render("✗ "+t(i18n.CliVersionServiceStatusNotRegistered)))
	// }

	// // 运行状态（通过健康接口探测）
	// _, _, healthURL := readHealthURL()
	// runtimeLabel := labelStyle.Render(t(i18n.CliVersionRuntimeStatus))
	// if healthURL != "" {
	// 	client := http.Client{Timeout: 2 * time.Second}
	// 	resp, err := client.Get(healthURL)
	// 	if err == nil && resp.StatusCode == 200 {
	// 		resp.Body.Close()
	// 		fmt.Println("  " + runtimeLabel + "  " + okStyle.Render("✓ "+t(i18n.CliVersionRuntimeRunning)))
	// 	} else {
	// 		if resp != nil {
	// 			resp.Body.Close()
	// 		}
	// 		fmt.Println("  " + runtimeLabel + "  " + errStyle.Render("✗ "+t(i18n.CliVersionRuntimeStopped)))
	// 	}
	// } else {
	// 	fmt.Println("  " + runtimeLabel + "  " + errStyle.Render("✗ "+t(i18n.CliVersionRuntimeStopped)))
	// }

	// fmt.Println()
	// fmt.Println(separatorLine)
	// fmt.Println()

	return nil
}
