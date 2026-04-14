package main

import (
	"os"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/hooks"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/server"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cmd/cli/operations/service"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
	clii18n "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n/cli"

	"github.com/quicklyon/kingpin/v2"

	_ "gitfox.zcorp.cc/git/Pangu/gitfox-community/docs" // 导入生成的 Swagger 文档
)

const (
	application = "GitFox"
	description = "GitFox"
)

// @title           GitFox API
// @version         1.0
// @description     GitFox 平台 API 文档
// @termsOfService  http://swagger.io/terms/

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:3000
// @BasePath  /api/v2

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

const customUsageTemplate = `usage: GitFox [<flags>] <command> [<args> ...]

GitFox help

{{if .Context.SelectedCommand}}\
{{if len .Context.SelectedCommand.Commands}}\
{{range .Context.SelectedCommand.Commands}}{{if not .Hidden}}{{if ne .Name "help"}}{{printf "%-11s %s" .Name .Help}}
{{end}}{{end}}{{end}}\
{{end}}\
{{else}}\
{{range .App.Commands}}{{if not .Hidden}}{{if ne .Name "help"}}{{printf "%-11s %s" .Name .Help}}
{{end}}{{end}}{{end}}\
{{end}}\
`

func main() {
	// 初始化 CLI 多语言
	clii18n.Init()

	// 解析命令行参数
	args := getArguments()
	app := kingpin.New(application, description)
	app.UsageTemplate(customUsageTemplate)
	server.Register(app)
	hooks.Register(app)
	service.Register(app)
	kingpin.MustParse(app.Parse(args))
}

// getArguments 处理命令行参数，对来自 Git 钩子的参数进行适配转换。
func getArguments() []string {
	cmdName := os.Args[0]
	rawArgs := os.Args[1:]

	// 处理 Git 钩子参数：转换为 gitfox 兼容格式
	gitArgs, isGitHook := hook.SanitizeArgsForGit(cmdName, rawArgs)
	if isGitHook {
		// 前缀添加 hooks.ParamHooks 标识，便于后续解析为钩子操作
		return append([]string{hooks.ParamHooks}, gitArgs...)
	}

	// 非 Git 钩子场景，直接返回原始参数
	return rawArgs
}
