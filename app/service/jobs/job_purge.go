package jobs

import (
	"context"
	"fmt"
	"time"

	jobsStore "github.com/easysoft/gitfox-community/app/store/jobs"
	pkgJobs "github.com/easysoft/gitfox-community/pkg/jobs"
	"github.com/easysoft/gitfox-community/pkg/lock"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

const (
	JobUIDPurge  = "gitfox:jobs:purge"
	JobTypePurge = "gitfox:jobs:purge"
	JobCronPurge = "15 */4 * * *" // every 4 hours at 15 minutes
)

type jobPurge struct {
	store     jobsStore.Interface
	mxManager lock.MutexManager
	minOldAge time.Duration
}

func newJobPurge(store jobsStore.Interface, mxManager lock.MutexManager, minOldAge time.Duration) *jobPurge {
	if minOldAge < 0 {
		minOldAge = 0
	}

	return &jobPurge{
		store:     store,
		mxManager: mxManager,
		minOldAge: minOldAge,
	}
}

func (j *jobPurge) Handle(ctx context.Context, _ string, _ pkgJobs.ProgressReporter) (string, error) {
	mx, err := globalLock(ctx, j.mxManager)
	if err != nil {
		return "", fmt.Errorf("failed to obtain the lock to clean up old jobs")
	}

	defer func() {
		if err := mx.Unlock(ctx); err != nil {
			logger.Errorw("failed to release global lock after cleaning up old jobs", "error", err)
		}
	}()

	olderThan := time.Now().Add(-j.minOldAge)

	n, err := j.store.DeleteOld(ctx, olderThan)
	if err != nil {
		return "", fmt.Errorf("failed to purge old jobs")
	}

	result := "no old jobs found"
	if n > 0 {
		result = fmt.Sprintf("deleted %d old jobs", n)
	}

	return result, nil
}
