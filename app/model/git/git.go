package git

type Repository interface {
	GetGitUID() string
}

// CreateReadParams creates base read parameters for git read operations.
// IMPORTANT: repo is assumed to be not nil!
// CreateReadParams为git的读操作创建基本读参数。
// 重要：假定repo不是nil！
func CreateReadParams(repo Repository) ReadParams {
	return ReadParams{
		RepoUID: repo.GetGitUID(),
	}
}
