// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package pullrequest

//go:generate go run go.uber.org/mock/mockgen@v0.6.0 -source=$GOFILE -destination=mock/pullrequest_store.go -package=mock

import (
	"context"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
)

type Interface interface {
	Create(ctx context.Context, pr *repoModel.PullRequest) error
	GetByID(ctx context.Context, targetRepoID uint, id uint) (*repoModel.PullRequest, error)
	List(ctx context.Context, filter *repoModel.PullRequestFilter) ([]*repoModel.PullRequest, error)
	UpdateOptLock(ctx context.Context, pr *repoModel.PullRequest, mutateFn func(pr *repoModel.PullRequest) error) (*repoModel.PullRequest, error)
	UpdateStatus(ctx context.Context, targetRepoID uint, id uint, status repoModel.PullRequestStatus, editedBy string) error
	UpdateMergeInfo(ctx context.Context, targetRepoID uint, id uint, mergedBy string, mergeMethod repoModel.MergeMethod, mergeSHA string, editedBy string) error
	UpdateMergeCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.MergeCheckStatus, conflicts *string) error
	UpdateRebaseCheckStatus(ctx context.Context, repoID uint, id uint, status repoModel.MergeCheckStatus, conflicts *string) error
	ResetMergeCheckStatus(ctx context.Context, targetRepoID uint, targetBranch string) error
	UpdateDiffStatsSafe(ctx context.Context, repoID uint, pr *repoModel.PullRequest, diffStats *repoModel.DiffStats) (*repoModel.PullRequest, error)
	UpdateMergeAfterSafe(ctx context.Context, repoID uint, pr *repoModel.PullRequest, mergeAfter *repoModel.MergeAfter) error
	CountByTargetRepoID(ctx context.Context, targetRepoID uint) (int64, error)
	CountByStatus(ctx context.Context, targetRepoID uint, status repoModel.PullRequestStatus) (int64, error)
	// GetReviewersByRequestID 根据请求ID（ppmID）查询 ppm 表（ops_ppm）的 reviewers 字段
	GetReviewersByRequestID(ctx context.Context, requestID uint) ([]string, error)
	// ListReviewerAccountsByRequestID 根据请求ID（ppmID）查询评审人账号列表
	ListReviewerAccountsByRequestID(ctx context.Context, requestID uint) ([]string, error)
}
