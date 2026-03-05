package scan

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSolutionServiceCreateAndGet(t *testing.T) {
	t.Parallel()

	t.Run("create duplicate blocked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().Find(gomock.Any(), "sol-a").Return(&scanModel.Solution{ID: 1, Name: "sol-a"}, nil)

		svc := NewSolutionService(store)
		_, err := svc.Create(context.Background(), newTestSession("tester"), scanModel.SolutionReqForCreate{Name: "sol-a"})
		require.Error(t, err)
	})

	t.Run("get by id aggregates fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(10)).Return(&scanModel.Solution{ID: 10, Name: "sol-a"}, nil)
		store.EXPECT().ListRulesets(gomock.Any(), uint(10)).Return([]*scanModel.RuleSet{
			{ID: 1, Plugin: "pmd", Lang: constant.ScanLangName("go"), RulesCount: 2},
			{ID: 2, Plugin: "semgrep", Lang: constant.ScanLangName("go"), RulesCount: 3},
			{ID: 3, Plugin: "pmd", Lang: constant.ScanLangName("java"), RulesCount: 1},
		}, nil)

		svc := NewSolutionService(store)
		got, err := svc.GetByID(context.Background(), 10)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(6), got.RulesCount)
		assert.Equal(t, int64(3), got.RuleSetsCount)
		assert.ElementsMatch(t, []string{"pmd", "semgrep"}, got.Plugins)
		assert.ElementsMatch(t, []string{"go", "java"}, got.Langs)
	})
}

func TestSolutionServiceStatusAndBind(t *testing.T) {
	t.Parallel()

	t.Run("update status toggles", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.Solution{ID: 1, Status: constant.ScanRuleStatusEnabled}, nil)
		store.EXPECT().Update(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, solution *scanStore.SolutionRules) error {
			assert.Equal(t, uint64(1), solution.ID)
			assert.Equal(t, constant.ScanRuleStatusDisabled, solution.Status)
			assert.Equal(t, "tester", solution.EditedBy)
			return nil
		})

		svc := NewSolutionService(store)
		err := svc.UpdateStatus(context.Background(), newTestSession("tester"), 1)
		require.NoError(t, err)
	})

	t.Run("bind rulesets requires ids", func(t *testing.T) {
		svc := NewSolutionService(nil)
		err := svc.BindRulesets(context.Background(), 1, scanModel.SolutionReqForBind{})
		require.Error(t, err)
	})

	t.Run("bind rulesets success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(2)).Return(&scanModel.Solution{ID: 2, IsCustom: true}, nil)
		store.EXPECT().BindRulesets(gomock.Any(), uint(2), []uint{7, 8}).Return(nil)

		svc := NewSolutionService(store)
		err := svc.BindRulesets(context.Background(), 2, scanModel.SolutionReqForBind{IDs: []uint{7, 8}})
		require.NoError(t, err)
	})

	t.Run("delete builtin blocked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.Solution{ID: 1, IsCustom: false}, nil)

		svc := NewSolutionService(store)
		err := svc.Delete(context.Background(), 1)
		require.Error(t, err)
	})
}
