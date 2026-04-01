package install

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamInit = "init"
)

func RegisterInit(app *kingpin.Application) {
	app.Command(ParamInit, clii18n.T(i18n.CliInstallInitDescription))
}
