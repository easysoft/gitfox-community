// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	scanStoreMock "github.com/easysoft/gitfox-community/app/store/scan/mock"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestScanPlanConditionService(t *testing.T) {
	t.Parallel()

	t.Run("create defaults unit to count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *scanStore.ScanPlanCondition) (*scanModel.ScanCondition, error) {
			assert.Equal(t, uint64(2), c.RepoID)
			assert.Equal(t, uint64(1), c.PlanID)
			assert.Equal(t, uint64(0), c.TriggerID)
			assert.Equal(t, "count", string(c.Unit))
			assert.Equal(t, "tester", c.CreatedBy)
			return &scanModel.ScanCondition{ID: 10, RepoID: 2, PlanID: 1, TriggerID: 0, Unit: constant.ScanRuleUnitCount, Threshold: 1}, nil
		})

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 2, 1, scanModel.ScanPlanConditionReqForCreate{
			Priority:  constant.ScanRulePriorityLow,
			Type:      constant.ScanRuleTypeDefect,
			Unit:      "",
			Threshold: 1,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(10), got.ID)
	})

	t.Run("create validates threshold for count", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 2, 1, scanModel.ScanPlanConditionReqForCreate{
			Priority:  constant.ScanRulePriorityLow,
			Type:      constant.ScanRuleTypeDefect,
			Unit:      constant.ScanRuleUnitCount,
			Threshold: 0,
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionBodyParamError, pkgErrors.GetI18nKey(err))
	})

	t.Run("create validates threshold for percent", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 2, 1, scanModel.ScanPlanConditionReqForCreate{
			Priority:  constant.ScanRulePriorityLow,
			Type:      constant.ScanRuleTypeDefect,
			Unit:      constant.ScanRuleUnitPercent,
			Threshold: 101,
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionBodyParamError, pkgErrors.GetI18nKey(err))
	})

	t.Run("list returns plan-scoped only", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().ListByPlan(gomock.Any(), uint(2), uint(1)).Return([]*scanModel.ScanCondition{
			{ID: 1, RepoID: 2, PlanID: 1, TriggerID: 0},
		}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		got, err := svc.List(context.Background(), 2, 1)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1), got[0].ID)
	})

	t.Run("find enforces repo+plan scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().GetByID(gomock.Any(), uint(9)).Return(&scanModel.ScanCondition{ID: 9, RepoID: 2, PlanID: 999, TriggerID: 0}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		_, err := svc.Find(context.Background(), 2, 1, 9)
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("update enforces scope and applies patch", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		unit := constant.ScanRuleUnitPercent
		threshold := int64(90)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil).Times(1)
		conditionStore.EXPECT().GetByID(gomock.Any(), uint(9)).Return(&scanModel.ScanCondition{ID: 9, RepoID: 2, PlanID: 1, TriggerID: 0, Unit: constant.ScanRuleUnitCount, Threshold: 10}, nil).Times(2)
		conditionStore.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, c *scanStore.ScanPlanCondition) error {
			assert.Equal(t, uint64(9), c.ID)
			assert.Equal(t, "tester", c.EditedBy)
			assert.Equal(t, "percent", string(c.Unit))
			assert.Equal(t, int64(90), c.Threshold)
			return nil
		})

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		got, err := svc.Update(context.Background(), newTestSession("tester"), 2, 1, 9, scanModel.ScanPlanConditionReqForUpdate{
			Unit:      &unit,
			Threshold: &threshold,
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(9), got.ID)
	})

	t.Run("update validates combined unit+threshold", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		unit := constant.ScanRuleUnitPercent
		threshold := int64(200)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().GetByID(gomock.Any(), uint(9)).Return(&scanModel.ScanCondition{ID: 9, RepoID: 2, PlanID: 1, TriggerID: 0, Unit: constant.ScanRuleUnitCount, Threshold: 10}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		_, err := svc.Update(context.Background(), newTestSession("tester"), 2, 1, 9, scanModel.ScanPlanConditionReqForUpdate{
			Unit:      &unit,
			Threshold: &threshold,
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionBodyParamError, pkgErrors.GetI18nKey(err))
	})

	t.Run("delete enforces scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().GetByID(gomock.Any(), uint(9)).Return(&scanModel.ScanCondition{ID: 9, RepoID: 999, PlanID: 1, TriggerID: 0}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		err := svc.Delete(context.Background(), 2, 1, 9)
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("batch create empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		err := svc.BatchCreate(context.Background(), newTestSession("tester"), 2, 1, scanModel.ScanPlanConditionReqForBatchCreate{})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionBodyParamError, pkgErrors.GetI18nKey(err))
	})

	t.Run("batch delete validates ids in scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().ListByPlan(gomock.Any(), uint(2), uint(1)).Return([]*scanModel.ScanCondition{{ID: 7, RepoID: 2, PlanID: 1, TriggerID: 0}}, nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		err := svc.BatchDelete(context.Background(), 2, 1, scanModel.ScanPlanConditionReqForBatchDelete{IDs: []uint{7, 8}})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanConditionNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("batch delete passes scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 2}, nil)
		conditionStore.EXPECT().ListByPlan(gomock.Any(), uint(2), uint(1)).Return([]*scanModel.ScanCondition{{ID: 7, RepoID: 2, PlanID: 1, TriggerID: 0}}, nil)
		conditionStore.EXPECT().BatchDeleteByIDs(gomock.Any(), uint(2), uint(1), []uint{7}).Return(nil)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		err := svc.BatchDelete(context.Background(), 2, 1, scanModel.ScanPlanConditionReqForBatchDelete{IDs: []uint{7}})
		require.NoError(t, err)
	})

	t.Run("create maps repo/plan not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(1)).Return(nil, pkgErrors.ErrResourceNotFound)

		svc := NewScanPlanConditionService(newRepoStoreMock(ctrl), planStore, conditionStore)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 2, 1, scanModel.ScanPlanConditionReqForCreate{Threshold: 1})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanNotFound, pkgErrors.GetI18nKey(err))
	})
}
