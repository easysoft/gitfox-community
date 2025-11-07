package tag

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	tagModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/tag"
)

type Interface interface {
	Create(ctx context.Context, req tagModel.ReqForCreate) (*tagModel.Tag, error)
	List(ctx context.Context, req tagModel.ReqFilter, repo repoModel.Repo) ([]tagModel.Tag, error)
	Delete(ctx context.Context, repo repoModel.Repo) ([]tagModel.Tag, error)
}
