package install

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamInstall = "install"
)

func Register(app *kingpin.Application) {
	RegisterInstall(app)
	RegisterInit(app)
	RegisterStart(app)
}

func RegisterInstall(app *kingpin.Application) {
	app.Command(ParamInstall, clii18n.T(i18n.CliInstallDescription))
}
