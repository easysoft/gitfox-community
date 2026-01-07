package pullrequest

import (
	"context"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
)

type Interface interface {
	GetByID(ctx context.Context, repoID uint, id uint) (*repoModel.PullRequest, error)
	UpdateStatus(ctx context.Context, repoID uint, id uint, status repoModel.PullRequestStatus, editedBy string) error
	UpdateMergeInfo(ctx context.Context, repoID uint, id uint, mergedBy string, mergeMethod repoModel.MergeMethod, mergeSHA string, editedBy string) error
	UpdateMergeCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.MergeCheckStatus, conflicts *string) error
	UpdateRebaseCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.RebaseCheckStatus, conflicts *string) error
	UpdateDiffStats(ctx context.Context, repoID uint, id uint, diffStats *repoModel.DiffStats) error
	CountByRepoID(ctx context.Context, repoID uint) (int64, error)
	CountByStatus(ctx context.Context, repoID uint, status repoModel.PullRequestStatus) (int64, error)
}
