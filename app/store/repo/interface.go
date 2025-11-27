package repo

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error)
	Update(ctx context.Context, repo *Repository, id uint) (*repoModel.Repo, error)
	FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error)
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
	Delete(ctx context.Context, id uint) error
	FindBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error)
}
