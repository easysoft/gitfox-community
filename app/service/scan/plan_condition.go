package scan

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
)

var _ IScanPlanCondition = (*ScanPlanConditionService)(nil)

type ScanPlanConditionService struct {
	repoStore      repoStore.Interface
	planStore      scanStore.IScanPlan
	conditionStore scanStore.IScanPlanCondition
}

func NewScanPlanConditionService(repoStore repoStore.Interface, planStore scanStore.IScanPlan, conditionStore scanStore.IScanPlanCondition) *ScanPlanConditionService {
	return &ScanPlanConditionService{repoStore: repoStore, planStore: planStore, conditionStore: conditionStore}
}

func (s *ScanPlanConditionService) Create(ctx context.Context, session *authModel.Session, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForCreate) (*scanModel.ScanCondition, error) {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return nil, err
	}
	if err := validatePlanConditionCreate(&req); err != nil {
		return nil, err
	}

	unit := req.Unit
	if unit == "" {
		unit = constant.ScanRuleUnitCount
	}

	condition := &scanStore.ScanPlanCondition{
		RepoID:    uint64(repoID),
		PlanID:    uint64(planID),
		TriggerID: 0,
		Priority:  constant.ScanRulePriority(req.Priority),
		Type:      constant.ScanRuleType(req.Type),
		Unit:      constant.ScanRuleUnit(unit),
		Threshold: req.Threshold,
		CreatedBy: session.User.Account,
		EditedBy:  session.User.Account,
	}
	return s.conditionStore.Create(ctx, condition)
}

func (s *ScanPlanConditionService) List(ctx context.Context, repoID uint, planID uint) ([]*scanModel.ScanCondition, error) {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.conditionStore.ListByPlan(ctx, repoID, planID)
}

func (s *ScanPlanConditionService) Find(ctx context.Context, repoID uint, planID uint, conditionID uint) (*scanModel.ScanCondition, error) {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.getByIDAndCheckScope(ctx, repoID, planID, conditionID)
}

func (s *ScanPlanConditionService) Update(ctx context.Context, session *authModel.Session, repoID uint, planID uint, conditionID uint, req scanModel.ScanPlanConditionReqForUpdate) (*scanModel.ScanCondition, error) {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return nil, err
	}
	current, err := s.getByIDAndCheckScope(ctx, repoID, planID, conditionID)
	if err != nil {
		return nil, err
	}
	if err := validatePlanConditionUpdate(&req, current); err != nil {
		return nil, err
	}

	patch := &scanStore.ScanPlanCondition{
		ID:       uint64(conditionID),
		EditedBy: session.User.Account,
	}
	if req.Priority != nil {
		patch.Priority = constant.ScanRulePriority(*req.Priority)
	}
	if req.Type != nil {
		patch.Type = constant.ScanRuleType(*req.Type)
	}
	if req.Unit != nil {
		patch.Unit = constant.ScanRuleUnit(*req.Unit)
	}
	if req.Threshold != nil {
		patch.Threshold = *req.Threshold
	}

	if err := s.conditionStore.Update(ctx, patch); err != nil {
		return nil, err
	}
	return s.getByIDAndCheckScope(ctx, repoID, planID, conditionID)
}

func (s *ScanPlanConditionService) Delete(ctx context.Context, repoID uint, planID uint, conditionID uint) error {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return err
	}
	if _, err := s.getByIDAndCheckScope(ctx, repoID, planID, conditionID); err != nil {
		return err
	}
	return s.conditionStore.Delete(ctx, conditionID)
}

func (s *ScanPlanConditionService) getByIDAndCheckScope(ctx context.Context, repoID uint, planID uint, conditionID uint) (*scanModel.ScanCondition, error) {
	condition, err := s.conditionStore.GetByID(ctx, conditionID)
	if err != nil {
		return nil, err
	}
	if uint(condition.RepoID) != repoID || uint(condition.PlanID) != planID || condition.TriggerID != 0 {
		return nil, errors.WithI18nKey(errors.New("plan condition not found"), i18n.PlanConditionNotFound)
	}
	return condition, nil
}

func (s *ScanPlanConditionService) BatchCreate(ctx context.Context, session *authModel.Session, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForBatchCreate) error {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return err
	}
	if len(req.Conditions) == 0 {
		return errors.WithI18nKey(fmt.Errorf("conditions cannot be empty"), i18n.PlanConditionBodyParamError)
	}

	conditions := make([]*scanStore.ScanPlanCondition, 0, len(req.Conditions))
	for i := range req.Conditions {
		item := req.Conditions[i]
		if err := validatePlanConditionCreate(&item); err != nil {
			return err
		}
		unit := item.Unit
		if unit == "" {
			unit = constant.ScanRuleUnitCount
		}
		conditions = append(conditions, &scanStore.ScanPlanCondition{
			RepoID:    uint64(repoID),
			PlanID:    uint64(planID),
			TriggerID: 0,
			Priority:  constant.ScanRulePriority(item.Priority),
			Type:      constant.ScanRuleType(item.Type),
			Unit:      constant.ScanRuleUnit(unit),
			Threshold: item.Threshold,
			CreatedBy: session.User.Account,
			EditedBy:  session.User.Account,
		})
	}
	return s.conditionStore.BatchCreateByPlan(ctx, conditions)
}

func (s *ScanPlanConditionService) BatchDelete(ctx context.Context, repoID uint, planID uint, req scanModel.ScanPlanConditionReqForBatchDelete) error {
	if err := s.ensureRepoAndPlanExist(ctx, repoID, planID); err != nil {
		return err
	}
	if len(req.IDs) == 0 {
		return errors.WithI18nKey(fmt.Errorf("ids cannot be empty"), i18n.PlanConditionBodyParamError)
	}

	// Strict scope validation: all ids must belong to repo+plan and triggerID=0.
	conditions, err := s.conditionStore.ListByPlan(ctx, repoID, planID)
	if err != nil {
		return err
	}
	allowed := make(map[uint]struct{}, len(conditions))
	for _, c := range conditions {
		allowed[uint(c.ID)] = struct{}{}
	}
	for _, id := range req.IDs {
		if _, ok := allowed[id]; !ok {
			return errors.WithI18nKey(errors.New("plan condition not found"), i18n.PlanConditionNotFound)
		}
	}
	return s.conditionStore.BatchDeleteByIDs(ctx, repoID, planID, req.IDs)
}

func validatePlanConditionCreate(req *scanModel.ScanPlanConditionReqForCreate) error {
	if req == nil {
		return errors.WithI18nKey(fmt.Errorf("request body is required"), i18n.PlanConditionBodyParamError)
	}
	unit := req.Unit
	if unit == "" {
		unit = constant.ScanRuleUnitCount
	}
	switch unit {
	case constant.ScanRuleUnitCount:
		if req.Threshold <= 0 {
			return errors.WithI18nKey(fmt.Errorf("threshold must be greater than 0 for unit count"), i18n.PlanConditionBodyParamError)
		}
	case constant.ScanRuleUnitPercent:
		if req.Threshold < 0 || req.Threshold > 100 {
			return errors.WithI18nKey(fmt.Errorf("threshold must be between 0 and 100 for unit percent"), i18n.PlanConditionBodyParamError)
		}
	default:
		// keep same behavior as reference: let invalid unit be treated as validation error
		return errors.WithI18nKey(fmt.Errorf("invalid unit: %s", unit), i18n.PlanConditionBodyParamError)
	}
	return nil
}

func validatePlanConditionUpdate(req *scanModel.ScanPlanConditionReqForUpdate, current *scanModel.ScanCondition) error {
	if req == nil {
		return errors.WithI18nKey(fmt.Errorf("request body is required"), i18n.PlanConditionBodyParamError)
	}
	unit := current.Unit
	if req.Unit != nil {
		unit = *req.Unit
	}
	threshold := current.Threshold
	if req.Threshold != nil {
		threshold = *req.Threshold
	}
	switch unit {
	case constant.ScanRuleUnitCount:
		if threshold <= 0 {
			return errors.WithI18nKey(fmt.Errorf("threshold must be greater than 0 for unit count"), i18n.PlanConditionBodyParamError)
		}
	case constant.ScanRuleUnitPercent:
		if threshold < 0 || threshold > 100 {
			return errors.WithI18nKey(fmt.Errorf("threshold must be between 0 and 100 for unit percent"), i18n.PlanConditionBodyParamError)
		}
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid unit: %s", unit), i18n.PlanConditionBodyParamError)
	}
	return nil
}

func (s *ScanPlanConditionService) ensureRepoAndPlanExist(ctx context.Context, repoID uint, planID uint) error {
	if _, err := s.repoStore.GetByID(ctx, repoID); err != nil {
		if isDBNotFoundError(err) {
			return errors.WithI18nKey(err, i18n.RepoNotFound)
		}
		return err
	}
	if _, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID); err != nil {
		if isDBNotFoundError(err) {
			return errors.WithI18nKey(err, i18n.PlanNotFound)
		}
		return err
	}
	return nil
}
