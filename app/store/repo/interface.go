package repo

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)


type RepoInterface interface {
	Create(ctx context.Context, repo *Repository) error
	FindByRef(ctx context.Context, ref string) (*repo.Repository, error)

}