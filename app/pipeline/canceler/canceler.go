package canceler

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"

	executionStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/execution"
	stageStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/stage"
	stepStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type service struct {
	executionStore executionStore.Interface
	stageStore     stageStore.Interface
	stepStore      stepStore.Interface
	// scheduler      scheduler.Scheduler
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
	stageStore stageStore.Interface,
	stepStore stepStore.Interface,
	// scheduler scheduler.Scheduler,
) Canceler {
	return &service{
		executionStore: executionStore,
		stageStore:     stageStore,
		stepStore:      stepStore,
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
	execution.Finished = now
	if execution.Started == nil {
		execution.Started = now
	}

	err = s.executionStore.Update(ctx, execution)
	if err != nil {
		return fmt.Errorf("could not update execution status to canceled: %w", err)
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
		} else {
			stage.Status = constant.CIStatusSkipped
			stage.Started = now
		}
		stage.Finished = now
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
			} else {
				step.Status = constant.CIStatusSkipped
				step.Started = now
			}
			step.Finished = now
			err := s.stepStore.Update(ctx, &step)
			if err != nil {
				log.Errorf("[canceler]: could not update step %d status: %v", step.ID, err)
			}
		}
	}

	log.Infof("[canceler]: cancelled execution %d successfully", execution.ID)

	return nil
}
