// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
