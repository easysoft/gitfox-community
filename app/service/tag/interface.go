package tag

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	tagModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/tag"
)

type Interface interface {
	Create(ctx context.Context, req tagModel.ReqForCreate, session *auth.Session, repo *repoModel.Repo) (*tagModel.Tag, error)
	List(ctx context.Context, req tagModel.ReqFilter, repo repoModel.Repo) ([]tagModel.Tag, error)
	Delete(ctx context.Context, name string, session *auth.Session, repo *repoModel.Repo) error
}
