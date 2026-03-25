package repo

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/repo_store.go -package=mock

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type RepoPageResult = *database.PageResult[*repoModel.Repo]

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error)
	Update(ctx context.Context, repo *Repository, id uint) (*repoModel.Repo, error)
	FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error)
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
	Delete(ctx context.Context, id uint) error
	GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error)
	List(ctx context.Context, filter repoModel.ReqFilter) (RepoPageResult, error)
}
