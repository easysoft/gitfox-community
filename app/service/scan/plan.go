package scan

import (
	"context"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
)

var _ IScanPlan = (*ScanPlanService)(nil)

type ScanPlanService struct {
	planStore scanStore.IScanPlan
}

func NewScanPlanService(planStore scanStore.IScanPlan) *ScanPlanService {
	return &ScanPlanService{planStore: planStore}
}

func (s *ScanPlanService) Create(ctx context.Context, session *authModel.Session, repoID uint, req scanModel.ScanPlanReqForCreate) (*scanModel.ScanPlan, error) {
	if len(req.SolutionIDs) == 0 {
		return nil, errors.New("solutionIDs is required")
	}

	existing, err := s.planStore.FindByName(ctx, repoID, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("scan plan name already exists")
	}

	plan := &scanStore.ScanPlan{
		RepoID:    uint64(repoID),
		Name:      req.Name,
		Desc:      req.Desc,
		ScanType:  req.ScanType,
		Branches:  req.Branches,
		Files:     req.Files,
		CreatedBy: session.User.Account,
		EditedBy:  session.User.Account,
	}

	created, err := s.planStore.Create(ctx, plan)
	if err != nil {
		return nil, err
	}

	// 如果创建时附带了方案ID，自动绑定
	if len(req.SolutionIDs) > 0 {
		if err := s.planStore.BindSolutions(ctx, uint(created.ID), req.SolutionIDs); err != nil {
			return nil, err
		}
	}

	return s.GetByID(ctx, repoID, uint(created.ID))
}

func (s *ScanPlanService) GetByID(ctx context.Context, repoID uint, id uint) (*scanModel.ScanPlan, error) {
	plan, err := s.planStore.GetByRepoIDAndID(ctx, repoID, id)
	if err != nil {
		return nil, err
	}
	solutionIDs, err := s.planStore.GetSolutionIDList(ctx, id)
	if err != nil {
		return nil, err
	}
	plan.SolutionIDs = solutionIDs
	plan.SolutionCount = len(solutionIDs)
	return plan, nil
}

func (s *ScanPlanService) Update(ctx context.Context, session *authModel.Session, repoID uint, id uint, req scanModel.ScanPlanReqForUpdate) (*scanModel.ScanPlan, error) {
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, id); err != nil {
		return nil, err
	}

	plan := &scanStore.ScanPlan{
		ID:       uint64(id),
		EditedBy: session.User.Account,
	}
	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.Desc != "" {
		plan.Desc = req.Desc
	}
	if req.ScanType != "" {
		plan.ScanType = req.ScanType
	}
	if req.Branches != "" {
		plan.Branches = req.Branches
	}
	if req.Files != "" {
		plan.Files = req.Files
	}

	if err := s.planStore.Update(ctx, plan); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, repoID, id)
}

func (s *ScanPlanService) Delete(ctx context.Context, repoID uint, id uint) error {
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, id); err != nil {
		return err
	}
	return s.planStore.Delete(ctx, id)
}

func (s *ScanPlanService) List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error) {
	return s.planStore.List(ctx, repoID, filter)
}

func (s *ScanPlanService) BindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error {
	if len(req.SolutionIDs) == 0 {
		return errors.New("solutionIDs is required")
	}
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID); err != nil {
		return err
	}
	return s.planStore.BindSolutions(ctx, planID, req.SolutionIDs)
}

func (s *ScanPlanService) UnbindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error {
	if len(req.SolutionIDs) == 0 {
		return errors.New("solutionIDs is required")
	}
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID); err != nil {
		return err
	}
	return s.planStore.UnbindSolutions(ctx, planID, req.SolutionIDs)
}

func (s *ScanPlanService) ListUnbindSolutions(ctx context.Context, repoID uint, planID uint) ([]*scanModel.Solution, error) {
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.planStore.ListUnbindSolutions(ctx, planID)
}
