package repo

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)


type RepoInterface interface {
	Create(ctx context.Context, req *repoModel.CreateRepoReq) (*repoModel.CreateRepoResp, error)
	GetRepo(ctx context.Context, repoRef string, allowedStates []constant.RepoState) (*repo.Repository, error)
	GetRepoCheckAccess(ctx context.Context, session *auth.Session, repoRef string, allowedStates []constant.Permission) (*repo.Repository, error)

}
