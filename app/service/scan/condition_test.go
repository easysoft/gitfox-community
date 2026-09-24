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

func TestScanConditionService(t *testing.T) {
	t.Parallel()

	t.Run("batch create resyncs cron job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(3), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 3}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(3), uint(1), uint(2)).Return(&scanModel.ScanTrigger{ID: 2, RepoID: 3, PlanID: 1}, nil)
		conditionStore.EXPECT().BatchCreate(gomock.Any(), gomock.Any()).Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(3), uint(1), uint(2)).Return(&scanModel.ScanTrigger{
			ID:         2,
			RepoID:     3,
			PlanID:     1,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "0 * * * *",
			CronBranch: "main",
		}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, cronScheduler)
		err := svc.BatchCreate(context.Background(), newTestSession("tester"), 1, 2, 3, scanModel.ScanConditionReqForBatchCreate{
			Conditions: []scanModel.ScanConditionItem{{Priority: constant.ScanRulePriorityLow, Type: constant.ScanRuleTypeDefect, Unit: constant.ScanRuleUnitCount, Threshold: 10}},
		})
		require.NoError(t, err)
		require.Len(t, cronScheduler.addCalls, 1)
		assert.Equal(t, scanCronJobName(2), cronScheduler.addCalls[0].jobName)
	})

	t.Run("batch delete resyncs cron job", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		cronScheduler := &fakeCronScheduler{}
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{ID: 99, RepoID: 1, PlanID: 10}, nil)
		conditionStore.EXPECT().List(gomock.Any(), uint(99)).Return([]*scanModel.ScanCondition{{ID: 7, RepoID: 1, PlanID: 10, TriggerID: 99}, {ID: 8, RepoID: 1, PlanID: 10, TriggerID: 99}}, nil)
		conditionStore.EXPECT().BatchDelete(gomock.Any(), uint(1), uint(10), uint(99), []uint{7, 8}).Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{
			ID:         99,
			RepoID:     1,
			PlanID:     10,
			Type:       scanModel.ScanTriggerTypeCron,
			Cron:       "*/10 * * * *",
			CronBranch: "main",
		}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, cronScheduler)
		err := svc.BatchDelete(context.Background(), 1, 10, 99, scanModel.ScanConditionReqForBatchDelete{IDs: []uint{7, 8}})
		require.NoError(t, err)
		require.Len(t, cronScheduler.addCalls, 1)
		assert.Equal(t, scanCronJobName(99), cronScheduler.addCalls[0].jobName)
	})

	t.Run("batch create returns trigger not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(3), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 3}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(3), uint(1), uint(2)).Return(nil, pkgErrors.ErrResourceNotFound)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, nil)
		err := svc.BatchCreate(context.Background(), newTestSession("tester"), 1, 2, 3, scanModel.ScanConditionReqForBatchCreate{})
		require.Error(t, err)
		assert.Equal(t, i18n.TriggerNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("batch create maps fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(3), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 3}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(3), uint(1), uint(2)).Return(&scanModel.ScanTrigger{ID: 2, RepoID: 3, PlanID: 1}, nil)
		conditionStore.EXPECT().BatchCreate(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, conditions []*scanStore.ScanPlanCondition) error {
			require.Len(t, conditions, 1)
			assert.Equal(t, uint64(1), conditions[0].PlanID)
			assert.Equal(t, uint64(2), conditions[0].TriggerID)
			assert.Equal(t, uint64(3), conditions[0].RepoID)
			assert.Equal(t, "tester", conditions[0].CreatedBy)
			return nil
		})
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(3), uint(1), uint(2)).Return(&scanModel.ScanTrigger{ID: 2, RepoID: 3, PlanID: 1}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, nil)
		err := svc.BatchCreate(context.Background(), newTestSession("tester"), 1, 2, 3, scanModel.ScanConditionReqForBatchCreate{
			Conditions: []scanModel.ScanConditionItem{{Priority: constant.ScanRulePriorityLow, Type: constant.ScanRuleTypeDefect, Unit: constant.ScanRuleUnitCount, Threshold: 10}},
		})
		require.NoError(t, err)
	})

	t.Run("list filters by repo and plan", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{ID: 99, RepoID: 1, PlanID: 10}, nil)
		conditionStore.EXPECT().List(gomock.Any(), uint(99)).Return([]*scanModel.ScanCondition{
			{ID: 1, RepoID: 1, PlanID: 10, TriggerID: 99},
			{ID: 2, RepoID: 2, PlanID: 10, TriggerID: 99},
			{ID: 3, RepoID: 1, PlanID: 11, TriggerID: 99},
		}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, nil)
		got, err := svc.List(context.Background(), 1, 10, 99)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1), got[0].ID)
	})

	t.Run("batch delete returns condition not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{ID: 99, RepoID: 1, PlanID: 10}, nil)
		conditionStore.EXPECT().List(gomock.Any(), uint(99)).Return([]*scanModel.ScanCondition{{ID: 7, RepoID: 1, PlanID: 10, TriggerID: 99}}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, nil)
		err := svc.BatchDelete(context.Background(), 1, 10, 99, scanModel.ScanConditionReqForBatchDelete{IDs: []uint{7, 8}})
		require.Error(t, err)
		assert.Equal(t, i18n.ConditionNotFound, pkgErrors.GetI18nKey(err))
	})

	t.Run("batch delete passes scope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		conditionStore := scanStoreMock.NewMockIScanCondition(ctrl)
		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(10)).Return(&scanModel.ScanPlan{ID: 10, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{ID: 99, RepoID: 1, PlanID: 10}, nil)
		conditionStore.EXPECT().List(gomock.Any(), uint(99)).Return([]*scanModel.ScanCondition{{ID: 7, RepoID: 1, PlanID: 10, TriggerID: 99}, {ID: 8, RepoID: 1, PlanID: 10, TriggerID: 99}}, nil)
		conditionStore.EXPECT().BatchDelete(gomock.Any(), uint(1), uint(10), uint(99), []uint{7, 8}).Return(nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(10), uint(99)).Return(&scanModel.ScanTrigger{ID: 99, RepoID: 1, PlanID: 10}, nil)

		svc := NewScanConditionService(conditionStore, newRepoStoreMock(ctrl), planStore, triggerStore, nil)
		err := svc.BatchDelete(context.Background(), 1, 10, 99, scanModel.ScanConditionReqForBatchDelete{IDs: []uint{7, 8}})
		require.NoError(t, err)
	})
}
