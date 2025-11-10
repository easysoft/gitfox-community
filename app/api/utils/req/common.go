package req

const (
	HeaderContentEncoding = "Content-Encoding"
)

type ContextKey int

const (
	authSessionKey ContextKey = iota
)

type RepoURI struct {
	RepoRef string `uri:"repo_ref" binding:"required"`
}
