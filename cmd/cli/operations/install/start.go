package install

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	ParamStart = "start"
)

func RegisterStart(app *kingpin.Application) {
	app.Command(ParamStart, clii18n.T(i18n.CliInstallStartDescription))
}
