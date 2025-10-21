package cli

import (
	"os"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/cli/operations/hooks"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git/hook"
)

// GetArguments 处理命令行参数，对来自 Git 钩子的参数进行适配转换。
func GetArguments() []string {
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