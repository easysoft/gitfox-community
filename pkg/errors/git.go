package errors

import "github.com/easysoft/gitfox-community/pkg/i18n"

// Git related errors
var (
	// ErrGitRepoNotFound 表示仓库不存在
	ErrGitRepoNotFound = WithI18nKey(GitError.New("repository not found"), i18n.GitRepoNotFound)

	// ErrGitRefNotFound 表示引用不存在（分支、标签等）
	ErrGitRefNotFound       = WithI18nKey(GitError.New("reference not found"), i18n.GitRefNotFound)
	ErrGitBranchNotFound    = WithI18nKey(GitError.New("branch not found"), i18n.GitBranchNotFound)
	ErrGitTagNotFound       = WithI18nKey(GitError.New("tag not found"), i18n.GitTagNotFound)
	ErrGitRefFirstNotFound  = WithI18nKey(GitError.New("first reference not found"), i18n.GitRefNotFound)
	ErrGitRefSecondNotFound = WithI18nKey(GitError.New("second reference not found"), i18n.GitRefNotFound)

	// ErrGitRefAlreadyExists 表示引用已存在
	ErrGitRefAlreadyExists       = WithI18nKey(GitError.New("reference already exists"), i18n.GitRefAlreadyExists)
	ErrGitBranchRefAlreadyExists = WithI18nKey(GitError.New("branch already exists"), i18n.GitBranchAlreadyExists)
	ErrGitBranchRefConflict      = WithI18nKey(GitError.New("branch ref conflict"), i18n.GitBranchRefConflict)
	ErrGitTagRefAlreadyExists    = WithI18nKey(GitError.New("tag already exists"), i18n.GitTagAlreadyExists)

	// ErrGitInvalidRef 表示无效的引用
	ErrGitInvalidRef = WithI18nKey(GitError.New("invalid reference"), i18n.GitInvalidRef)

	// ErrGitAmbiguousRef 表示引用不明确
	ErrGitAmbiguousRef = WithI18nKey(GitError.New("ambiguous reference"), i18n.GitAmbiguousRef)

	// ErrGitEmptyRepo 表示空仓库
	ErrGitEmptyRepo = WithI18nKey(GitError.New("repository is empty"), i18n.GitEmptyRepo)

	// ErrGitPushRejected 表示推送被拒绝
	ErrGitPushRejected = WithI18nKey(GitError.New("push rejected"), i18n.GitPushRejected)

	// ErrGitPushOutOfDate 表示推送过期（需要先拉取）
	ErrGitPushOutOfDate = WithI18nKey(GitError.New("push out of date"), i18n.GitPushOutOfDate)

	// ErrGitMergeMethodUnsupported 表示不支持的合并方法
	ErrGitMergeMethodUnsupported = WithI18nKey(GitError.New("unsupported merge method"), i18n.GitInvalidMergeMethod)

	// ErrGitMergeConflict 表示合并冲突
	ErrGitMergeConflict                = WithI18nKey(GitError.New("merge conflict"), i18n.GitMergeConflict)
	ErrGitMergeCheckBaseBranchNotFound = WithI18nKey(GitError.New("either base branch or commit SHA is mandatory"), i18n.GitMergeCheckBaseBranchNotFound)
	ErrGitMergeCheckHeadBranchNotFound = WithI18nKey(GitError.New("head branch is mandatory"), i18n.GitMergeCheckHeadBranchNotFound)
	ErrGitMergeCheckRefTypeNotFound    = WithI18nKey(GitError.New("ref name has to be provided if type is defined"), i18n.GitMergeCheckRefTypeNotFound)
	ErrGitMergeCheckNoCommits          = WithI18nKey(GitError.New("head branch doesn't contain any new commits"), i18n.GitMergeCheckNoCommits)

	// ErrGitUnrelatedHistories 表示不相关的历史记录
	ErrGitUnrelatedHistories = WithI18nKey(GitError.New("unrelated histories"), i18n.GitUnrelatedHistories)

	// ErrGitPathNotFound 表示路径不存在
	ErrGitPathNotFound = WithI18nKey(GitError.New("path not found"), i18n.GitPathNotFound)

	// ErrGitInvalidPath 表示无效的路径
	ErrGitInvalidPath = WithI18nKey(GitError.New("invalid path"), i18n.GitInvalidPath)

	// ErrGitObjectNotFound 表示对象不存在
	ErrGitObjectNotFound = WithI18nKey(GitError.New("object not found"), i18n.GitObjectNotFound)

	// ErrGitInvalidSignature 表示无效的签名
	ErrGitInvalidSignature = WithI18nKey(GitError.New("invalid signature"), i18n.GitInvalidSignature)

	// ErrGitCommandFailed 表示 Git 命令执行失败
	ErrGitCommandFailed = WithI18nKey(GitError.New("git command failed"), i18n.GitExecError)

	// ErrGitHookFailed 表示 Git 钩子执行失败
	ErrGitHookFailed = WithI18nKey(GitError.New("git hook failed"), i18n.GitHookFailed)

	// ErrGitTimeout 表示 Git 操作超时
	ErrGitTimeout = WithI18nKey(GitError.New("git operation timeout"), i18n.GitTimeout)

	// ErrGitUnsupportedOption 表示 Git 版本不支持某个参数
	ErrGitUnsupportedOption = WithI18nKey(GitError.New("unsupported git option"), i18n.GitUnsupportedOption)
)
