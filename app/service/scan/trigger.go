package scan

import (
	"context"
	"fmt"

	authModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/auth"
	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	scanStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/scan"
	"github.com/gorhill/cronexpr"
)

var _ IScanTrigger = (*ScanTriggerService)(nil)

type ScanTriggerService struct {
	triggerStore scanStore.IScanTrigger
}

func NewScanTriggerService(triggerStore scanStore.IScanTrigger) *ScanTriggerService {
	return &ScanTriggerService{triggerStore: triggerStore}
}

func (s *ScanTriggerService) Create(ctx context.Context, session *authModel.Session, planID uint, repoID uint, req scanModel.ScanTriggerReqForCreate) (*scanModel.ScanTrigger, error) {
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
	return s.triggerStore.Create(ctx, trigger)
}

func (s *ScanTriggerService) GetByID(ctx context.Context, repoID uint, planID uint, id uint) (*scanModel.ScanTrigger, error) {
	return s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, id)
}

func (s *ScanTriggerService) Update(ctx context.Context, session *authModel.Session, repoID uint, planID uint, id uint, req scanModel.ScanTriggerReqForUpdate) (*scanModel.ScanTrigger, error) {
	if _, err := s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, id); err != nil {
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
	return s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, id)
}

func (s *ScanTriggerService) Delete(ctx context.Context, repoID uint, planID uint, id uint) error {
	if _, err := s.triggerStore.GetByRepoPlanAndID(ctx, repoID, planID, id); err != nil {
		return err
	}
	triggers, err := s.List(ctx, repoID, planID)
	if err != nil {
		return err
	}
	if len(triggers) <= 1 {
		return fmt.Errorf("at least one trigger must be retained")
	}
	return s.triggerStore.Delete(ctx, id)
}

func (s *ScanTriggerService) List(ctx context.Context, repoID uint, planID uint) ([]*scanModel.ScanTrigger, error) {
	triggers, err := s.triggerStore.List(ctx, planID)
	if err != nil {
		return nil, err
	}
	filtered := make([]*scanModel.ScanTrigger, 0, len(triggers))
	for _, trigger := range triggers {
		if uint(trigger.RepoID) == repoID {
			filtered = append(filtered, trigger)
		}
	}
	return filtered, nil
}

func validateTriggerCreate(req scanModel.ScanTriggerReqForCreate) error {
	if req.Type == "" {
		return fmt.Errorf("type is required")
	}
	switch req.Type {
	case scanModel.ScanTriggerTypeCron:
		if req.Cron == "" {
			return fmt.Errorf("cron is required when type is cron")
		}
		if _, err := cronexpr.Parse(req.Cron); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		if req.CronBranch == "" {
			return fmt.Errorf("cronBranch is required when type is cron")
		}
	case scanModel.ScanTriggerTypeManual, scanModel.ScanTriggerTypeEvent:
		if req.Actions == "" {
			return fmt.Errorf("actions is required when type is %s", req.Type)
		}
	default:
		return fmt.Errorf("invalid trigger type: %s", req.Type)
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
			return fmt.Errorf("cron is required when type is cron")
		}
		if _, err := cronexpr.Parse(req.Cron); err != nil {
			return fmt.Errorf("invalid cron expression: %w", err)
		}
		if req.CronBranch == "" {
			return fmt.Errorf("cronBranch is required when type is cron")
		}
	case scanModel.ScanTriggerTypeManual, scanModel.ScanTriggerTypeEvent:
		if req.Actions == "" {
			return fmt.Errorf("actions is required when type is %s", req.Type)
		}
	default:
		return fmt.Errorf("invalid trigger type: %s", req.Type)
	}
	return nil
}
