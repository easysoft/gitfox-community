// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	authModel "github.com/easysoft/gitfox-community/app/model/auth"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	scanStore "github.com/easysoft/gitfox-community/app/store/scan"
	"github.com/easysoft/gitfox-community/pkg/database"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/i18n"
	"github.com/gorhill/cronexpr"
)

var _ IScanTrigger = (*ScanTriggerService)(nil)

type ScanTriggerService struct {
	triggerStore         scanStore.IScanTrigger
	triggerSolutionStore scanStore.IScanTriggerSolution
	repoStore            repoStore.Interface
	planStore            scanStore.IScanPlan
	solutionStore        scanStore.ISolution
	cronScheduler        CronScheduler
}

func NewScanTriggerService(triggerStore scanStore.IScanTrigger, triggerSolutionStore scanStore.IScanTriggerSolution, repoStore repoStore.Interface, planStore scanStore.IScanPlan, solutionStore scanStore.ISolution, cronScheduler CronScheduler) *ScanTriggerService {
	return &ScanTriggerService{triggerStore: triggerStore, triggerSolutionStore: triggerSolutionStore, repoStore: repoStore, planStore: planStore, solutionStore: solutionStore, cronScheduler: cronScheduler}
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
	if len(req.SolutionIDs) > 0 {
		if err := s.validateSolutionIDsExist(ctx, req.SolutionIDs); err != nil {
			return nil, err
		}
	}

	trigger := &scanStore.ScanPlanTrigger{
		PlanID:     uint64(planID),
		RepoID:     uint64(repoID),
		Name:       req.Name,
		Desc:       req.Desc,
		Type:       string(req.Type),
		ScanType:   string(req.ScanType),
		Cron:       req.Cron,
		CronBranch: req.CronBranch,
		Keywords:   req.Keywords,
		Actions:    scanStore.JoinActions(deduplicateActions(req.Actions)),
		Disabled:   database.BoolToUint8(req.Disabled),
		CreatedBy:  session.User.Account,
		EditedBy:   session.User.Account,
	}
	createdTrigger, err := s.triggerStore.Create(ctx, trigger)
	if err != nil {
		return nil, err
	}
	if len(req.SolutionIDs) > 0 {
		if err := s.triggerSolutionStore.ReplaceBatch(ctx, uint64(createdTrigger.ID), req.SolutionIDs, session.User.Account); err != nil {
			return nil, err
		}
		createdTrigger.SolutionIDs = req.SolutionIDs
		createdTrigger.SolutionCount = len(req.SolutionIDs)
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
	if len(req.SolutionIDs) > 0 {
		if err := s.validateSolutionIDsExist(ctx, req.SolutionIDs); err != nil {
			return nil, err
		}
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
	if len(req.Actions) > 0 {
		trigger.Actions = scanStore.JoinActions(deduplicateActions(req.Actions))
	}
	if req.ScanType != "" {
		trigger.ScanType = string(req.ScanType)
	}
	if req.Disabled != nil {
		trigger.Disabled = database.BoolToUint8(*req.Disabled)
	}

	if err := s.triggerStore.Update(ctx, trigger); err != nil {
		return nil, err
	}
	if len(req.SolutionIDs) > 0 {
		if err := s.triggerSolutionStore.ReplaceBatch(ctx, uint64(id), req.SolutionIDs, session.User.Account); err != nil {
			return nil, err
		}
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
	if err := s.triggerSolutionStore.DeleteByTriggerID(ctx, uint64(id)); err != nil {
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
		if len(req.Actions) == 0 {
			return errors.WithI18nKey(fmt.Errorf("actions is required when type is %s", req.Type), i18n.TriggerActionsRequiredError)
		}
		if err := checkActions(req.Actions); err != nil {
			return err
		}
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid trigger type: %s", req.Type), i18n.TriggerInvalidTypeError)
	}
	if req.ScanType != "" {
		if _, ok := req.ScanType.Sanitize(); !ok {
			return errors.WithI18nKey(fmt.Errorf("invalid scan type: %s", req.ScanType), i18n.TriggerInvalidTypeError)
		}
	}
	return nil
}

func validateTriggerUpdate(req scanModel.ScanTriggerReqForUpdate) error {
	if req.Type == "" {
		// 没有改 type，但如果传了 actions 也要校验
		if len(req.Actions) > 0 {
			if err := checkActions(req.Actions); err != nil {
				return err
			}
		}
		if req.ScanType != "" {
			if _, ok := req.ScanType.Sanitize(); !ok {
				return errors.WithI18nKey(fmt.Errorf("invalid scan type: %s", req.ScanType), i18n.TriggerInvalidTypeError)
			}
		}
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
		if len(req.Actions) == 0 {
			return errors.WithI18nKey(fmt.Errorf("actions is required when type is %s", req.Type), i18n.TriggerActionsRequiredError)
		}
		if err := checkActions(req.Actions); err != nil {
			return err
		}
	default:
		return errors.WithI18nKey(fmt.Errorf("invalid trigger type: %s", req.Type), i18n.TriggerInvalidTypeError)
	}
	if _, ok := req.ScanType.Sanitize(); !ok {
		return errors.WithI18nKey(fmt.Errorf("invalid scan type: %s", req.ScanType), i18n.TriggerInvalidTypeError)
	}
	return nil
}

// checkActions 校验 actions 是否合法。
func checkActions(actions []constant.TriggerAction) error {
	for _, action := range actions {
		if _, ok := action.Sanitize(); !ok {
			return errors.WithI18nKey(
				fmt.Errorf("invalid trigger action: %s", action),
				i18n.TriggerActionsInvalidError,
			)
		}
	}
	return nil
}

// deduplicateActions 去重 actions。
func deduplicateActions(in []constant.TriggerAction) []constant.TriggerAction {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[constant.TriggerAction]struct{}, len(in))
	out := make([]constant.TriggerAction, 0, len(in))
	for _, a := range in {
		if _, ok := seen[a]; ok {
			continue
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	return out
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

// validateSolutionIDsExist 校验传入的 solutionIDs 是否全部存在。
func (s *ScanTriggerService) validateSolutionIDsExist(ctx context.Context, solutionIDs []int64) error {
	for _, id := range solutionIDs {
		if _, err := s.solutionStore.GetByID(ctx, uint(id)); err != nil {
			if isDBNotFoundError(err) {
				return errors.WithI18nKey(
					fmt.Errorf("solution %d not found", id),
					i18n.SolutionNotFound,
				)
			}
			return err
		}
	}
	return nil
}

// FindByRepoIDAndActions 根据 repoID 和 actions 查询触发器。
func (s *ScanTriggerService) FindByRepoIDAndActions(ctx context.Context, repoID uint, actions []constant.TriggerAction) ([]*scanModel.ScanTrigger, error) {
	return s.triggerStore.FindByRepoIDAndActions(ctx, repoID, actions)
}
