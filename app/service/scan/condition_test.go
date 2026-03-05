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

func TestScanConditionService(t *testing.T) {
	t.Parallel()

	t.Run("batch create maps fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanCondition(ctrl)
		store.EXPECT().BatchCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, conditions []*scanStore.ScanPlanCondition) error {
			require.Len(t, conditions, 1)
			assert.Equal(t, uint64(1), conditions[0].PlanID)
			assert.Equal(t, uint64(2), conditions[0].TriggerID)
			assert.Equal(t, uint64(3), conditions[0].RepoID)
			assert.Equal(t, "tester", conditions[0].CreatedBy)
			return nil
		})

		svc := NewScanConditionService(store)
		err := svc.BatchCreate(context.Background(), newTestSession("tester"), 1, 2, 3, scanModel.ScanConditionReqForBatchCreate{
			Conditions: []scanModel.ScanConditionItem{{Priority: scanModel.ScanRulePriorityLow, Type: scanModel.ScanRuleTypeDefect, Unit: scanModel.ScanRuleUnitCount, Threshold: 10}},
		})
		require.NoError(t, err)
	})

	t.Run("list filters by repo and plan", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanCondition(ctrl)
		store.EXPECT().List(gomock.Any(), uint(99)).Return([]*scanModel.ScanCondition{
			{ID: 1, RepoID: 1, PlanID: 10, TriggerID: 99},
			{ID: 2, RepoID: 2, PlanID: 10, TriggerID: 99},
			{ID: 3, RepoID: 1, PlanID: 11, TriggerID: 99},
		}, nil)

		svc := NewScanConditionService(store)
		got, err := svc.List(context.Background(), 1, 10, 99)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1), got[0].ID)
	})

	t.Run("batch delete passes scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanCondition(ctrl)
		store.EXPECT().BatchDelete(gomock.Any(), uint(1), uint(10), uint(99), []uint{7, 8}).Return(nil)

		svc := NewScanConditionService(store)
		err := svc.BatchDelete(context.Background(), 1, 10, 99, scanModel.ScanConditionReqForBatchDelete{IDs: []uint{7, 8}})
		require.NoError(t, err)
	})
}
