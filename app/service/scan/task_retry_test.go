// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scan

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	pipelineMock "github.com/easysoft/gitfox-community/app/store/pipeline/mock"
	scanStoreMock "github.com/easysoft/gitfox-community/app/store/scan/mock"
	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// fakeTriggerer 记录触发时的 hook，用于断言重试传给流水线的提交范围。
type fakeTriggerer struct {
	calls int
	hook  *triggerer.Hook
}

func (f *fakeTriggerer) Trigger(ctx context.Context, pipelineID uint, hook *triggerer.Hook) (*executionModel.Execution, error) {
	f.calls++
	f.hook = hook
	return &executionModel.Execution{ID: 77}, nil
}

// fakeEmbeddedPipelineSvc 只实现重试链路会用到的方法；其余方法来自嵌入的接口，
// 测试中不会被调用（调用会因空接口 panic，反而能暴露多余的依赖）。
type fakeEmbeddedPipelineSvc struct {
	pipelineSvc.Interface
	updateCalls int
}

func (f *fakeEmbeddedPipelineSvc) UpdateEmbeddedPipelineContent(ctx context.Context, yamlLabel string, pipelineID int64) error {
	f.updateCalls++
	return nil
}

// TestRetryCommitRange 覆盖内置重试的扫描范围解析：复用原执行的分支，按当时扫描
// 计划的 scanType 取最新两笔（增量）或同一笔（全量）；ref 缺失或查不到时直接失败。
func TestRetryCommitRange(t *testing.T) {
	t.Parallel()

	const (
		latest = "sha-latest"
		second = "sha-second"
		branch = "main"
	)
	repo := &repoModel.Repo{ID: 1, GitUID: "uid-1", SpaceID: 9}
	session := newTestSession("tester")

	newRetryFixture := func(t *testing.T, scanType, executionRef string, lister *fakeCommitLister) (*ScanTaskService, *scanStoreMock.MockIScanTask, *fakeTriggerer) {
		t.Helper()
		ctrl := gomock.NewController(t)
		taskStore := scanStoreMock.NewMockIScanTask(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		pipelineStore := pipelineMock.NewMockInterface(ctrl)
		triggerSvc := &fakeTriggerer{}

		// 内置流水线已存在 → 走 updatePipelineContent（不触发创建流程）。
		// AnyTimes：范围解析失败的分支在查流水线之前就返回了
		pipelineStore.EXPECT().GetPipelineByRepoIDAndName(gomock.Any(), int64(1), "_codescan").
			Return(&pipelineModel.Pipeline{ID: 88}, nil).AnyTimes()

		taskStore.EXPECT().GetTaskInfo(gomock.Any(), uint(1)).Return(&scanModel.ScanTaskInfo{
			PlanID:       456,
			RepoID:       1,
			Status:       constant.ScanTaskStatusFailed,
			ExecutionRef: executionRef,
		}, nil)
		planStore.EXPECT().GetByID(gomock.Any(), uint(456)).
			Return(&scanModel.ScanPlan{ID: 456, RepoID: 1, ScanType: scanType}, nil)
		planStore.EXPECT().GetSolutionIDList(gomock.Any(), uint(456)).Return(nil, nil).AnyTimes()

		svc := NewScanTaskService(
			&config.Config{
				URL:          config.URL{Base: "http://gitfox"},
				PipelineAuth: config.PipelineAuthConfig{Password: "tok"},
			},
			taskStore, planStore,
			nil, nil, nil, nil, nil,
			newRepoStoreWithRepoMock(ctrl, repo),
			pipelineStore, nil, nil,
			&fakeEmbeddedPipelineSvc{}, // 内置流水线内容更新
			triggerSvc,
			nil,
			lister,
			nil,
		)
		return svc, taskStore, triggerSvc
	}

	t.Run("增量计划复用原执行分支并取最新两笔", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest, second}}
		svc, taskStore, triggerSvc := newRetryFixture(t, string(constant.ScanTypeIncremental), branch, lister)
		taskStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		taskStore.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.ScanTask{ID: 1}, nil)

		_, err := svc.Retry(context.Background(), session, 1, repo)
		require.NoError(t, err)
		require.Equal(t, 1, triggerSvc.calls)
		assert.Equal(t, branch, lister.lastGitREF)
		assert.Equal(t, "uid-1", lister.lastRepoUID)
		// 增量：最新为 after、次新为 before、Ref 为原分支
		assert.Equal(t, latest, triggerSvc.hook.Commit)
		assert.Equal(t, second, triggerSvc.hook.CommitBefore)
		assert.Equal(t, branch, triggerSvc.hook.Ref)
		assert.Equal(t, "true", triggerSvc.hook.Params["INNER_PIPELINE"])
		assert.Equal(t, "1", triggerSvc.hook.Params["SCAN_TASK_ID"])
	})

	t.Run("全量计划取同一笔（before = after）", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest, second}}
		svc, taskStore, triggerSvc := newRetryFixture(t, string(constant.ScanTypeFull), branch, lister)
		taskStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		taskStore.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.ScanTask{ID: 1}, nil)

		_, err := svc.Retry(context.Background(), session, 1, repo)
		require.NoError(t, err)
		assert.Equal(t, latest, triggerSvc.hook.Commit)
		assert.Equal(t, latest, triggerSvc.hook.CommitBefore)
	})

	t.Run("增量计划分支只有一笔提交时两值相同", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		svc, taskStore, triggerSvc := newRetryFixture(t, string(constant.ScanTypeIncremental), branch, lister)
		taskStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		taskStore.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.ScanTask{ID: 1}, nil)

		_, err := svc.Retry(context.Background(), session, 1, repo)
		require.NoError(t, err)
		assert.Equal(t, latest, triggerSvc.hook.Commit)
		assert.Equal(t, latest, triggerSvc.hook.CommitBefore)
	})

	t.Run("executionRef 为空时报错且不触发流水线", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{shas: []string{latest}}
		svc, _, triggerSvc := newRetryFixture(t, string(constant.ScanTypeIncremental), "", lister)

		_, err := svc.Retry(context.Background(), session, 1, repo)
		require.Error(t, err)
		assert.ErrorContains(t, err, "扫描分支为空")
		assert.Equal(t, 0, lister.calls)
		assert.Equal(t, 0, triggerSvc.calls)
	})

	t.Run("分支查不到提交时报错且不触发流水线", func(t *testing.T) {
		t.Parallel()

		lister := &fakeCommitLister{}
		svc, _, triggerSvc := newRetryFixture(t, string(constant.ScanTypeIncremental), branch, lister)

		_, err := svc.Retry(context.Background(), session, 1, repo)
		require.Error(t, err)
		assert.ErrorContains(t, err, "no commits found on branch "+branch)
		assert.Equal(t, 0, triggerSvc.calls)
	})
}
