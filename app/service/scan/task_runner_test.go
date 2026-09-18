package scan

import (
	"context"
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
		newRepoStoreWithRepoMock(ctrl, repo),
		nil, nil, nil, nil, nil, nil, // pipeline / pullRequest / execution / pipelineService / triggerSvc / cronScheduler
		lister,
		nil, // tx
	)
	return svc, planStore, taskStore
}

// TestCreateRunnerTask 覆盖非内嵌流水线创建扫描任务：只按 body 落任务记录并回
// taskID，提交范围不再解析（由插件的 SCAN_METHOD + SCAN_REF_RANGE 声明）。
func TestCreateRunnerTask(t *testing.T) {
	t.Parallel()

	repo := &repoModel.Repo{ID: 1, GitUID: "uid-1"}
	req := &scanModel.ScanTaskRunnerCreateReq{
		RepoID:      1,
		PlanID:      456,
		ExecutionID: 789,
		TriggerType: string(constant.TriggerActionBranchUpdated),
	}

	t.Run("回 taskID 且不查提交范围", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{}
		svc, _, taskStore := newRunnerTaskFixture(t, string(constant.ScanTypeIncremental), repo, lister)
		taskStore.EXPECT().NextPlanNumber(gomock.Any(), uint64(456)).Return(int64(1), nil)
		taskStore.EXPECT().NextRepoNumber(gomock.Any(), int64(1)).Return(int64(1), nil)
		taskStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTask{ID: 9001}, nil)

		res, err := svc.CreateRunnerTask(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int64(9001), res.TaskID)
		// 范围不再由服务端解析：不应触发任何 git 查询（git 不可用时也能建任务）
		assert.Equal(t, 0, lister.calls)
	})

	t.Run("计划不属于该仓库时不落任务记录", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{}
		ctrl := gomock.NewController(t)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		taskStore := scanStoreMock.NewMockIScanTask(ctrl)
		planStore.EXPECT().GetByID(gomock.Any(), gomock.Any()).
			Return(&scanModel.ScanPlan{ID: 456, RepoID: 2}, nil)
		svc := NewScanTaskService(
			nil, taskStore, planStore,
			nil, nil, nil, nil, nil,
			newRepoStoreWithRepoMock(ctrl, repo),
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
