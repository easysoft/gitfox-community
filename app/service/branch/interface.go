package branch

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/branch"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, req branch.ReqForCreate) (*branch.RespForCreate, error)
	List(ctx context.Context, req branch.ReqFilter, repo repo.Repo) ([]branch.Branch, error)
	ListLables(ctx context.Context, repo repo.Repo) ([]branch.Label, error)
}
