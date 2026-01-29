package jobs

import (
	"context"
	"fmt"
	"time"

	jobsStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/jobs"
	pkgJobs "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/lock"
	"github.com/rs/zerolog/log"
)

const (
	JobUIDOverdue  = "gitfox:jobs:overdue"
	JobTypeOverdue = "gitfox:jobs:overdue"
	JobCronOverdue = "*/20 * * * *" // every 20 min
)

type jobOverdue struct {
	store     jobsStore.Interface
	mxManager lock.MutexManager
	scheduler *Scheduler
}

func newJobOverdue(store jobsStore.Interface, mxManager lock.MutexManager, scheduler *Scheduler) *jobOverdue {
	return &jobOverdue{
		store:     store,
		mxManager: mxManager,
		scheduler: scheduler,
	}
}

// Handle reclaims overdue jobs. Normally this shouldn't happen.
// But, it can occur if DB update after a job execution fails,
// or the server suddenly terminates while the job is still running.
func (j *jobOverdue) Handle(ctx context.Context, _ string, _ pkgJobs.ProgressReporter) (string, error) {
	mx, err := globalLock(ctx, j.mxManager)
	if err != nil {
		return "", fmt.Errorf("failed to obtain the lock to reclaim overdue jobs")
	}

	defer func() {
		if err := mx.Unlock(ctx); err != nil {
			log.Err(err).Msg("[JobScheduler] failed to release global lock after reclaiming overdue jobs")
		}
	}()

	now := time.Now()
	overdueJobs, err := j.store.ListDeadlineExceeded(ctx, now)
	if err != nil {
		return "", fmt.Errorf("failed to list overdue jobs")
	}

	if len(overdueJobs) == 0 {
		return "", nil
	}

	var minScheduled time.Time

	for _, job := range overdueJobs {
		const errorMessage = "deadline exceeded"
		// Use postExec to handle the job state update with complete logic
		postExec(job, "", errorMessage)

		err = j.store.UpdateExecution(ctx, job)
		if err != nil {
			return "", fmt.Errorf("failed update overdue job")
		}

		if job.State == "scheduled" {
			if job.Scheduled != nil {
				scheduled := time.Time(*job.Scheduled)
				if minScheduled.IsZero() || minScheduled.After(scheduled) {
					minScheduled = scheduled
				}
			}
		}
	}

	if !minScheduled.IsZero() {
		j.scheduler.scheduleProcessing(minScheduled)
	}

	result := fmt.Sprintf("found %d overdue jobs", len(overdueJobs))

	return result, nil
}
