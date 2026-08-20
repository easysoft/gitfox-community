package manager

import (
	"context"
	"errors"
	"fmt"

	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"

	appErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

// ErrStepTerminal is returned when a step update is refused because the
// step is already in a terminal state (e.g. a state written by the
// canceler) or does not exist. The API layer maps this sentinel to an
// HTTP 409 so the agent stops updating the step instead of overwriting
// the server's terminal state.
var ErrStepTerminal = errors.New("step already in terminal state")

type updater struct {
	SSEStreamer  sse.Streamer
	PipelineSvc  pipeline.Interface
	ExecutionSvc execution.Interface
	StageSvc     stage.Interface
	StepSvc      step.Interface
}

func (u *updater) do(ctx context.Context, step *stepModel.Step) error {
	log := logger.GetLogger(ctx).With(
		"step.name", step.Name,
		"step.status", string(step.Status),
		"step.id", int64(step.ID),
	)

	// The canceler may have already written a terminal state (killed/
	// skipped) for this step while it was executing — or the step may
	// not exist at all because the stage was refused to start. In both
	// cases the agent's update must not overwrite the server's state:
	// refuse with a conflict error so the agent abandons silently.
	existing, err := u.StepSvc.GetByID(noContext, step.ID)
	if err != nil {
		if errors.Is(err, appErrors.ErrResourceNotFound) {
			return fmt.Errorf("%w: step %d", ErrStepTerminal, step.ID)
		}
		log.Errorw("manager: cannot find step", "error", err)
		return err
	}
	if existing.Status.IsDone() {
		return fmt.Errorf("%w: step %d status %s", ErrStepTerminal, step.ID, existing.Status)
	}

	if len(step.ErrorLog) > 500 {
		step.ErrorLog = step.ErrorLog[:500]
	}
	err = u.StepSvc.Update(noContext, step)
	if err != nil {
		log.Errorw("manager: cannot update step", "error", err)
		return err
	}

	stage, err := u.StageSvc.GetByID(noContext, step.StageID)
	if err != nil {
		log.Errorw("manager: cannot find stage", "error", err)
		return nil
	}

	execution, err := u.ExecutionSvc.GetByID(noContext, stage.ExecutionID)
	if err != nil {
		log.Errorw("manager: cannot find execution", "error", err)
		return nil
	}

	pipeline, err := u.PipelineSvc.GetPipelineByID(noContext, int64(execution.PipelineID))
	if err != nil {
		log.Errorw("manager: cannot find pipeline", "error", err)
		return nil
	}

	stages, err := u.StageSvc.ListWithSteps(noContext, stage.ExecutionID)
	if err != nil {
		log.Errorw("manager: cannot find stages", "error", err)
		return nil
	}

	executionDetail := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	err = u.SSEStreamer.Publish(noContext, pipeline.SpaceID, eventModel.SSETypeExecutionUpdated, executionDetail)
	if err != nil {
		log.Warnw("manager: cannot publish execution updated event", "error", err)
	}

	return nil
}
