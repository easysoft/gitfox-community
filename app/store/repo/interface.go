package repo

//go:generate mockgen -source=$GOFILE -destination=mock/repo_store.go -package=mock

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

type RepoPageResult = *database.PageResult[*repoModel.Repo]
type PullRequestPageResult = *database.PageResult[*repoModel.PullRequest]

type Interface interface {
	Create(ctx context.Context, repo *Repository) (*repoModel.Repo, error)
	Update(ctx context.Context, repo *Repository, id uint) (*repoModel.Repo, error)
	FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error)
	GetByID(ctx context.Context, id uint) (*repoModel.Repo, error)
	Delete(ctx context.Context, id uint) error
	GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error)
	List(ctx context.Context, filter repoModel.ReqFilter) (RepoPageResult, error)

	GetPullRequestByID(ctx context.Context, repoID uint, id uint) (*repoModel.PullRequest, error)
	UpdatePullRequestStatus(ctx context.Context, repoID uint, id uint, status repoModel.PullRequestStatus, editedBy string) error
	UpdatePullRequestMergeInfo(ctx context.Context, repoID uint, id uint, mergedBy string, mergeMethod repoModel.MergeMethod, mergeSHA string, editedBy string) error
	UpdatePullRequestMergeCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.MergeCheckStatus, conflicts *string) error
	UpdatePullRequestRebaseCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.RebaseCheckStatus, conflicts *string) error
	UpdatePullRequestDiffStats(ctx context.Context, repoID uint, id uint, diffStats *repoModel.DiffStats) error
	PullRequestCountByRepoID(ctx context.Context, repoID uint) (int64, error)
	PullRequestCountByStatus(ctx context.Context, repoID uint, status repoModel.PullRequestStatus) (int64, error)
}
