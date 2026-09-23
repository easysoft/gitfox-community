// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"context"
	"testing"

	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	scanStoreMock "github.com/easysoft/gitfox-community/app/store/scan/mock"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRuleSetServiceCreateAndDelete(t *testing.T) {
	t.Parallel()

	t.Run("create maps fields", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, ruleSet *scanStore.RuleSet) (*scanModel.RuleSet, error) {
			assert.Equal(t, "rs-1", ruleSet.Name)
			assert.Equal(t, "tester", ruleSet.CreatedBy)
			return &scanModel.RuleSet{ID: 1, Name: ruleSet.Name}, nil
		})

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		got, err := svc.Create(context.Background(), newTestSession("tester"), scanModel.RuleSetReqForCreate{Name: "rs-1", Lang: "go", Type: "defect"})
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("delete builtin blocked", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().GetByID(gomock.Any(), uint(3)).Return(&scanModel.RuleSet{ID: 3, IsCustom: false}, nil)

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		err := svc.Delete(context.Background(), 3)
		require.Error(t, err)
	})

	t.Run("delete success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().GetByID(gomock.Any(), uint(8)).Return(&scanModel.RuleSet{ID: 8, IsCustom: true}, nil)
		ruleSetStore.EXPECT().IsUsedBySolution(gomock.Any(), uint(8)).Return(false, nil)
		ruleSetStore.EXPECT().Delete(gomock.Any(), uint(8)).Return(nil)

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		err := svc.Delete(context.Background(), 8)
		require.NoError(t, err)
	})
}

func TestRuleSetServiceBindAndQuery(t *testing.T) {
	t.Parallel()

	t.Run("GetBindRule returns empty page when no ids", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().GetRuleIDList(gomock.Any(), uint(1)).Return([]int64{}, nil)

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		got, err := svc.GetBindRule(context.Background(), 1, scanModel.RuleReqFilter{Page: 2, Size: 15})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(0), got.Total)
	})

	t.Run("GetUnBindRule delegates", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().GetRuleIDList(gomock.Any(), uint(1)).Return([]int64{1, 2}, nil)
		ruleStore.EXPECT().ExcludeByRuleSetID(gomock.Any(), []int64{1, 2}, gomock.Any()).Return(&database.PageResult[*scanModel.Rule]{Data: []*scanModel.Rule{{ID: 3}}, Total: 1}, nil)

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		got, err := svc.GetUnBindRule(context.Background(), 1, scanModel.RuleReqFilter{Page: 1, Size: 10})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(1), got.Total)
	})

	t.Run("UnBindRule success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		ruleSetStore.EXPECT().UnBindRule(gomock.Any(), uint(1), []uint{9}).Return(nil)

		svc := NewRuleSetService(ruleSetStore, ruleStore)
		err := svc.UnBindRule(context.Background(), 1, scanModel.RuleSetReqForBind{IDs: []uint{9}})
		require.NoError(t, err)
	})
}
