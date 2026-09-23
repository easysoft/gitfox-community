// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package canceler

import (
	"context"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"

	"github.com/easysoft/gitfox-community/app/pipeline/scheduler"
	executionStore "github.com/easysoft/gitfox-community/app/store/execution"
	pipelineStore "github.com/easysoft/gitfox-community/app/store/pipeline"
	stageStore "github.com/easysoft/gitfox-community/app/store/stage"
	stepStore "github.com/easysoft/gitfox-community/app/store/step"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

type service struct {
	executionStore executionStore.Interface
	pipelineStore  pipelineStore.Interface
	stageStore     stageStore.Interface
	stepStore      stepStore.Interface
	scheduler      scheduler.Scheduler
}

// Canceler cancels a build.
type Canceler interface {
	// Cancel cancels the provided execution.
	Cancel(ctx context.Context, executionID uint, account string) error
}

// New returns a cancellation service that encapsulates
// all cancellation operations.
func New(
	executionStore executionStore.Interface,
	pipelineStore pipelineStore.Interface,
	stageStore stageStore.Interface,
	stepStore stepStore.Interface,
	scheduler scheduler.Scheduler,
) Canceler {
	return &service{
		executionStore: executionStore,
		pipelineStore:  pipelineStore,
		stageStore:     stageStore,
		stepStore:      stepStore,
		scheduler:      scheduler,
	}
}

//nolint:gocognit // refactor if needed.
func (s *service) Cancel(ctx context.Context, executionID uint, account string) error {

	log := logger.GetLogger(ctx).With(
		"account", account,
		"execution.id", executionID,
	)
	execution, err := s.executionStore.GetByID(ctx, executionID)
	if err != nil {
		log.Errorf("[canceler]: could not get execution: %v", err)
		return err
	}
	// do not cancel the build if the build status is
	// complete. only cancel the build if the status is
	// running or pending.
	if !constant.CIStatus(execution.Status).IsCancelable() {
		return nil
	}

	// update the build status to killed. if the update fails
	// due to an optimistic lock error it means the build has
	// already started, and should now be ignored.
	now := datetime.NowLocalTime()
	execution.Status = constant.CIStatusKilled
	execution.FinishedDate = now
	if execution.StartedDate == nil {
		execution.StartedDate = now
	}

	err = s.executionStore.Update(ctx, execution)
	if err != nil {
		return fmt.Errorf("could not update execution status to canceled: %w", err)
	}

	pipeline, err := s.pipelineStore.GetPipelineByID(ctx, int64(execution.PipelineID))
	if err != nil {
		log.Errorf("[canceler]: could not get pipeline %d: %v", execution.PipelineID, err)
	} else {
		pipeline.LastExec = execution.FinishedDate
		pipeline.LastResult = string(execution.Status)
		if err := s.pipelineStore.UpdatePipeline(ctx, pipeline); err != nil {
			log.Errorf("[canceler]: could not update pipeline %d: %v", pipeline.ID, err)
		}
	}

	stages, err := s.stageStore.ListWithSteps(ctx, execution.ID)
	if err != nil {
		return fmt.Errorf("could not list stages with steps: %w", err)
	}

	// update the status of all steps to indicate they
	// were killed or skipped.
	for _, stage := range stages {
		if stage.Status.IsDone() {
			continue
		}
		if stage.Started != nil {
			stage.Status = constant.CIStatusKilled
			stage.Finished = now
		} else {
			stage.Status = constant.CIStatusSkipped
		}
		err := s.stageStore.Update(ctx, &stage.Stage)
		if err != nil {
			log.Errorf("[canceler]: could not update stage %d status: %v", stage.ID, err)
		}

		// update the status of all steps to indicate they
		// were killed or skipped.
		for _, step := range stage.Steps {
			if step.Status.IsDone() {
				continue
			}
			if step.Started != nil {
				step.Status = constant.CIStatusKilled
				step.Finished = now
			} else {
				step.Status = constant.CIStatusSkipped
			}
			err := s.stepStore.Update(ctx, step)
			if err != nil {
				log.Errorf("[canceler]: could not update step %d status: %v", step.ID, err)
			}
		}
	}

	// notify the scheduler so any runner watching for cancellation
	// requests (via Watch) is woken up immediately and kills the
	// running containers. Without this, the in-memory cancel signal
	// is never delivered and the runner keeps running to completion.
	if err := s.scheduler.Cancel(ctx, int64(execution.ID)); err != nil {
		log.Errorf("[canceler]: could not notify scheduler: %v", err)
	}

	log.Infof("[canceler]: cancelled execution %d successfully", execution.ID)

	return nil
}
