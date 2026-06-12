package importer

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Run(ctx context.Context, provider repoModel.Provider, repo *repoModel.Repo, cloneURL string, mirror bool, noticeLang string) error
	Create(ctx context.Context, session *auth.Session, req *repoModel.Repo) (*repoModel.Repo, error)
	LoadRepositoryFromProvider(ctx context.Context, provider repoModel.Provider, slug string) (RepositoryInfo, repoModel.Provider, error)
	GetProgress(ctx context.Context, repoID uint) (repoModel.Progress, error)
	CreateMirrorSyncJob(ctx context.Context, repoID int64) error
	GetMirrorSyncProgress(ctx context.Context, repoID int64) (repoModel.Progress, error)
}
