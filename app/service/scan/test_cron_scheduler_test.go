// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package scan

import (
	"context"
	"time"
)

type fakeCronScheduler struct {
	addCalls    []fakeCronAddCall
	removeCalls []string
	addErr      error
	removeErr   error
}

type fakeCronAddCall struct {
	jobName string
	jobType string
	jobData string
	cronDef string
	maxDur  time.Duration
}

func (f *fakeCronScheduler) AddRecurringWithData(ctx context.Context, jobName, jobType, jobData, cronDef string, maxDur time.Duration) error {
	f.addCalls = append(f.addCalls, fakeCronAddCall{
		jobName: jobName,
		jobType: jobType,
		jobData: jobData,
		cronDef: cronDef,
		maxDur:  maxDur,
	})
	return f.addErr
}

func (f *fakeCronScheduler) RemoveRecurring(ctx context.Context, jobName string) error {
	f.removeCalls = append(f.removeCalls, jobName)
	return f.removeErr
}
