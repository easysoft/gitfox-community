package git

type RepoURI struct {
	RepoRef string `uri:"repo_ref" binding:"required"`
}
