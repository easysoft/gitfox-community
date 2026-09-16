package hooks

import (
	"github.com/easysoft/gitfox-community/app/githook"
	"github.com/easysoft/gitfox-community/pkg/git/hook"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	clii18n "github.com/easysoft/gitfox-community/pkg/i18n/cli"
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
