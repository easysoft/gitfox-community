package repo

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error)
	FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error)
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
}
