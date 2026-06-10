package svn

import "path/filepath"

// RepositoryPaths 包含 SVN 本地仓库和工作副本的文件系统路径。
type RepositoryPaths struct {
	// RepoPath 是本地 SVN 仓库路径 ({svnRoot}/repos/{encodedDir})。
	RepoPath string
	// WorkingCopyPath 是工作副本检出路径 ({svnRoot}/repos/{encodedDir}-wc)。
	WorkingCopyPath string
}

// ResolveRepositoryPaths 根据编码目录名拼装仓库路径与工作副本路径。
func (s *Service) ResolveRepositoryPaths(encodedDir string) RepositoryPaths {
	return RepositoryPaths{
		RepoPath:        filepath.Join(s.svnRoot, "repos", encodedDir),
		WorkingCopyPath: filepath.Join(s.svnRoot, "repos", encodedDir+"-wc"),
	}
}
