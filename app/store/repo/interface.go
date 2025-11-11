package repo

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.CreateRepoResp, error)
	FindByRef(ctx context.Context, ref string) (*Repository, error)
}
