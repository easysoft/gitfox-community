package manager

import (
	"context"

	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

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

	if len(step.ErrorLog) > 500 {
		step.ErrorLog = step.ErrorLog[:500]
	}
	err := u.StepSvc.Update(noContext, step)
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
