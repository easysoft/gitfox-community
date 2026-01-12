package pullrequest

//go:generate mockgen -source=$GOFILE -destination=mock/pullrequest_store.go -package=mock

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	GetByID(ctx context.Context, targetRepoID uint, id uint) (*repoModel.PullRequest, error)
	UpdateStatus(ctx context.Context, targetRepoID uint, id uint, status repoModel.PullRequestStatus, editedBy string) error
	UpdateMergeInfo(ctx context.Context, targetRepoID uint, id uint, mergedBy string, mergeMethod repoModel.MergeMethod, mergeSHA string, editedBy string) error
	UpdateMergeCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.MergeCheckStatus, conflicts *string) error
	UpdateRebaseCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.RebaseCheckStatus, conflicts *string) error
	UpdateDiffStats(ctx context.Context, repoID uint, id uint, diffStats *repoModel.DiffStats) error
	CountByTargetRepoID(ctx context.Context, targetRepoID uint) (int64, error)
	CountByStatus(ctx context.Context, targetRepoID uint, status repoModel.PullRequestStatus) (int64, error)
}
