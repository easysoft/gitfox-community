package scan

import (
	"context"
	"testing"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRuleServiceDelegation(t *testing.T) {
	t.Parallel()

	t.Run("GetByID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIRule(ctrl)
		store.EXPECT().GetByID(gomock.Any(), uint(7)).Return(&scanModel.Rule{ID: 7, Name: "r1"}, nil)

		svc := NewRuleService(store, nil)
		got, err := svc.GetByID(context.Background(), &authModel.Session{}, 7)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(7), got.ID)
	})

	t.Run("List", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIRule(ctrl)
		store.EXPECT().List(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, filter scanModel.RuleListReqFilter) (*database.PageResult[*scanModel.Rule], error) {
			assert.Equal(t, "demo", filter.Name)
			return &database.PageResult[*scanModel.Rule]{Data: []*scanModel.Rule{{ID: 1}}, Total: 1}, nil
		})

		svc := NewRuleService(store, nil)
		got, err := svc.List(context.Background(), &authModel.Session{}, scanModel.RuleListReqFilter{Name: "demo"})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, int64(1), got.Total)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIRule(ctrl)
		store.EXPECT().UpdateStatus(gomock.Any(), uint(9)).Return(nil)

		svc := NewRuleService(store, nil)
		err := svc.UpdateStatus(context.Background(), &authModel.Session{}, 9)
		require.NoError(t, err)
	})

	t.Run("AllScanBaseByField", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		store := scanStoreMock.NewMockIRule(ctrl)
		store.EXPECT().AllScanBaseByField(gomock.Any(), "plugin").Return([]scanModel.ScanBaseResp{{Plugin: "pmd", Count: 3}}, nil)

		svc := NewRuleService(store, nil)
		got, err := svc.AllScanBaseByField(context.Background(), &authModel.Session{}, "plugin")
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, "pmd", got[0].Plugin)
	})
}
