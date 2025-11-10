package hooks

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/githook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	"github.com/quicklyon/kingpin/v2"
)

const (
	// ParamHooks defines the parameter for the git hooks sub-commands.
	ParamHooks = "hooks"
)

func Register(app *kingpin.Application) {
	subCmd := app.Command(ParamHooks, "manage git server hooks")
	hook.RegisterAll(subCmd, githook.LoadFromEnvironment)
}
