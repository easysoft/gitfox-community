package scan

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// fakeCommitLister 记录调用参数并返回预置结果，替代真实的 pkg/git 查询。
type fakeCommitLister struct {
	shas        []string
	err         error
	calls       int
	lastGitREF  string
	lastRepoUID string
	lastPage    int32
	lastLimit   int32
}

func (f *fakeCommitLister) ListCommitSHAs(ctx context.Context, params *git.ListCommitsParams) ([]string, error) {
	f.calls++
	f.lastGitREF = params.GitREF
	f.lastRepoUID = params.RepoUID
	f.lastPage = params.Page
	f.lastLimit = params.Limit
	return f.shas, f.err
}

// newRunnerTaskFixture 组装只带 CreateRunnerTask 所需依赖的 ScanTaskService。
func newRunnerTaskFixture(t *testing.T, scanType string, repo *repoModel.Repo, lister *fakeCommitLister) (*ScanTaskService, *scanStoreMock.MockIScanPlan, *scanStoreMock.MockIScanTask) {
	t.Helper()

	ctrl := gomock.NewController(t)
	planStore := scanStoreMock.NewMockIScanPlan(ctrl)
	taskStore := scanStoreMock.NewMockIScanTask(ctrl)
	planStore.EXPECT().GetByID(gomock.Any(), gomock.Any()).
		Return(&scanModel.ScanPlan{ID: 456, RepoID: int64(repo.ID), ScanType: scanType}, nil).
		AnyTimes()

	svc := NewScanTaskService(
		nil, // config
		taskStore, planStore,
		nil, nil, nil, nil, nil, // solution / issue / condition / ruleSet / rule
		&fakeRepoStore{repo: repo},
		nil, nil, nil, nil, nil, nil, // pipeline / pullRequest / execution / pipelineService / triggerSvc / cronScheduler
		lister,
		nil, // tx
	)
	return svc, planStore, taskStore
}

// TestCreateRunnerTaskCommitRange 覆盖非内嵌流水线创建扫描任务时的提交范围解析：
// 增量返回 次新/最新，全量返回 最新/最新，git 失败或分支无提交时不落任务记录。
func TestCreateRunnerTaskCommitRange(t *testing.T) {
	t.Parallel()

	const (
		latest = "sha-latest"
		second = "sha-second"
		branch = "main"
	)
	repo := &repoModel.Repo{ID: 1, GitUID: "uid-1"}
	req := &scanModel.ScanTaskRunnerCreateReq{
		RepoID:      1,
		PlanID:      456,
		ExecutionID: 789,
		TriggerType: string(constant.TriggerActionBranchUpdated),
		Branch:      branch,
	}

	t.Run("增量计划返回最新与次新并透传查询参数", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest, second}}
		svc, _, taskStore := newRunnerTaskFixture(t, string(constant.ScanTypeIncremental), repo, lister)
		taskStore.EXPECT().NextPlanNumber(gomock.Any(), uint64(456)).Return(int64(1), nil)
		taskStore.EXPECT().NextRepoNumber(gomock.Any(), int64(1)).Return(int64(1), nil)
		taskStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTask{ID: 9001}, nil)

		res, err := svc.CreateRunnerTask(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int64(9001), res.TaskID)
		assert.Equal(t, second, res.CommitBefore)
		assert.Equal(t, latest, res.CommitAfter)
		assert.Equal(t, 1, lister.calls)
		assert.Equal(t, branch, lister.lastGitREF)
		assert.Equal(t, "uid-1", lister.lastRepoUID)
		assert.Equal(t, int32(1), lister.lastPage)
		assert.Equal(t, int32(2), lister.lastLimit)
	})

	t.Run("全量计划两值相同", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest, second}}
		svc, _, taskStore := newRunnerTaskFixture(t, string(constant.ScanTypeFull), repo, lister)
		taskStore.EXPECT().NextPlanNumber(gomock.Any(), uint64(456)).Return(int64(1), nil)
		taskStore.EXPECT().NextRepoNumber(gomock.Any(), int64(1)).Return(int64(1), nil)
		taskStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTask{ID: 9002}, nil)

		res, err := svc.CreateRunnerTask(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int64(9002), res.TaskID)
		assert.Equal(t, latest, res.CommitBefore)
		assert.Equal(t, latest, res.CommitAfter)
	})

	t.Run("git 查询失败时不落任务记录", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{err: errors.New("boom")}
		svc, _, _ := newRunnerTaskFixture(t, string(constant.ScanTypeIncremental), repo, lister)

		_, err := svc.CreateRunnerTask(context.Background(), req)
		require.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("failed to list commit SHAs for branch %s", branch))
	})

	t.Run("分支无提交时不落任务记录", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{}
		svc, _, _ := newRunnerTaskFixture(t, string(constant.ScanTypeIncremental), repo, lister)

		_, err := svc.CreateRunnerTask(context.Background(), req)
		require.Error(t, err)
		assert.ErrorContains(t, err, fmt.Sprintf("no commits found on branch %s", branch))
	})

	t.Run("计划不属于该仓库时不查提交", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		ctrl := gomock.NewController(t)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		taskStore := scanStoreMock.NewMockIScanTask(ctrl)
		planStore.EXPECT().GetByID(gomock.Any(), gomock.Any()).
			Return(&scanModel.ScanPlan{ID: 456, RepoID: 2}, nil)
		svc := NewScanTaskService(
			nil, taskStore, planStore,
			nil, nil, nil, nil, nil,
			&fakeRepoStore{repo: repo},
			nil, nil, nil, nil, nil, nil,
			lister,
			nil,
		)

		_, err := svc.CreateRunnerTask(context.Background(), req)
		require.Error(t, err)
		assert.ErrorContains(t, err, "does not belong to repo")
		assert.Equal(t, 0, lister.calls)
	})
}
