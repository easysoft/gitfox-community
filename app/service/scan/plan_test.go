package scan

import (
	"context"
	"testing"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	scanStoreMock "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan/mock"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type fakeRepoStore struct {
	repo *repoModel.Repo
	err  error
}

func (f *fakeRepoStore) Create(ctx context.Context, repo *repoStore.Repository) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) Update(ctx context.Context, repo *repoStore.Repository, id uint) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) FindByRef(ctx context.Context, ref string) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) GetByID(ctx context.Context, id uint) (*repoModel.Repo, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.repo != nil {
		return f.repo, nil
	}
	return &repoModel.Repo{ID: id}, nil
}

func (f *fakeRepoStore) Delete(ctx context.Context, id uint) error {
	return nil
}

func (f *fakeRepoStore) GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) List(ctx context.Context, filter repoModel.ReqFilter) (*database.PageResult[*repoModel.Repo], error) {
	return nil, nil
}

func TestScanPlanServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("repo not found", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{err: errors.ErrResourceNotFound}, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{1}})
		require.Error(t, err)
		assert.Equal(t, i18n.RepoNotFound, errors.GetI18nKey(err))
	})

	t.Run("solutionIDs required", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a"})
		require.Error(t, err)
	})

	t.Run("invalid branches pattern json", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{
			Name:        "plan-a",
			SolutionIDs: []uint{1},
			Branches:    "{bad-json}",
		})
		require.Error(t, err)
	})

	t.Run("invalid scan type", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{
			Name:        "plan-a",
			SolutionIDs: []uint{1},
			ScanType:    "delta",
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanInvalidScanTypeError, errors.GetI18nKey(err))
	})

	t.Run("solution not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(1)).Return(nil, errors.ErrResourceNotFound)

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{1}})
		require.Error(t, err)
		assert.Equal(t, i18n.SolutionNotFound, errors.GetI18nKey(err))
	})

	t.Run("duplicate name", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(1)).Return(&scanModel.Solution{ID: 1}, nil)
		planStore.EXPECT().FindByName(gomock.Any(), uint(1), "plan-a").Return(&scanModel.ScanPlan{ID: 9}, nil)

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{1}})
		require.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(2)).Return(&scanModel.Solution{ID: 2}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(3)).Return(&scanModel.Solution{ID: 3}, nil)
		planStore.EXPECT().FindByName(gomock.Any(), uint(1), "plan-a").Return(nil, nil)
		planStore.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, plan *scanStore.ScanPlan) (*scanModel.ScanPlan, error) {
			assert.Equal(t, uint64(1), plan.RepoID)
			assert.Equal(t, "tester", plan.CreatedBy)
			assert.Equal(t, "full", plan.ScanType)
			return &scanModel.ScanPlan{ID: 11, RepoID: 1}, nil
		})
		planStore.EXPECT().BindSolutions(gomock.Any(), uint(11), []uint{2, 3}).Return(nil)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(11)).Return(&scanModel.ScanPlan{ID: 11, RepoID: 1, Name: "plan-a"}, nil)
		planStore.EXPECT().GetSolutionIDList(gomock.Any(), uint(11)).Return([]int64{2, 3}, nil)

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{2, 3}, ScanType: "full"})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []int64{2, 3}, got.SolutionIDs)
		assert.Equal(t, 2, got.SolutionCount)
	})
}

func TestScanPlanServiceUpdateDuplicateName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	planStore := scanStoreMock.NewMockIScanPlan(ctrl)
	planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 1, Name: "plan-a"}, nil)
	planStore.EXPECT().FindByName(gomock.Any(), uint(1), "plan-b").Return(&scanModel.ScanPlan{ID: 2, RepoID: 1, Name: "plan-b"}, nil)

	svc := NewScanPlanService(planStore, &fakeRepoStore{}, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		Name: "plan-b",
	})
	require.Error(t, err)
	assert.Equal(t, i18n.PlanAlreadyExists, errors.GetI18nKey(err))
}

func TestScanPlanServiceUpdateValidatePatterns(t *testing.T) {
	t.Parallel()

	svc := NewScanPlanService(nil, &fakeRepoStore{}, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		Branches: "{bad-json}",
	})
	require.Error(t, err)
}

func TestScanPlanServiceUpdateInvalidScanType(t *testing.T) {
	t.Parallel()

	svc := NewScanPlanService(nil, &fakeRepoStore{}, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		ScanType: "delta",
	})
	require.Error(t, err)
	assert.Equal(t, i18n.PlanInvalidScanTypeError, errors.GetI18nKey(err))
}
