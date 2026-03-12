package scan

import (
	"context"
	"fmt"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/i18n"
	"github.com/gorhill/cronexpr"
)

var _ IScanTrigger = (*ScanTriggerService)(nil)

type ScanTriggerService struct {
	triggerStore  scanStore.IScanTrigger
	repoStore     repoStore.Interface
	planStore     scanStore.IScanPlan
	cronScheduler CronScheduler
}

func NewScanTriggerService(triggerStore scanStore.IScanTrigger, repoStore repoStore.Interface, planStore scanStore.IScanPlan, cronScheduler CronScheduler) *ScanTriggerService {
	return &ScanTriggerService{triggerStore: triggerStore, repoStore: repoStore, planStore: planStore, cronScheduler: cronScheduler}
}

func (s *ScanTriggerService) Create(ctx context.Context, session *authModel.Session, planID uint, repoID uint, req scanModel.ScanTriggerReqForCreate) (*scanModel.ScanTrigger, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	if err := validateTriggerCreate(req); err != nil {
		return nil, err
	}

	trigger := &scanStore.ScanPlanTrigger{
		PlanID:     uint64(planID),
		RepoID:     uint64(repoID),
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       string(req.Type),
		Cron:       req.Cron,
		CronBranch: req.CronBranch,
		Keywords:   req.Keywords,
		Actions:    req.Actions,
		Disabled:   req.Disabled,
		CreatedBy:  session.User.Account,
		EditedBy:   session.User.Account,
	}
	createdTrigger, err := s.triggerStore.Create(ctx, trigger)
	if err != nil {
		return nil, err
	}
	if err := s.syncTriggerCronJob(ctx, createdTrigger); err != nil {
		return nil, err
	}
	return createdTrigger, nil
}

func (s *ScanTriggerService) GetByID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.getTriggerByRepoPlanAndID(ctx, repoID, planID, id)
}

func (s *ScanTriggerService) Update(ctx context.Context, session *authModel.Session, repoID uint, planID uint, id uint, req scanModel.ScanTriggerReqForUpdate) (*scanModel.ScanTrigger, error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, id); err != nil {
		return nil, err
	}
	if err := validateTriggerUpdate(req); err != nil {
		return nil, err
	}

	trigger := &scanStore.ScanPlanTrigger{
		ID:       uint64(id),
		EditedBy: session.User.Account,
	}
	if req.Name != "" {
		trigger.Name = req.Name
	}
	if req.Desc != "" {
		trigger.Desc = req.Desc
	}
	if req.Type != "" {
		trigger.Type = string(req.Type)
	}
	if req.Cron != "" {
		trigger.Cron = req.Cron
	}
	if req.CronBranch != "" {
		trigger.CronBranch = req.CronBranch
	}
	if req.Keywords != "" {
		trigger.Keywords = req.Keywords
	}
	if req.Actions != "" {
		trigger.Actions = req.Actions
	}
	if req.Disabled != nil {
		trigger.Disabled = *req.Disabled
	}

	if err := s.triggerStore.Update(ctx, trigger); err != nil {
		return nil, err
	}
	updatedTrigger, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, id)
	if err != nil {
		return nil, err
	}
	if err := s.syncTriggerCronJob(ctx, updatedTrigger); err != nil {
		return nil, err
	}
	return updatedTrigger, nil
}

func (s *ScanTriggerService) Delete(ctx context.Context, repoID uint, planID uint, id uint) error {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return err
	}
	if _, err := s.getTriggerByRepoPlanAndID(ctx, repoID, planID, id); err != nil {
		return err
	}
	count, err := s.triggerStore.Count(ctx, repoID, planID)
	if err != nil {
		return err
	}
	if count <= 1 {
		return errors.WithI18nKey(fmt.Errorf("at least one trigger must be retained"), i18n.TriggerRetainOneError)
	}
	if err := s.triggerStore.Delete(ctx, id); err != nil {
		return err
	}
	return s.removeTriggerCronJob(ctx, int64(id))
}

func (s *ScanTriggerService) List(ctx context.Context, repoID uint, planID uint, filter scanModel.ScanTriggerReqFilter) (*database.PageResult[*scanModel.ScanTrigger], error) {
	if err := s.ensureRepoExists(ctx, repoID); err != nil {
		return nil, err
	}
	if _, err := s.getPlanByRepoAndID(ctx, repoID, planID); err != nil {
		return nil, err
	}
	return s.triggerStore.List(ctx, repoID, planID, filter)
}

func validateTriggerCreate(req scanModel.ScanTriggerReqForCreate) error {
	if req.Type == "" {
		return errors.WithI18nKey(fmt.Errorf("type is required"), i18n.TriggerTypeRequiredError)
	}
	switch req.Type {
	case scanModel.ScanTriggerTypeCron:
		if req.Cron == "" {
			return errors.WithI18nKey(fmt.Errorf("cron is required when type is cron"), i18n.TriggerCronRequiredError)
		}
		if _, err := cronexpr.Parse(req.Cron); err != nil {
			return errors.WithI18nKey(fmt.Errorf("invalid cron expression: %w", err), i18n.TriggerInvalidCronError)
		}
		if req.CronBranch == "" {
			return errors.WithI18nKey(fmt.Errorf("cronBranch is required when type is cron"), i18n.TriggerCronBranchRequiredError)
		}
	case scanModel.ScanTriggerTypeManual, scanModel.ScanTriggerTypeEvent:
		if req.Actions == "" {
			return errors.WithI18nKey(fmt.Errorf("actions is required when type is %s", req.Type), i18n.TriggerActionsRequiredError)
		}
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid trigger type: %s", req.Type), i18n.TriggerInvalidTypeError)
	}
	return nil
}

func validateTriggerUpdate(req scanModel.ScanTriggerReqForUpdate) error {
	if req.Type == "" {
		return nil
	}
	switch req.Type {
	case scanModel.ScanTriggerTypeCron:
		if req.Cron == "" {
			return errors.WithI18nKey(fmt.Errorf("cron is required when type is cron"), i18n.TriggerCronRequiredError)
		}
		if _, err := cronexpr.Parse(req.Cron); err != nil {
			return errors.WithI18nKey(fmt.Errorf("invalid cron expression: %w", err), i18n.TriggerInvalidCronError)
		}
		if req.CronBranch == "" {
			return errors.WithI18nKey(fmt.Errorf("cronBranch is required when type is cron"), i18n.TriggerCronBranchRequiredError)
		}
	case scanModel.ScanTriggerTypeManual, scanModel.ScanTriggerTypeEvent:
		if req.Actions == "" {
			return errors.WithI18nKey(fmt.Errorf("actions is required when type is %s", req.Type), i18n.TriggerActionsRequiredError)
		}
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid trigger type: %s", req.Type), i18n.TriggerInvalidTypeError)
	}
	return nil
}

// ensureRepoExists 校验 repo 是否存在，并把底层 not found 转成 RepoNotFound.
func (s *ScanTriggerService) ensureRepoExists(ctx context.Context, repoID uint) error {
	if _, err := s.repoStore.GetByID(ctx, repoID); err != nil {
		if isDBNotFoundError(err) {
			return errors.WithI18nKey(err, i18n.RepoNotFound)
		}
		return err
	}
	return nil
}

// getPlanByRepoAndID 按 repo 作用域加载计划，并把底层 not found 转成 PlanNotFound.
func (s *ScanTriggerService) getPlanByRepoAndID(ctx context.Context, repoID uint, planID uint) (*scanModel.ScanPlan, error) {
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
func (s *ScanTriggerService) getTriggerByRepoPlanAndID(ctx context.Context, repoID uint, planID uint, triggerID uint) (*scanModel.ScanTrigger, error) {
	trigger, err := s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, triggerID)
	if err != nil {
		if isDBNotFoundError(err) {
			return nil, errors.WithI18nKey(err, i18n.TriggerNotFound)
		}
		return nil, err
	}
	return trigger, nil
}
