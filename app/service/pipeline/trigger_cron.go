package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"github.com/gorhill/cronexpr"
)

const (
	// JobTypePipelineCron is the job type registered in the scheduler for pipeline cron triggers.
	JobTypePipelineCron = "gitfox:pipeline:cron-trigger"

	// maxCronExprs is the upper bound of cron expressions per trigger.
	maxCronExprs = 5

	// cronJobMaxDuration is the maximum execution duration for a single cron trigger job.
	cronJobMaxDuration = 5 * time.Minute
)

// cronJobData is the JSON payload stored in the recurring job.
type cronJobData struct {
	PipelineID int64 `json:"pipelineID"`
}

// cronJobName returns the deterministic job name for a trigger cron entry.
// Keyed by triggerID (not pipelineID) so multiple triggers under the same
// pipeline do not overwrite each other's cron jobs.
func cronJobName(triggerID int64, index int) string {
	return fmt.Sprintf("gitfox:pipeline:cron:%d:%d", triggerID, index)
}

// SyncTriggerCronJobs synchronises the recurring scheduler jobs for a trigger's
// cron expressions.  It first removes all existing cron jobs for the given
// trigger (up to maxCronExprs) and then creates new ones for each expression in
// newCronExprs.
//
// If the CronScheduler is nil the method returns immediately (backward compat).
func (s *Service) SyncTriggerCronJobs(ctx context.Context, triggerID int64, pipelineID int64, newCronExprs []string) error {
	if s.cronScheduler == nil {
		return nil
	}

	// Pre-validate all cron expressions with the same library the scheduler
	// uses (gorhill/cronexpr) to avoid partial state: some jobs created while
	// a later expression turns out to be invalid.
	for i, expr := range newCronExprs {
		if i >= maxCronExprs {
			break
		}
		if _, err := cronexpr.Parse(expr); err != nil {
			return errors.Newf("invalid cron expression %q at index %d: %w", expr, i, err)
		}
	}

	// 1. Remove existing cron jobs for this trigger.
	for i := 0; i < maxCronExprs; i++ {
		if err := s.cronScheduler.RemoveRecurring(ctx, cronJobName(triggerID, i)); err != nil {
			return errors.Newf("failed to remove cron job %s: %w", cronJobName(triggerID, i), err)
		}
	}

	// 2. Add new cron jobs.
	if len(newCronExprs) == 0 {
		// No new cron expressions to add, just return after cleanup.
		return nil
	}

	data, err := json.Marshal(cronJobData{PipelineID: pipelineID})
	if err != nil {
		return errors.Newf("failed to marshal cron job data: %w", err)
	}
	jobData := string(data)

	for i, cronExpr := range newCronExprs {
		if i >= maxCronExprs {
			break
		}
		name := cronJobName(triggerID, i)
		if err := s.cronScheduler.AddRecurringWithData(
			ctx, name, JobTypePipelineCron, jobData, cronExpr, cronJobMaxDuration,
		); err != nil {
			return errors.Newf("failed to add cron job %s: %w", name, err)
		}
	}

	return nil
}

// ReconcileCronJobs iterates over all triggers in the database and ensures
// the schedule_jobs table has the correct recurring jobs for each trigger's
// cron expressions.  This is intended to be called once at server startup,
// before the scheduler's Run() loop begins, so that any inconsistencies
// caused by a previous crash are repaired.
func (s *Service) ReconcileCronJobs(ctx context.Context) error {
	if s.cronScheduler == nil {
		return nil
	}

	triggers, err := s.pipelineStore.GetTriggers(ctx, &pipelineModel.TriggerFilter{})
	if err != nil {
		return errors.Newf("failed to list triggers for cron reconciliation: %w", err)
	}

	for _, t := range triggers {
		// Always sync to handle all cases:
		// 1. Triggers with cron: create/update scheduler jobs
		// 2. Triggers without cron: clean up any orphaned jobs (e.g., user deleted cron)
		if err := s.SyncTriggerCronJobs(ctx, t.ID, t.PipelineID, t.Cron); err != nil {
			return errors.Newf("failed to reconcile cron jobs for trigger %d: %w", t.ID, err)
		}
	}

	return nil
}
