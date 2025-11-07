package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	// repo_git

	// repo
	GetRepo(ctx context.Context, repo_ref string, allowedStates []constant.RepoState) (*repo.Repository, error)
	GetRepoCheckAccess(ctx context.Context, session *auth.Session, repo_ref string, allowedStates []constant.Permission) (*repo.Repository, error)
}
