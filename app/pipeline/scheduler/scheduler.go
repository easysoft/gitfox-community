// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package scheduler

import (
	"context"

	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
	runnerStore "github.com/easysoft/gitfox-community/app/store/runner"
	stageStore "github.com/easysoft/gitfox-community/app/store/stage"
	"github.com/easysoft/gitfox-community/pkg/lock"
)

// Filter provides filter criteria to limit stages requested
// from the scheduler.
type Filter struct {
	Token    string
	Type     string
	OS       string
	Arch     string
	Labels   map[string]string
	Metadata map[string]string
}

// Scheduler schedules Build stages for execution.
type Scheduler interface {
	// Schedule schedules the stage for execution.
	Schedule() error

	// Request requests the next stage scheduled for execution.
	Request(ctx context.Context, filter Filter) (*stageModel.Stage, error)

	// Cancel cancels scheduled or running jobs associated
	// with the parent build ID.
	Cancel(context.Context, int64) error

	// Cancelled blocks and listens for a cancellation event and
	// returns true if the build has been cancelled.
	Cancelled(context.Context, int64) (bool, error)
}

type scheduler struct {
	*queue
	*canceler
}

// New provides an instance of a scheduler with cancel abilities.
func New(stageStore stageStore.Interface, runnerStore runnerStore.Interface, lock lock.MutexManager) (Scheduler, error) {
	q, err := newQueue(stageStore, runnerStore, lock)
	if err != nil {
		return nil, err
	}
	return scheduler{
		q,
		newCanceler(),
	}, nil
}
