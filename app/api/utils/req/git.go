package req

const (
	HeaderParamGitProtocol       = "Git-Protocol"        // Git 协议版本，例如：git/2.0
	QueryParamIncludeCommit      = "include_commit"      // 是否包含提交信息 默认 false
	QueryParamGitRef             = "git_ref"             // 仓库引用，例如：refs/heads/main 默认 refs/heads/main
	QueryParamFlattenDirectories = "flatten_directories" // 是否展平目录结构 默认 false
)
