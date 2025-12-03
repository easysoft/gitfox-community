package req

const (
	HeaderContentEncoding = "Content-Encoding"
)

type ContextKey int

const (
	authSessionKey ContextKey = iota
)

type RepoRef struct {
	RepoRef uint `uri:"repoRef" binding:"required"`
}

type RepoId struct {
	RepoID uint `uri:"repoID" binding:"required"`
}

type SpaceId struct {
	SpaceID uint `uri:"spaceID" binding:"required"`
}

type SpaceRepoRef struct {
	SpaceID uint   `uri:"spaceID" binding:"required"`
	RepoRef string `uri:"repoRef" binding:"required"`
}

const (
	HeaderParamGitProtocol       = "Git-Protocol"        // Git 协议版本，例如：git/2.0
	QueryParamIncludeCommit      = "includeCommit"       // 是否包含提交信息 默认 false
	QueryParamGitRef             = "gitRef"              // 仓库引用，例如：refs/heads/main 默认 refs/heads/main
	QueryParamFlattenDirectories = "flatten_directories" // 是否展平目录结构 默认 false

	PathParamTokenIdentifier = "token_identifier" // 令牌标识符

	PathParamPublicKeyIdentifier = "public_key_identifier"
)

const (
	QueryParamAccessToken   = "access_token"
	QueryParamIncludeCookie = "include_cookie"
)

type ListFilter struct {
	// 分页参数（对应 pagination.PageParam）
	Page     int `form:"page" binding:"omitempty,min=1"`             // 最小1
	PageSize int `form:"pageSize" binding:"omitempty,min=1,max=100"` // 1-100
}
