// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"context"
	"fmt"
	"time"

	jobModel "github.com/easysoft/gitfox-community/app/model/jobs"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
	"github.com/gorhill/cronexpr"
)

func (s *Scheduler) AddRecurringWithData(ctx context.Context, jobName, jobType, jobData, cronDef string, maxDur time.Duration) error {
	cronExp, err := cronexpr.Parse(cronDef)
	if err != nil {
		return fmt.Errorf("invalid cron definition string for job type=%s: %w", jobType, err)
	}

	// 必须先转换为 CST 时区，确保 cron 计算基于正确的本地时间
	now := time.Now().In(datetime.CST)
	nowLocal := datetime.Local(now)
	nextExec := cronExp.Next(now)
	nextExecLocal := datetime.Local(nextExec)

	job := &jobModel.ScheduleJob{
		Name:                jobName,
		Type:                jobType,
		Priority:            jobModel.PriorityElevated,
		Data:                jobData,
		Result:              "",
		MaxDurationSeconds:  int(maxDur / time.Second),
		MaxRetries:          0,
		State:               jobModel.StateScheduled,
		Scheduled:           &nextExecLocal,
		TotalExecutions:     0,
		RunBy:               "",
		RunDeadline:         nil,
		RunProgress:         0,
		LastExecuted:        nil,
		IsRecurring:         true,
		RecurringCron:       cronDef,
		ConsecutiveFailures: 0,
		LastFailureError:    "",

		CreatedDate: &nowLocal,
		EditedDate:  &nowLocal,
	}

	err = s.jobStore.Upsert(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to upsert job name=%s type=%s: %w", jobName, jobType, err)
	}

	return nil
}

// RemoveRecurring deletes a job by its name.
func (s *Scheduler) RemoveRecurring(ctx context.Context, jobName string) error {
	err := s.jobStore.DeleteByName(ctx, jobName)
	if err != nil {
		return fmt.Errorf("failed to delete job with name=%s: %w", jobName, err)
	}
	return nil
}
