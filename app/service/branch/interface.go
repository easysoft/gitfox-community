package branch

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, req branch.ReqForCreate, session *auth.Session, repo *repo.Repo) (*branch.Branch, error)
	List(ctx context.Context, req branch.ReqFilter, repo *repo.Repo) ([]branch.Branch, error)
	ListPrefixes(ctx context.Context, repo repo.Repo) ([]branch.Prefix, error)
}
