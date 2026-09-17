package scan

import (
	"context"
	"testing"

	"github.com/easysoft/gitfox-community/app/constant"
	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	scanStoreMock "github.com/easysoft/gitfox-community/app/store/scan/mock"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type fakeRepoStore struct {
	repo *repoModel.Repo
	err  error
}

// GetMigrate implements [repo.Interface].
func (f *fakeRepoStore) GetMigrate(ctx context.Context) ([]*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) Create(ctx context.Context, repo *repoStore.Repository) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) Update(ctx context.Context, repo *repoStore.Repository, id uint) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) UpdateOptLock(ctx context.Context, repo *repoModel.Repo, mutateFn func(repo *repoModel.Repo) error) (*repoModel.Repo, error) {
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

func (f *fakeRepoStore) HardDelete(ctx context.Context, id uint) error {
	return nil
}

func (f *fakeRepoStore) GetBySpaceAndIdentifier(ctx context.Context, spaceID uint, identifier string) (*repoModel.Repo, error) {
	return nil, nil
}

func (f *fakeRepoStore) List(ctx context.Context, filter repoModel.ReqFilter) (*database.PageResult[*repoModel.Repo], error) {
	return nil, nil
}

func (f *fakeRepoStore) ListGitRepos(ctx context.Context) ([]*repoModel.Repo, error) {
	return nil, nil
}

func newScanPlanServiceForTest(
	planStore scanStore.IScanPlan,
	solutionStore scanStore.ISolution,
	ruleSetStore scanStore.IRuleSet,
	ruleStore scanStore.IRule,
	triggerStore scanStore.IScanTrigger,
	conditionStore scanStore.IScanPlanCondition,
) *ScanPlanService {
	return NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore, ruleSetStore, ruleStore, triggerStore, conditionStore)
}

func TestScanPlanServiceCreate(t *testing.T) {
	t.Parallel()

	t.Run("repo not found", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{err: errors.ErrResourceNotFound}, nil, nil, nil, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a", SolutionIDs: []uint{1}})
		require.Error(t, err)
		assert.Equal(t, i18n.RepoNotFound, errors.GetI18nKey(err))
	})

	t.Run("solutionIDs required", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil, nil, nil, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{Name: "plan-a"})
		require.Error(t, err)
	})

	t.Run("invalid branches pattern json", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil, nil, nil, nil, nil)
		_, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{
			Name:        "plan-a",
			SolutionIDs: []uint{1},
			Branches: scanModel.Pattern{
				Include: []string{""},
			},
		})
		require.Error(t, err)
		assert.Equal(t, i18n.PlanPatternEmptyError, errors.GetI18nKey(err))
	})

	t.Run("invalid scan type", func(t *testing.T) {
		svc := NewScanPlanService(nil, &fakeRepoStore{}, nil, nil, nil, nil, nil)
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

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore, nil, nil, nil, nil)
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

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore, nil, nil, nil, nil)
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
			assert.JSONEq(t, `{"default":true,"include":["main"],"exclude":["release/*"]}`, plan.Branches)
			assert.JSONEq(t, `{"include":["**/*.go"]}`, plan.Files)
			return &scanModel.ScanPlan{ID: 11, RepoID: 1}, nil
		})
		planStore.EXPECT().BindSolutions(gomock.Any(), uint(11), []uint{2, 3}).Return(nil)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(11)).Return(&scanModel.ScanPlan{
			ID:       11,
			RepoID:   1,
			Name:     "plan-a",
			Branches: &scanModel.Pattern{Default: true, Include: []string{"main"}, Exclude: []string{"release/*"}},
			Files:    &scanModel.Pattern{Include: []string{"**/*.go"}},
		}, nil)
		solutionStore.EXPECT().GetSolutionsByPlanID(gomock.Any(), uint(11)).Return([]*scanModel.Solution{{ID: 2}, {ID: 3}}, []int64{2, 3}, nil)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)
		triggerStore.EXPECT().Count(gomock.Any(), uint(1), uint(11)).Return(int64(0), nil)
		triggerStore.EXPECT().List(gomock.Any(), uint(1), uint(11), scanModel.ScanTriggerReqFilter{Page: 1, PageSize: 1000}).Return(&database.PageResult[*scanModel.ScanTrigger]{Data: nil}, nil)
		conditionStore.EXPECT().ListByPlan(gomock.Any(), uint(1), uint(11)).Return(nil, nil)

		svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore, nil, nil, triggerStore, conditionStore)
		got, err := svc.Create(context.Background(), newTestSession("tester"), 1, scanModel.ScanPlanReqForCreate{
			Name:        "plan-a",
			SolutionIDs: []uint{2, 3},
			ScanType:    "full",
			Branches: scanModel.Pattern{
				Default: true,
				Include: []string{"main"},
				Exclude: []string{"release/*"},
			},
			Files: scanModel.Pattern{
				Include: []string{"**/*.go"},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, []int64{2, 3}, got.SolutionIDs)
		assert.Equal(t, 2, got.SolutionCount)
		require.NotNil(t, got.Branches)
		require.NotNil(t, got.Files)
		assert.True(t, got.Branches.Default)
		assert.Equal(t, []string{"main"}, got.Branches.Include)
		assert.Equal(t, []string{"release/*"}, got.Branches.Exclude)
		assert.Equal(t, []string{"**/*.go"}, got.Files.Include)
	})
}

func TestScanPlanServiceGetByIDEnrichesCounts(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	planStore := scanStoreMock.NewMockIScanPlan(ctrl)
	solutionStore := scanStoreMock.NewMockISolution(ctrl)
	triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
	conditionStore := scanStoreMock.NewMockIScanPlanCondition(ctrl)
	result := constant.ScanTaskResultPass
	status := constant.ScanTaskStatusSuccess

	planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(2), uint(11)).Return(&scanModel.ScanPlan{
		ID:               11,
		RepoID:           2,
		LatestTaskResult: &result,
		LatestTaskStatus: &status,
	}, nil)
	solutionStore.EXPECT().GetSolutionsByPlanID(gomock.Any(), uint(11)).Return([]*scanModel.Solution{{ID: 7}}, []int64{7}, nil)
	triggerStore.EXPECT().Count(gomock.Any(), uint(2), uint(11)).Return(int64(3), nil)
	triggerStore.EXPECT().List(gomock.Any(), uint(2), uint(11), scanModel.ScanTriggerReqFilter{Page: 1, PageSize: 1000}).Return(&database.PageResult[*scanModel.ScanTrigger]{Data: []*scanModel.ScanTrigger{{ID: 21}, {ID: 22}}}, nil)
	conditionStore.EXPECT().ListByPlan(gomock.Any(), uint(2), uint(11)).Return([]*scanModel.ScanCondition{{ID: 31}, {ID: 32}}, nil)

	svc := NewScanPlanService(planStore, &fakeRepoStore{}, solutionStore, nil, nil, triggerStore, conditionStore)
	got, err := svc.GetByID(context.Background(), 2, 11)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, []int64{7}, got.SolutionIDs)
	assert.Equal(t, 1, got.SolutionCount)
	assert.Len(t, got.Solutions, 1)
	assert.Equal(t, 3, got.TriggerCount)
	assert.Equal(t, []int64{21, 22}, got.TriggerIDs)
	assert.Len(t, got.Triggers, 2)
	assert.Equal(t, []int64{31, 32}, got.ConditionIDs)
	assert.Len(t, got.Conditions, 2)
	assert.Equal(t, 2, got.ConditionCount)
	require.NotNil(t, got.LatestTaskResult)
	require.NotNil(t, got.LatestTaskStatus)
}

func TestScanPlanServiceUpdateDuplicateName(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	planStore := scanStoreMock.NewMockIScanPlan(ctrl)
	planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(1)).Return(&scanModel.ScanPlan{ID: 1, RepoID: 1, Name: "plan-a"}, nil)
	planStore.EXPECT().FindByName(gomock.Any(), uint(1), "plan-b").Return(&scanModel.ScanPlan{ID: 2, RepoID: 1, Name: "plan-b"}, nil)

	svc := NewScanPlanService(planStore, &fakeRepoStore{}, nil, nil, nil, nil, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		Name: "plan-b",
	})
	require.Error(t, err)
	assert.Equal(t, i18n.PlanAlreadyExists, errors.GetI18nKey(err))
}

func TestScanPlanServiceUpdateValidatePatterns(t *testing.T) {
	t.Parallel()

	svc := NewScanPlanService(nil, &fakeRepoStore{}, nil, nil, nil, nil, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		Branches: &scanModel.Pattern{
			Include: []string{""},
		},
	})
	require.Error(t, err)
	assert.Equal(t, i18n.PlanPatternEmptyError, errors.GetI18nKey(err))
}

func TestScanPlanServiceUpdateInvalidScanType(t *testing.T) {
	t.Parallel()

	svc := NewScanPlanService(nil, &fakeRepoStore{}, nil, nil, nil, nil, nil)
	_, err := svc.Update(context.Background(), newTestSession("tester"), 1, 1, scanModel.ScanPlanReqForUpdate{
		ScanType: "delta",
	})
	require.Error(t, err)
	assert.Equal(t, i18n.PlanInvalidScanTypeError, errors.GetI18nKey(err))
}

func TestScanPlanServiceListRules(t *testing.T) {
	t.Parallel()

	t.Run("repo not found", func(t *testing.T) {
		svc := newScanPlanServiceForTest(nil, nil, nil, nil, nil, nil)
		svc.repoStore = &fakeRepoStore{err: errors.ErrResourceNotFound}

		_, _, err := svc.ListRules(context.Background(), 1, 2, 0)
		require.Error(t, err)
		assert.Equal(t, i18n.RepoNotFound, errors.GetI18nKey(err))
	})

	t.Run("plan not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(2)).Return(nil, errors.ErrResourceNotFound)

		svc := newScanPlanServiceForTest(planStore, nil, nil, nil, nil, nil)
		_, _, err := svc.ListRules(context.Background(), 1, 2, 0)
		require.Error(t, err)
		assert.Equal(t, i18n.PlanNotFound, errors.GetI18nKey(err))
	})

	t.Run("trigger not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(2)).Return(&scanModel.ScanPlan{ID: 2, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(2), uint(9)).Return(nil, errors.ErrResourceNotFound)

		svc := newScanPlanServiceForTest(planStore, nil, nil, nil, triggerStore, nil)
		_, _, err := svc.ListRules(context.Background(), 1, 2, 9)
		require.Error(t, err)
		assert.Equal(t, i18n.TriggerNotFound, errors.GetI18nKey(err))
	})

	t.Run("no bound solutions returns empty list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(2)).Return(&scanModel.ScanPlan{ID: 2, RepoID: 1}, nil)
		planStore.EXPECT().GetSolutionIDList(gomock.Any(), uint(2)).Return([]int64{}, nil)

		svc := newScanPlanServiceForTest(planStore, nil, nil, nil, nil, nil)
		_, got, err := svc.ListRules(context.Background(), 1, 2, 0)
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("filters disabled entries and deduplicates rules", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)
		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(2)).Return(&scanModel.ScanPlan{ID: 2, RepoID: 1}, nil)
		planStore.EXPECT().GetSolutionIDList(gomock.Any(), uint(2)).Return([]int64{11, 12, 13}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(11)).Return(&scanModel.Solution{ID: 11, Status: constant.ScanRuleStatusEnabled}, nil)
		solutionStore.EXPECT().ListRulesets(gomock.Any(), uint(11)).Return([]*scanModel.RuleSet{{ID: 101, Status: constant.ScanRuleStatusEnabled}, {ID: 102, Status: constant.ScanRuleStatusDisabled}}, nil)
		ruleSetStore.EXPECT().GetRuleIDList(gomock.Any(), uint(101)).Return([]int64{1001, 1002}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(12)).Return(&scanModel.Solution{ID: 12, Status: constant.ScanRuleStatusDisabled}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(13)).Return(&scanModel.Solution{ID: 13, Status: constant.ScanRuleStatusEnabled}, nil)
		solutionStore.EXPECT().ListRulesets(gomock.Any(), uint(13)).Return([]*scanModel.RuleSet{{ID: 103, Status: constant.ScanRuleStatusEnabled}}, nil)
		ruleSetStore.EXPECT().GetRuleIDList(gomock.Any(), uint(103)).Return([]int64{1001, 1003}, nil)
		ruleStore.EXPECT().ListByIDs(gomock.Any(), []int64{1001, 1002, 1003}).Return([]*scanModel.Rule{{ID: 1001, Status: constant.ScanRuleStatusEnabled, Name: "r1"}, {ID: 1002, Status: constant.ScanRuleStatusDisabled, Name: "r2"}, {ID: 1003, Status: constant.ScanRuleStatusEnabled, Name: "r3"}}, nil)

		svc := newScanPlanServiceForTest(planStore, solutionStore, ruleSetStore, ruleStore, nil, nil)
		_, got, err := svc.ListRules(context.Background(), 1, 2, 0)
		require.NoError(t, err)
		require.Len(t, got, 2)
		assert.Equal(t, int64(1001), got[0].ID)
		assert.Equal(t, int64(1003), got[1].ID)
	})

	t.Run("valid trigger still uses plan solutions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		planStore := scanStoreMock.NewMockIScanPlan(ctrl)
		solutionStore := scanStoreMock.NewMockISolution(ctrl)
		ruleSetStore := scanStoreMock.NewMockIRuleSet(ctrl)
		ruleStore := scanStoreMock.NewMockIRule(ctrl)
		triggerStore := scanStoreMock.NewMockIScanTrigger(ctrl)

		planStore.EXPECT().GetByRepoIDAndID(gomock.Any(), uint(1), uint(2)).Return(&scanModel.ScanPlan{ID: 2, RepoID: 1}, nil)
		triggerStore.EXPECT().GetByRepoPlanAndID(gomock.Any(), uint(1), uint(2), uint(8)).Return(&scanModel.ScanTrigger{ID: 8, RepoID: 1, PlanID: 2}, nil)
		planStore.EXPECT().GetSolutionIDList(gomock.Any(), uint(2)).Return([]int64{11}, nil)
		solutionStore.EXPECT().GetByID(gomock.Any(), uint(11)).Return(&scanModel.Solution{ID: 11, Status: constant.ScanRuleStatusEnabled}, nil)
		solutionStore.EXPECT().ListRulesets(gomock.Any(), uint(11)).Return([]*scanModel.RuleSet{{ID: 101, Status: constant.ScanRuleStatusEnabled}}, nil)
		ruleSetStore.EXPECT().GetRuleIDList(gomock.Any(), uint(101)).Return([]int64{1001}, nil)
		ruleStore.EXPECT().ListByIDs(gomock.Any(), []int64{1001}).Return([]*scanModel.Rule{{ID: 1001, Status: constant.ScanRuleStatusEnabled}}, nil)

		svc := newScanPlanServiceForTest(planStore, solutionStore, ruleSetStore, ruleStore, triggerStore, nil)
		_, got, err := svc.ListRules(context.Background(), 1, 2, 8)
		require.NoError(t, err)
		require.Len(t, got, 1)
		assert.Equal(t, int64(1001), got[0].ID)
	})
}
