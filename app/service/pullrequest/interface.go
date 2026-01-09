package pullrequest

import (
	"context"
	"net/http"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/api/utils/resp"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/commit"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
)

type Interface interface {
	GetMetadata(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint) (*repoModel.DiffStats, error)
	GetDiffStats(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.DiffStats, error)
	MergeCheck(ctx context.Context, session *auth.Session, repoID uint, pathToPath string) (*repoModel.MergeCheck, error)

	RawDiff(ctx context.Context, w http.ResponseWriter, session *auth.Session, repoID uint, pullreqID uint, setSHAs func(sourceSHA, mergeBaseSHA string), files ...repoModel.FileDiff) error
	Diff(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, setSHAs func(sourceSHA, mergeBaseSHA string), includePatch bool, files ...repoModel.FileDiff) (resp.Stream[*git.FileDiff], error)
	Commits(ctx context.Context, session *auth.Session, repoID uint, pullreqID uint, page, limit uint) ([]commit.Commit, int, error)
}
