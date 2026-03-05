package scan

import (
	"context"
	"testing"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestScanTriggerServiceCreateAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("invalid cron expression", func(t *testing.T) {
		svc := NewScanTriggerService(nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "bad-cron",
			CronBranch: "main",
		})
		require.Error(t, err)
	})

	t.Run("create success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanTrigger(ctrl)
		store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, trigger *scanStore.ScanPlanTrigger) (*scanModel.ScanTrigger, error) {
			assert.Equal(t, uint64(1), trigger.PlanID)
			assert.Equal(t, uint64(2), trigger.RepoID)
			assert.Equal(t, "tester", trigger.CreatedBy)
			return &scanModel.ScanTrigger{ID: 100, PlanID: 1, RepoID: 2}, nil
		})

		svc := NewScanTriggerService(store)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, 2, scanModel.ScanTriggerReqForCreate{
			Name:       "t1",
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		})
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("retain at least one trigger", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanTrigger(ctrl)
		store.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(1)).Return(&scanModel.ScanTrigger{ID: 1, RepoID: 1, PlanID: 10}, nil)
		store.EXPECT().List(gomock.Any(), uint(10)).Return([]*scanModel.ScanTrigger{{ID: 1, RepoID: 1, PlanID: 10}}, nil)

		svc := NewScanTriggerService(store)
		err := svc.Delete(context.Background(), 1, 10, 1)
		require.Error(t, err)
	})

	t.Run("list filters by repo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanTrigger(ctrl)
		store.EXPECT().List(gomock.Any(), uint(10)).Return([]*scanModel.ScanTrigger{{ID: 1, RepoID: 1}, {ID: 2, RepoID: 2}}, nil)

		svc := NewScanTriggerService(store)
		got, err := svc.List(context.Background(), 1, 10)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1), got[0].RepoID)
	})
}
