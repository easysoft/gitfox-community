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

func TestScanPlanServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("solutionIDs required", func(t *testing.T) {
		svc := NewScanPlanService(nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a"})
		require.Error(t, err)
	})

	t.Run("duplicate name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanPlan(ctrl)
		store.EXPECT().FindByName(gomock.Any(), uint(1), "plan-a").Return(&scanModel.ScanPlan{ID: 9}, nil)

		svc := NewScanPlanService(store)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{1}})
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIScanPlan(ctrl)
		store.EXPECT().FindByName(gomock.Any(), uint(1), "plan-a").Return(nil, nil)
		store.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, plan *scanStore.ScanPlan) (*scanModel.ScanPlan, error) {
			assert.Equal(t, uint64(1), plan.RepoID)
			assert.Equal(t, "tester", plan.CreatedBy)
			return &scanModel.ScanPlan{ID: 11, RepoID: 1}, nil
		})
		store.EXPECT().BindSolutions(gomock.Any(), uint(11), []uint{2, 3}).Return(nil)
		store.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(11)).Return(&scanModel.ScanPlan{ID: 11, RepoID: 1, Name: "plan-a"}, nil)
		store.EXPECT().GetSolutionIDList(gomock.Any(), uint(11)).Return([]int64{2, 3}, nil)

		svc := NewScanPlanService(store)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{2, 3}})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []int64{2, 3}, got.SolutionIDs)
		assert.Equal(t, 2, got.SolutionCount)
	})
}
