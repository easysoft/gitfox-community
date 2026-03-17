package scan

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
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

	t.Run("get by id not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(10)).Return(nil, errors.ErrResourceNotFound)

		svc := NewSolutionService(store)
		_, err := svc.GetByID(context.Background(), 10)
		require.Error(t, err)
		assert.Equal(t, i18n.SolutionNotFound, errors.GetI18nKey(err))
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

	t.Run("list aggregates fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockISolution(ctrl)
		store.EXPECT().List(gomock.Any(), gomock.Any()).Return(&database.PageResult[*scanModel.Solution]{
			Data: []*scanModel.Solution{
				{ID: 10, Name: "sol-a"},
				{ID: 11, Name: "sol-b"},
			},
		}, nil)
		store.EXPECT().ListRulesets(gomock.Any(), uint(10)).Return([]*scanModel.RuleSet{
			{ID: 1, Plugin: "pmd", Lang: constant.ScanLangName("go"), RulesCount: 2},
			{ID: 2, Plugin: "semgrep", Lang: constant.ScanLangName("java"), RulesCount: 3},
		}, nil)
		store.EXPECT().ListRulesets(gomock.Any(), uint(11)).Return([]*scanModel.RuleSet{
			{ID: 3, Plugin: "eslint", Lang: constant.ScanLangName("js"), RulesCount: 1},
		}, nil)

		svc := NewSolutionService(store)
		got, err := svc.List(context.Background(), scanModel.SolutionReqFilter{})
		require.NoError(t, err)
		require.Len(t, got.Data, 2)
		assert.Equal(t, int64(5), got.Data[0].RulesCount)
		assert.Equal(t, int64(2), got.Data[0].RuleSetsCount)
		assert.ElementsMatch(t, []string{"pmd", "semgrep"}, got.Data[0].Plugins)
		assert.ElementsMatch(t, []string{"go", "java"}, got.Data[0].Langs)
		assert.Nil(t, got.Data[0].RuleSets)
		assert.Equal(t, int64(1), got.Data[1].RulesCount)
		assert.Equal(t, int64(1), got.Data[1].RuleSetsCount)
		assert.ElementsMatch(t, []string{"eslint"}, got.Data[1].Plugins)
		assert.ElementsMatch(t, []string{"js"}, got.Data[1].Langs)
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
