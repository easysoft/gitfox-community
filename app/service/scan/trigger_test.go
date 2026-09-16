package scan

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	scanStoreMock "github.com/easysoft/gitfox-community/app/store/scan/mock"
	"github.com/easysoft/gitfox-community/pkg/database"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestScanTriggerServiceCreateAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("create cron trigger syncs recurring job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTrigger{
			ID:         100,
			PlanID:     1,
			RepoID:     2,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		}, nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, cronScheduler)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, cronScheduler.addCalls, 1)
		assert.Equal(t, scanCronJobName(100), cronScheduler.addCalls[0].jobName)
		assert.Equal(t, JobTypeScanCron, cronScheduler.addCalls[0].jobType)
		assert.Equal(t, "0 * * * *", cronScheduler.addCalls[0].cronDef)
		assert.Equal(t, 5*time.Minute, cronScheduler.addCalls[0].maxDur)
	})

	t.Run("update cron trigger syncs recurring job using latest state", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{ID: 9, RepoID: 2, PlanID: 1}, nil)
		triggerStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{
			ID:         9,
			RepoID:     2,
			PlanID:     1,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "*/5 * * * *",
			CronBranch: "release/devops062",
		}, nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, cronScheduler)
		got, err := svc.Update(context.Background(), newTestSession("tester"), 2, 1, 9, scanModel.ScanTriggerReqForUpdate{
			Cron:       "*/5 * * * *",
			CronBranch: "release/devops062",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, cronScheduler.addCalls, 1)
		assert.Equal(t, scanCronJobName(9), cronScheduler.addCalls[0].jobName)
		assert.Equal(t, "*/5 * * * *", cronScheduler.addCalls[0].cronDef)
	})

	t.Run("update disabled cron trigger removes recurring job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		cronScheduler := &fakeCronScheduler{}
		disabled := true
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{ID: 9, RepoID: 2, PlanID: 1}, nil)
		triggerStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{
			ID:         9,
			RepoID:     2,
			PlanID:     1,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
			Disabled:   true,
		}, nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, cronScheduler)
		_, err := svc.Update(context.Background(), newTestSession("tester"), 2, 1, 9, scanModel.ScanTriggerReqForUpdate{Disabled: &disabled})
		require.NoError(t, err)
		require.Len(t, cronScheduler.removeCalls, 1)
		assert.Equal(t, scanCronJobName(9), cronScheduler.removeCalls[0])
	})

	t.Run("delete trigger removes recurring job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerSolutionStore := scanStoreMock.NewMockIScanTriggerSolution(ctrl)
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(&scanModel.ScanTrigger{ID: 1, RepoID: 1, PlanID: 10}, nil)
		triggerStore.EXPECT().Count(gomock.Any(), uint(1), uint(10)).Return(int64(2), nil)
		triggerStore.EXPECT().Delete(gomock.Any(), uint(1)).Return(nil)
		triggerSolutionStore.EXPECT().DeleteByTriggerID(gomock.Any(), uint64(1)).Return(nil)

		svc := NewScanTriggerService(triggerStore, triggerSolutionStore, &fakeRepoStore{}, planStore, nil, cronScheduler)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.NoError(t, err)
		require.Len(t, cronScheduler.removeCalls, 1)
		assert.Equal(t, scanCronJobName(1), cronScheduler.removeCalls[0])
	})

	t.Run("sync helper returns scheduler errors", func(t *testing.T) {
		cronScheduler := &fakeCronScheduler{addErr: stderrors.New("boom")}
		err := syncScanTriggerCronJob(context.Background(), cronScheduler, &scanModel.ScanTrigger{
			ID:         5,
			RepoID:     2,
			PlanID:     1,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to add scan cron job")
	})

	t.Run("invalid cron expression", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 1}, nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "bad-cron",
			CronBranch: "main",
		})
		require.Error(t, err)
	})

	t.Run("create returns trigger not found repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		repo := &fakeRepoStore{err: pkgErrors.ErrResourceNotFound}

		svc := NewScanTriggerService(triggerStore, nil, repo, planStore, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.Error(t, err)
		assert.Equal(t, i18n.RepoNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("create returns plan not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(nil, pkgErrors.ErrResourceNotFound)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("create success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, trigger *scanStore.ScanPlanTrigger) (*scanModel.ScanTrigger, error) {
			assert.Equal(t, uint64(1), trigger.PlanID)
			assert.Equal(t, uint64(2), trigger.RepoID)
			assert.Equal(t, "tester", trigger.CreatedBy)
			return &scanModel.ScanTrigger{ID: 100, PlanID: 1, RepoID: 2}, nil
		})

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("delete returns trigger not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(nil, pkgErrors.ErrResourceNotFound)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.Error(t, err)
		assert.Equal(t, i18n.TriggerNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("retain at least one trigger", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(&scanModel.ScanTrigger{ID: 1, RepoID: 1, PlanID: 10}, nil)
		triggerStore.EXPECT().Count(gomock.Any(), uint(1), uint(10)).Return(int64(1), nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.Error(t, err)
	})

	t.Run("list with pagination", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		filter := scanModel.ScanTriggerReqFilter{Page: 1, PageSize: 30}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().List(gomock.Any(), uint(1), uint(10), filter).Return(&database.PageResult[*scanModel.ScanTrigger]{
			Data:  []*scanModel.ScanTrigger{{ID: 1, RepoID: 1}},
			Total: 1,
		}, nil)

		svc := NewScanTriggerService(triggerStore, nil, &fakeRepoStore{}, planStore, nil, nil)
		got, err := svc.List(context.Background(), 1, 10, filter)
		require.NoError(t, err)
		require.Len(t, got.Data, 1)
		assert.Equal(t, int64(1), got.Data[0].RepoID)
		assert.Equal(t, int64(1), got.Total)
	})
}

func TestScanTriggerServiceSolutionSync(t *testing.T) {
	t.Parallel()

	t.Run("create with solutionIDs writes to trigger_solutions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerSolutionStore := scanStoreMock.NewMockIScanTriggerSolution(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(10)).Return(&scanModel.Solution{ID: 10}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(20)).Return(&scanModel.Solution{ID: 20}, nil)
		triggerStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTrigger{
			ID: 100, PlanID: 1, RepoID: 2, Type: scanModel.ScanTriggerTypeEvent,
		}, nil)
		triggerSolutionStore.EXPECT().ReplaceBatch(gomock.Any(), uint64(100), []int64{10, 20}, "tester").Return(nil)

		svc := NewScanTriggerService(triggerStore, triggerSolutionStore, &fakeRepoStore{}, planStore, solutionStore, nil)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:        "t1",
			Type:        scanModel.ScanTriggerTypeEvent,
			Actions:     []constant.TriggerAction{constant.TriggerActionBranchUpdated},
			SolutionIDs: []int64{10, 20},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []int64{10, 20}, got.SolutionIDs)
		assert.Equal(t, 2, got.SolutionCount)
	})

	t.Run("create without solutionIDs skips trigger_solutions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerSolutionStore := scanStoreMock.NewMockIScanTriggerSolution(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&scanModel.ScanTrigger{
			ID: 101, PlanID: 1, RepoID: 2, Type: scanModel.ScanTriggerTypeEvent,
		}, nil)
		// ReplaceBatch should NOT be called

		svc := NewScanTriggerService(triggerStore, triggerSolutionStore, &fakeRepoStore{}, planStore, nil, nil)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:    "t2",
			Type:    scanModel.ScanTriggerTypeEvent,
			Actions: []constant.TriggerAction{constant.TriggerActionBranchUpdated},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Empty(t, got.SolutionIDs)
	})

	t.Run("update with solutionIDs replaces trigger_solutions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerSolutionStore := scanStoreMock.NewMockIScanTriggerSolution(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{ID: 9, RepoID: 2, PlanID: 1, Type: scanModel.ScanTriggerTypeEvent}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(30)).Return(&scanModel.Solution{ID: 30}, nil)
		triggerStore.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
		triggerSolutionStore.EXPECT().ReplaceBatch(gomock.Any(), uint64(9), []int64{30}, "tester").Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(2), uint(1), uint(9)).Return(&scanModel.ScanTrigger{
			ID: 9, RepoID: 2, PlanID: 1, Type: scanModel.ScanTriggerTypeEvent,
		}, nil)

		svc := NewScanTriggerService(triggerStore, triggerSolutionStore, &fakeRepoStore{}, planStore, solutionStore, nil)
		got, err := svc.Update(context.Background(), newTestSession("tester"), 2, 1, 9, scanModel.ScanTriggerReqForUpdate{
			Actions:     []constant.TriggerAction{constant.TriggerActionBranchUpdated},
			SolutionIDs: []int64{30},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("delete cleans up trigger_solutions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerSolutionStore := scanStoreMock.NewMockIScanTriggerSolution(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(5)).Return(&scanModel.ScanTrigger{ID: 5, RepoID: 1, PlanID: 10}, nil)
		triggerStore.EXPECT().Count(gomock.Any(), uint(1), uint(10)).Return(int64(2), nil)
		triggerStore.EXPECT().Delete(gomock.Any(), uint(5)).Return(nil)
		triggerSolutionStore.EXPECT().DeleteByTriggerID(gomock.Any(), uint64(5)).Return(nil)

		svc := NewScanTriggerService(triggerStore, triggerSolutionStore, &fakeRepoStore{}, planStore, nil, nil)
		err := svc.Delete(context.Background(), 1, 10, 5)
		require.NoError(t, err)
	})
}
