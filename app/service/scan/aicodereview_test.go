package scan

import (
	"context"
	"errors"
	"fmt"
	"testing"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newReviewRangeFixture 组装只带评审范围解析所需依赖的 AIReviewService
// （aireviewStore / config / pipelineStore / pipelineService / triggerSvc / tokenStore 用不到）。
func newReviewRangeFixture(repo *repoModel.Repo, lister *fakeCommitLister) *AIReviewService {
	return NewService(nil, nil, nil, nil, nil, nil, &fakeRepoStore{repo: repo}, lister)
}

// TestGetReviewRange 覆盖非内嵌流水线的评审范围下发：取分支最新两笔
// （次新为起点、最新为终点）；只有一笔时两者相同（ocr 侧为 skipped 空报告）。
func TestGetReviewRange(t *testing.T) {
	t.Parallel()

	const (
		latest = "sha-latest"
		second = "sha-second"
		branch = "main"
	)
	repo := &repoModel.Repo{ID: 1, GitUID: "uid-1"}

	t.Run("两笔提交取次新与最新", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest, second}}
		svc := newReviewRangeFixture(repo, lister)

		res, err := svc.GetReviewRange(context.Background(), 1, branch)
		require.NoError(t, err)
		assert.Equal(t, second, res.CommitBefore)
		assert.Equal(t, latest, res.CommitAfter)
		assert.Equal(t, branch, lister.lastGitREF)
		assert.Equal(t, "uid-1", lister.lastRepoUID)
		assert.Equal(t, int32(2), lister.lastLimit)
	})

	t.Run("仅一笔提交时两值相同", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		svc := newReviewRangeFixture(repo, lister)

		res, err := svc.GetReviewRange(context.Background(), 1, branch)
		require.NoError(t, err)
		assert.Equal(t, latest, res.CommitBefore)
		assert.Equal(t, latest, res.CommitAfter)
	})

	t.Run("分支为空时报错且不查提交", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		svc := newReviewRangeFixture(repo, lister)

		_, err := svc.GetReviewRange(context.Background(), 1, "")
		require.Error(t, err)
		assert.ErrorContains(t, err, "评审分支为空")
		assert.Equal(t, 0, lister.calls)
	})

	t.Run("分支无提交时报错", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{}
		svc := newReviewRangeFixture(repo, lister)

		_, err := svc.GetReviewRange(context.Background(), 1, branch)
		require.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("no commits found on branch %s", branch))
	})

	t.Run("git 查询失败时包装错误", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{err: errors.New("boom")}
		svc := newReviewRangeFixture(repo, lister)

		_, err := svc.GetReviewRange(context.Background(), 1, branch)
		require.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("failed to list commit SHAs for branch %s", branch))
	})

	t.Run("仓库不存在时报错", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		svc := NewService(nil, nil, nil, nil, nil, nil, &fakeRepoStore{err: errors.New("not found")}, lister)

		_, err := svc.GetReviewRange(context.Background(), 1, branch)
		require.Error(t, err)
		assert.ErrorContains(t, err, "failed to get repo by ID 1")
		assert.Equal(t, 0, lister.calls)
	})
}
