package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
)

type Interface interface {
	Create(ctx context.Context, req *repoModel.CreateRepoReq) (*repoModel.CreateRepoResp, error)
	GetRepo(ctx context.Context, repoRef string, allowedStates []constant.RepoState) (*repo.Repository, error)
	GetContent(ctx context.Context, session *auth.Session, req *repoModel.GetContentReq) (*repoModel.GetContentOutput, error)
	GetRepoCheckAccess(ctx context.Context, session *auth.Session, repoRef string, allowedStates []constant.Permission) (*repo.Repository, error)
}
