package git

import "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"

// ReadParams contains the base parameters for read operations.
// 读取操作的基本参数
type ReadParams struct {
	RepoUID             string
	AlternateObjectDirs []string
}

func (p ReadParams) Validate() error {
	if p.RepoUID == "" {
		return errors.NotFound.New("repository id cannot be empty")
	}
	return nil
}
