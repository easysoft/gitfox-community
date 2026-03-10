package scan

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	scanModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/scan"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/gorhill/cronexpr"
)

const (
	JobTypeScanCron   = "gitfox:scan:cron-trigger"
	scanCronJobPrefix = "gitfox:scan:cron"
	scanCronJobMaxDur = 5 * time.Minute
)

type scanCronJobData struct {
	RepoID    int64 `json:"repoID"`
	PlanID    int64 `json:"planID"`
	TriggerID int64 `json:"triggerID"`
}

func scanCronJobName(triggerID int64) string {
	return fmt.Sprintf("%s:%d", scanCronJobPrefix, triggerID)
}

func (s *ScanTriggerService) syncTriggerCronJob(ctx context.Context, trigger *scanModel.ScanTrigger) error {
	return syncScanTriggerCronJob(ctx, s.cronScheduler, trigger)
}

func (s *ScanTriggerService) removeTriggerCronJob(ctx context.Context, triggerID int64) error {
	return removeScanTriggerCronJob(ctx, s.cronScheduler, triggerID)
}

func (s *ScanConditionService) syncTriggerCronJob(ctx context.Context, trigger *scanModel.ScanTrigger) error {
	return syncScanTriggerCronJob(ctx, s.cronScheduler, trigger)
}

func syncScanTriggerCronJob(ctx context.Context, scheduler CronScheduler, trigger *scanModel.ScanTrigger) error {
	if scheduler == nil {
		return nil
	}
	if trigger == nil {
		return fmt.Errorf("scan trigger is required")
	}
	if !shouldPersistScanCronJob(trigger) {
		return removeScanTriggerCronJob(ctx, scheduler, trigger.ID)
	}
	if _, err := cronexpr.Parse(trigger.Cron); err != nil {
		return errors.Newf("invalid cron expression for scan trigger %d: %v", trigger.ID, err)
	}

	data, err := json.Marshal(scanCronJobData{
		RepoID:    trigger.RepoID,
		PlanID:    trigger.PlanID,
		TriggerID: trigger.ID,
	})
	if err != nil {
		return errors.Newf("failed to marshal scan cron job data for trigger %d: %v", trigger.ID, err)
	}

	if err := scheduler.AddRecurringWithData(
		ctx,
		scanCronJobName(trigger.ID),
		JobTypeScanCron,
		string(data),
		trigger.Cron,
		scanCronJobMaxDur,
	); err != nil {
		return errors.Newf("failed to add scan cron job for trigger %d: %v", trigger.ID, err)
	}

	return nil
}

func removeScanTriggerCronJob(ctx context.Context, scheduler CronScheduler, triggerID int64) error {
	if scheduler == nil {
		return nil
	}
	if err := scheduler.RemoveRecurring(ctx, scanCronJobName(triggerID)); err != nil {
		return errors.Newf("failed to remove scan cron job for trigger %d: %v", triggerID, err)
	}
	return nil
}

func shouldPersistScanCronJob(trigger *scanModel.ScanTrigger) bool {
	if trigger == nil {
		return false
	}
	return trigger.Type == scanModel.ScanTriggerTypeCron && !trigger.Disabled && trigger.Cron != ""
}
