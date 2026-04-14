package hooks

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"
	"github.com/quicklyon/kingpin/v2"
)

const (
	// ParamHooks defines the parameter for the git hooks sub-commands.
	ParamHooks = "hooks"
)

func Register(app *kingpin.Application) {
	subCmd := app.Command(ParamHooks, clii18n.T(i18n.CliHooksDescription)).Hidden()
	hook.RegisterAll(subCmd, githook.LoadFromEnvironment)
}
