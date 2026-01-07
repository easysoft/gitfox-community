package pullrequest

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	GetMetadata(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint) (*repoModel.DiffStats, error)
	GetDiffStats(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.DiffStats, error)
	MergeCheck(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.MergeCheck, error)
}
