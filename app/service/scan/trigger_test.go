package scan

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, cronScheduler)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, cronScheduler)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, cronScheduler)
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
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil).Times(2)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(&scanModel.ScanTrigger{ID: 1, RepoID: 1, PlanID: 10}, nil)
		triggerStore.EXPECT().List(gomock.Any(), uint(10)).Return([]*scanModel.ScanTrigger{{ID: 1, RepoID: 1, PlanID: 10}, {ID: 2, RepoID: 1, PlanID: 10}}, nil)
		triggerStore.EXPECT().Delete(gomock.Any(), uint(1)).Return(nil)

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, cronScheduler)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
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

		svc := NewScanTriggerService(triggerStore, repo, planStore, nil)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
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

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.Error(t, err)
		assert.Equal(t, i18n.TriggerNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("retain at least one trigger", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil).Times(2)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(&scanModel.ScanTrigger{ID: 1, RepoID: 1, PlanID: 10}, nil)
		triggerStore.EXPECT().List(gomock.Any(), uint(10)).Return([]*scanModel.ScanTrigger{{ID: 1, RepoID: 1, PlanID: 10}}, nil)

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.Error(t, err)
	})

	t.Run("list filters by repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().List(gomock.Any(), uint(10)).Return([]*scanModel.ScanTrigger{{ID: 1, RepoID: 1}, {ID: 2, RepoID: 2}}, nil)

		svc := NewScanTriggerService(triggerStore, &fakeRepoStore{}, planStore, nil)
		got, err := svc.List(context.Background(), 1, 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1), got[0].RepoID)
	})
}
