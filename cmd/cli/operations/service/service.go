package service

import (
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/charmbracelet/lipgloss"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamInstall = "install"
)

func Register(app *kingpin.Application) {
	RegisterInstall(app)
	RegisterInit(app)
	RegisterStart(app)
	RegisterStop(app)
	RegisterUninstall(app)
	RegisterStatus(app)
	RegisterVersion(app)
}

func RegisterInstall(app *kingpin.Application) {
	app.Command(ParamInstall, clii18n.T(i18n.CliInstallDescription))
}

// printBanner 打印统一风格的欢迎横幅
func printBanner(text string) {
	bannerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 2)

	fmt.Println()
	fmt.Println(bannerStyle.Render(text))
	fmt.Println()
}

