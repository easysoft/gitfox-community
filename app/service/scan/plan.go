package scan

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

var _ IScanPlan = (*ScanPlanService)(nil)

type ScanPlanService struct {
	planStore     scanStore.IScanPlan
	repoStore     repoStore.Interface
	solutionStore scanStore.ISolution
	ruleSetStore  scanStore.IRuleSet
	ruleStore     scanStore.IRule
	triggerStore  scanStore.IScanTrigger
}

func NewScanPlanService(
	planStore scanStore.IScanPlan,
	repoStore repoStore.Interface,
	solutionStore scanStore.ISolution,
	ruleSetStore scanStore.IRuleSet,
	ruleStore scanStore.IRule,
	triggerStore scanStore.IScanTrigger,
) *ScanPlanService {
	return &ScanPlanService{
		planStore:     planStore,
		repoStore:     repoStore,
		solutionStore: solutionStore,
		ruleSetStore:  ruleSetStore,
		ruleStore:     ruleStore,
		triggerStore:  triggerStore,
	}
}

func (s *ScanPlanService) Create(ctx context.Context, session *authModel.Session, repoID uint, req scanModel.ScanPlanReqForCreate) (*scanModel.ScanPlan, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if len(req.SolutionIDs) == 0 {
		return nil, errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIdsRequired)
	}
	branches, err := validateAndMarshalScanPlanPattern("branches", req.Branches)
	if err != nil {
		return nil, err
	}
	files, err := validateAndMarshalScanPlanPattern("files", req.Files)
	if err != nil {
		return nil, err
	}
	if err := validateScanType(req.ScanType); err != nil {
		return nil, err
	}
	if err := s.validateSolutionIDsExist(ctx, req.SolutionIDs); err != nil {
		return nil, err
	}

	existing, err := s.planStore.FindByName(ctx, repoID, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.WithI18nKey(errors.New("scan plan name already exists"), i18n.PlanAlreadyExists)
	}

	plan := &scanStore.ScanPlan{
		RepoID:    uint64(repoID),
		Name:      req.Name,
		Desc:      req.Desc,
		ScanType:  req.ScanType,
		Branches:  branches,
		Files:     files,
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
	if repoID != 0 {
		if err := s.ensureRepoExists(ctx, repoID); err != nil {
			return nil, err
		}
	}
	plan, err := s.getPlanByRepoAndID(ctx, repoID, id)
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
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if req.Branches != nil {
		if _, err := validateAndMarshalScanPlanPattern("branches", *req.Branches); err != nil {
			return nil, err
		}
	}
	if req.Files != nil {
		if _, err := validateAndMarshalScanPlanPattern("files", *req.Files); err != nil {
			return nil, err
		}
	}
	if req.ScanType != "" {
		if err := validateScanType(req.ScanType); err != nil {
			return nil, err
		}
	}

	if _, err := s.getPlanByRepoAndID(ctx, repoID, id); err != nil {
		return nil, err
	}
	if req.Name != "" {
		existing, err := s.planStore.FindByName(ctx, repoID, req.Name)
		if err != nil {
			return nil, err
		}
		if existing != nil && uint(existing.ID) != id {
			return nil, errors.WithI18nKey(errors.New("scan plan name already exists"), i18n.PlanAlreadyExists)
		}
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
	if req.Branches != nil {
		branches, err := validateAndMarshalScanPlanPattern("branches", *req.Branches)
		if err != nil {
			return nil, err
		}
		plan.Branches = branches
	}
	if req.Files != nil {
		files, err := validateAndMarshalScanPlanPattern("files", *req.Files)
		if err != nil {
			return nil, err
		}
		plan.Files = files
	}

	if err := s.planStore.Update(ctx, plan); err != nil {
		return nil, err
	}
	return s.GetByID(ctx, repoID, id)
}

func (s *ScanPlanService) Delete(ctx context.Context, repoID uint, id uint) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, id); err != nil {
		return err
	}
	return s.planStore.Delete(ctx, id)
}

func (s *ScanPlanService) List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error) {
	if repoID != 0 {
		if err := s.ensureRepoExists(ctx, repoID); err != nil {
			return nil, err
		}
	}

	return s.planStore.List(ctx, repoID, filter)
}

func (s *ScanPlanService) ListRules(ctx context.Context, repoID uint, planID uint, triggerID uint) ([]*scanModel.Rule, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	if triggerID != 0 {
		if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID); err != nil {
			return nil, err
		}
	}

	solutionIDs, err := s.planStore.GetSolutionIDList(ctx, planID)
	if err != nil {
		return nil, err
	}
	if len(solutionIDs) == 0 {
		return []*scanModel.Rule{}, nil
	}

	ruleIDs := make([]int64, 0)
	seenRuleIDs := make(map[int64]struct{})
	for _, solutionID := range solutionIDs {
		solution, err := s.getSolutionByID(ctx, uint(solutionID))
		if err != nil {
			return nil, err
		}
		if solution.Status != constant.ScanRuleStatusEnabled {
			continue
		}

		rulesets, err := s.solutionStore.ListRulesets(ctx, uint(solutionID))
		if err != nil {
			return nil, err
		}
		for _, ruleset := range rulesets {
			if ruleset == nil || ruleset.Deleted || ruleset.Status != constant.ScanRuleStatusEnabled {
				continue
			}

			rulesetRuleIDs, err := s.ruleSetStore.GetRuleIDList(ctx, uint(ruleset.ID))
			if err != nil {
				return nil, err
			}
			for _, ruleID := range rulesetRuleIDs {
				if _, ok := seenRuleIDs[ruleID]; ok {
					continue
				}
				seenRuleIDs[ruleID] = struct{}{}
				ruleIDs = append(ruleIDs, ruleID)
			}
		}
	}
	if len(ruleIDs) == 0 {
		return []*scanModel.Rule{}, nil
	}

	rules, err := s.ruleStore.ListByIDs(ctx, ruleIDs)
	if err != nil {
		return nil, err
	}

	result := make([]*scanModel.Rule, 0, len(rules))
	for _, rule := range rules {
		if rule == nil || rule.Status != constant.ScanRuleStatusEnabled {
			continue
		}
		result = append(result, rule)
	}
	return result, nil
}

func (s *ScanPlanService) BindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if len(req.SolutionIDs) == 0 {
		return errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIdsRequired)
	}
	if err := s.validateSolutionIDsExist(ctx, req.SolutionIDs); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return err
	}
	return s.planStore.BindSolutions(ctx, planID, req.SolutionIDs)
}

func (s *ScanPlanService) UnbindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if len(req.SolutionIDs) == 0 {
		return errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIdsRequired)
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return err
	}
	return s.planStore.UnbindSolutions(ctx, planID, req.SolutionIDs)
}

func (s *ScanPlanService) ListUnbindSolutions(ctx context.Context, repoID uint, planID uint) ([]*scanModel.Solution, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.planStore.ListUnbindSolutions(ctx, planID)
}

// validateAndMarshalScanPlanPattern 校验并序列化 ScanPlan 的 Pattern 字段
func validateAndMarshalScanPlanPattern(field string, pattern scanModel.Pattern) (string, error) {
	if err := pattern.Validate(field); err != nil {
		return "", err
	}

	return pattern.JSONString()
}

// validateScanType 校验 ScanPlan 的 ScanType 是否有效
func validateScanType(scanType string) error {
	if scanType == "" {
		return nil
	}

	switch constant.ScanType(scanType) {
	case constant.ScanTypeFull, constant.ScanTypeIncremental:
		return nil
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid scan type: %s", scanType), i18n.PlanInvalidScanTypeError)
	}
}

// validateSolutionIDsExist 校验扫描方案是否存在，并把底层 not found 转成 SolutionNotFound。
func (s *ScanPlanService) validateSolutionIDsExist(ctx context.Context, solutionIDs []uint) error {
	for _, solutionID := range solutionIDs {
		if _, err := s.getSolutionByID(ctx, solutionID); err != nil {
			return err
		}
	}
	return nil
}

// ensureRepoExists 校验 repo 是否存在，并把底层 not found 转成 RepoNotFound。
func (s *ScanPlanService) ensureRepoExists(ctx context.Context, repoID uint) error {
	if _, err := s.repoStore.GetByID(ctx, repoID); err != nil {
		if isDBNotFoundError(err) {
			return errors.WithI18nKey(err, i18n.RepoNotFound)
		}
		return err
	}
	return nil
}

// getPlanByRepoAndID 按 repo 作用域加载计划，并把底层 not found 转成 PlanNotFound。
func (s *ScanPlanService) getPlanByRepoAndID(ctx context.Context, repoID uint, planID uint) (*scanModel.ScanPlan, error) {
	plan, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.PlanNotFound)
		}
		return nil, err
	}
	return plan, nil
}

// getSolutionByID 加载扫描方案，并把底层 not found 转成 SolutionNotFound。
func (s *ScanPlanService) getSolutionByID(ctx context.Context, solutionID uint) (*scanModel.Solution, error) {
	solution, err := s.solutionStore.GetByID(ctx, solutionID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.SolutionNotFound)
		}
		return nil, err
	}
	return solution, nil
}

// getTriggerByRepoPlanAndID 按 repo 和 plan 作用域加载触发器，并把底层 not found 转成 TriggerNotFound.
func (s *ScanPlanService) getTriggerByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, triggerID uint) (*scanModel.ScanTrigger, error) {
	trigger, err := s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, triggerID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.TriggerNotFound)
		}
		return nil, err
	}
	return trigger, nil
}

// isDBNotFoundError 判断错误是否应视为项目内统一的数据库 not found 错误。
func isDBNotFoundError(err error) bool {
	return errors.IsNotFound(err) || errors.GetI18nKey(err) == i18n.DbNotfound
}
