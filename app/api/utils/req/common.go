package req

const (
	HeaderContentEncoding = "Content-Encoding"
)

type ContextKey int

const (
	authSessionKey ContextKey = iota
)

type RepoRef struct {
	RepoRef string `uri:"repoRef" binding:"required"`
}

type RepoId struct {
	RepoID int64  `uri:"repoId" binding:"required"`
	Other  string `uri:"otherPath"`
}
