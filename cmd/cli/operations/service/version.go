package service

import (
	"encoding/json"
	"fmt"
	"os"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/common"
	"github.com/charmbracelet/lipgloss"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamVersion = "version"
)

// versionArgs 持有 version 子命令的运行时入参。
type versionArgs struct {
	jsonOutput bool
}

// RegisterVersion 注册 version 子命令
func RegisterVersion(app *kingpin.Application) {
	args := &versionArgs{}
	cmd := app.Command(ParamVersion, clii18n.T(i18n.CliVersionWelcome))
	cmd.Flag("json", clii18n.T(i18n.CliVersionJsonFlag)).
		BoolVar(&args.jsonOutput)
	cmd.Action(func(_ *kingpin.ParseContext) error {
		return runVersion(args)
	})
}

func runVersion(args *versionArgs) error {
	info := common.GetVersionInfo()

	// JSON 模式：机器可读，供 upgrade 等命令以稳定结构解析
	if args.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	// 默认模式：人类可读的对齐表格
	t := clii18n.T

	labelStyle := lipgloss.NewStyle().Width(20)
	valueStyle := lipgloss.NewStyle()

	printRow := func(label, value string) {
		fmt.Printf("%s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
	}

	printRow(t(i18n.CliVersionVersion), info.Version)
	printRow(t(i18n.CliVersionBuildDate), info.BuildDate)
	printRow(t(i18n.CliVersionGitBranch), info.GitBranch)
	printRow(t(i18n.CliVersionGitCommit), info.GitCommitHash)
	return nil
}
