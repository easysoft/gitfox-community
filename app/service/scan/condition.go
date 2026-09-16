package scan

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
)

var _ IScanCondition = (*ScanConditionService)(nil)

type ScanConditionService struct {
	conditionStore scanStore.IScanCondition
	repoStore      repoStore.Interface
	planStore      scanStore.IScanPlan
	triggerStore   scanStore.IScanTrigger
	cronScheduler  CronScheduler
}

func NewScanConditionService(conditionStore scanStore.IScanCondition, repoStore repoStore.Interface, planStore scanStore.IScanPlan, triggerStore scanStore.IScanTrigger, cronScheduler CronScheduler) *ScanConditionService {
	return &ScanConditionService{conditionStore: conditionStore, repoStore: repoStore, planStore: planStore, triggerStore: triggerStore, cronScheduler: cronScheduler}
}

func (s *ScanConditionService) BatchCreate(ctx context.Context, session *authModel.Session, planID uint, triggerID uint, repoID uint, req scanModel.ScanConditionReqForBatchCreate) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return err
	}
	if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID); err != nil {
		return err
	}

	conditions := make([]*scanStore.ScanPlanCondition, 0, len(req.Conditions))
	for _, item := range req.Conditions {
		conditions = append(conditions, &scanStore.ScanPlanCondition{
			PlanID:    uint64(planID),
			TriggerID: uint64(triggerID),
			RepoID:    uint64(repoID),
			Priority:  constant.ScanRulePriority(item.Priority),
			Type:      constant.ScanRuleType(item.Type),
			Unit:      constant.ScanRuleUnit(item.Unit),
			Threshold: item.Threshold,
			CreatedBy: session.User.Account,
			EditedBy:  session.User.Account,
		})
	}
	if err := s.conditionStore.BatchCreate(ctx, conditions); err != nil {
		return err
	}
	trigger, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID)
	if err != nil {
		return err
	}
	return s.syncTriggerCronJob(ctx, trigger)
}

func (s *ScanConditionService) BatchDelete(ctx context.Context, repoID uint, planID uint, triggerID uint, req scanModel.ScanConditionReqForBatchDelete) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return err
	}
	if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID); err != nil {
		return err
	}

	conditions, err := s.conditionStore.List(ctx, triggerID)
	if err != nil {
		return err
	}
	conditionIDs := make(map[uint]struct{}, len(req.IDs))
	for _, id := range req.IDs {
		conditionIDs[id] = struct{}{}
	}
	for _, condition := range conditions {
		if uint(condition.RepoID) != repoID || uint(condition.PlanID) != planID {
			continue
		}
		if _, ok := conditionIDs[uint(condition.ID)]; ok {
			delete(conditionIDs, uint(condition.ID))
		}
	}
	if len(conditionIDs) > 0 {
		return errors.WithI18nKey(errors.New("condition not found"), i18n.ConditionNotFound)
	}
	if err := s.conditionStore.BatchDelete(ctx, repoID, planID, triggerID, req.IDs); err != nil {
		return err
	}
	trigger, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID)
	if err != nil {
		return err
	}
	return s.syncTriggerCronJob(ctx, trigger)
}

func (s *ScanConditionService) List(ctx context.Context, repoID uint, planID uint, triggerID uint) ([]*scanModel.ScanCondition, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, triggerID); err != nil {
		return nil, err
	}

	conditions, err := s.conditionStore.List(ctx, triggerID)
	if err != nil {
		return nil, err
	}
	filtered := make([]*scanModel.ScanCondition, 0, len(conditions))
	for _, condition := range conditions {
		if uint(condition.RepoID) == repoID && uint(condition.PlanID) == planID {
			filtered = append(filtered, condition)
		}
	}
	return filtered, nil
}

// ensureRepoExists 校验 repo 是否存在，并把底层 not found 转成 RepoNotFound.
func (s *ScanConditionService) ensureRepoExists(ctx context.Context, repoID uint) error {
	if _, err := s.repoStore.GetByID(ctx, repoID); err != nil {
		if isDBNotFoundError(err) {
			return errors.WithI18nKey(err, i18n.RepoNotFound)
		}
		return err
	}
	return nil
}

// getPlanByRepoAndID 按 repo 作用域加载计划，并把底层 not found 转成 PlanNotFound.
func (s *ScanConditionService) getPlanByRepoAndID(ctx context.Context, repoID uint, planID uint) (*scanModel.ScanPlan, error) {
	plan, err := s.planStore.GetByRepoIDAndID(ctx, repoID, planID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.PlanNotFound)
		}
		return nil, err
	}
	return plan, nil
}

// getTriggerByRepoPlanAndID 按 repo 和 plan 作用域加载触发器，并把底层 not found 转成 TriggerNotFound.
func (s *ScanConditionService) getTriggerByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, triggerID uint) (*scanModel.ScanTrigger, error) {
	trigger, err := s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, triggerID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.TriggerNotFound)
		}
		return nil, err
	}
	return trigger, nil
}
