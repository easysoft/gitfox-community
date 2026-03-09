package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/bmatcuk/doublestar/v4"
)

var _ IScanPlan = (*ScanPlanService)(nil)

type ScanPlanService struct {
	planStore     scanStore.IScanPlan
	repoStore     repoStore.Interface
	solutionStore scanStore.ISolution
}

func NewScanPlanService(planStore scanStore.IScanPlan, repoStore repoStore.Interface, solutionStore scanStore.ISolution) *ScanPlanService {
	return &ScanPlanService{planStore: planStore, repoStore: repoStore, solutionStore: solutionStore}
}

func (s *ScanPlanService) Create(ctx context.Context, session *authModel.Session, repoID uint, req scanModel.ScanPlanReqForCreate) (*scanModel.ScanPlan, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if len(req.SolutionIDs) == 0 {
		return nil, errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIDsRequired)
	}
	if err := validateScanPlanPatterns(req.Branches, req.Files); err != nil {
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
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
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
	if err := validateScanPlanPatterns(req.Branches, req.Files); err != nil {
		return nil, err
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
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, id); err != nil {
		return err
	}
	return s.planStore.Delete(ctx, id)
}

func (s *ScanPlanService) List(ctx context.Context, repoID uint, filter scanModel.ScanPlanReqFilter) (*database.PageResult[*scanModel.ScanPlan], error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	return s.planStore.List(ctx, repoID, filter)
}

func (s *ScanPlanService) BindSolutions(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanReqForBind) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if len(req.SolutionIDs) == 0 {
		return errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIDsRequired)
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
		return errors.WithI18nKey(errors.New("solutionIDs is required"), i18n.PlanSolutionIDsRequired)
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

type scanPlanPattern struct {
	Default bool     `json:"default,omitempty"`
	Include []string `json:"include,omitempty"`
	Exclude []string `json:"exclude,omitempty"`
}

// validateScanPlanPatterns 校验扫描计划中的分支和文件模式配置，避免落库前写入非法规则。
func validateScanPlanPatterns(branches string, files string) error {
	if err := validateScanPlanPatternField("branches", branches); err != nil {
		return err
	}
	if err := validateScanPlanPatternField("files", files); err != nil {
		return err
	}
	return nil
}

// validateScanPlanPatternField 解析并校验单个模式字段的 JSON 与 include/exclude 规则。
func validateScanPlanPatternField(field string, raw string) error {
	if raw == "" {
		return nil
	}

	var pattern scanPlanPattern
	if err := json.Unmarshal([]byte(raw), &pattern); err != nil {
		return planPatternError(field, "pattern_json", fmt.Errorf("invalid %s pattern json: %w", field, err))
	}

	if err := validateGlobPatterns(field, "include", pattern.Include); err != nil {
		return err
	}
	if err := validateGlobPatterns(field, "exclude", pattern.Exclude); err != nil {
		return err
	}

	return nil
}

// validateGlobPatterns 逐个校验 glob 模式，并把错误映射成对应字段的 i18n key。
func validateGlobPatterns(field string, patternType string, patterns []string) error {
	for _, pattern := range patterns {
		if pattern == "" {
			return errors.WithI18nKey(fmt.Errorf("%s %s pattern cannot be empty", field, patternType), i18n.PlanPatternEmptyError)
		}
		if _, err := doublestar.Match(pattern, "test"); err != nil {
			return planPatternError(field, patternType, fmt.Errorf("invalid %s %s pattern: %w", field, patternType, err))
		}
	}
	return nil
}

// planPatternError 根据字段和模式类型选择对应的多语言错误 key 并包装原始错误。
func planPatternError(field string, patternType string, err error) error {
	key := i18n.PlanPatternEmptyError
	switch strings.ToLower(field) {
	case "branches":
		switch patternType {
		case "pattern_json":
			key = i18n.PlanBranchesPatternJSONError
		case "include":
			key = i18n.PlanBranchesIncludeError
		case "exclude":
			key = i18n.PlanBranchesExcludeError
		}
	case "files":
		switch patternType {
		case "pattern_json":
			key = i18n.PlanFilesPatternJSONError
		case "include":
			key = i18n.PlanFilesIncludeError
		case "exclude":
			key = i18n.PlanFilesExcludeError
		}
	}
	return errors.WithI18nKey(err, key)
}

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

// isDBNotFoundError 判断错误是否应视为项目内统一的数据库 not found 错误。
func isDBNotFoundError(err error) bool {
	return errors.IsNotFound(err) || errors.GetI18nKey(err) == i18n.DbNotfound
}
