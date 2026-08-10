package manager

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
)

type setup struct {
	SSEStreamer  sse.Streamer
	PipelineSvc  pipeline.Interface
	ExecutionSvc execution.Interface
	StageSvc     stage.Interface
	StepSvc      step.Interface
}

func (s *setup) do(ctx context.Context, stage *stageModel.StageWithSteps) error {
	execution, err := s.ExecutionSvc.GetByID(noContext, stage.ExecutionID)
	if err != nil {
		logger.Errorw("manager: cannot find the execution", "error", err)
		return err
	}

	log := logger.GetLogger(ctx).With(
		"execution.id", execution.ID,
		"stage.id", stage.ID,
	)

	if len(stage.ErrorLog) > 500 {
		stage.ErrorLog = stage.ErrorLog[:500]
	}
	err = s.StageSvc.Update(noContext, &stage.Stage)
	if err != nil {
		log.Errorw("manager: cannot update the stage",
			"error", err,
			"stage.status", string(stage.Status),
		)
		return err
	}

	// TODO: create all the steps as part of a single transaction?
	for _, step := range stage.Steps {
		if len(step.ErrorLog) > 500 {
			step.ErrorLog = step.ErrorLog[:500]
		}
		err := s.StepSvc.Create(noContext, step)
		if err != nil {
			log.Errorw("manager: cannot persist the step",
				"error", err,
				"stage.status", string(stage.Status),
				"step.name", step.Name,
				"step.id", step.ID,
			)
			return err
		}
	}

	_, err = s.updateExecution(noContext, execution)
	if err != nil {
		log.Errorw("manager: cannot update the execution", "error", err)
		return err
	}
	pipeline, err := s.PipelineSvc.GetPipelineByID(ctx, int64(execution.PipelineID))
	if err != nil {
		log.Errorw("manager: cannot find pipeline", "error", err)
		return err
	}
	stages, err := s.StageSvc.ListWithSteps(noContext, execution.ID)
	if err != nil {
		log.Errorw("manager: could not list stages with steps", "error", err)
		return err
	}
	executionDetail := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	err = s.SSEStreamer.Publish(noContext, pipeline.SpaceID, eventModel.SSETypeExecutionRunning, executionDetail)
	if err != nil {
		log.Warnw("manager: could not publish execution event", "error", err)
	}

	return nil
}

// helper function that updates the execution status from pending to running.
// This accounts for the fact that another agent may have already updated
// the execution status, which may happen if two stages execute concurrently.
func (s *setup) updateExecution(ctx context.Context, execution *executionModel.Execution) (bool, error) {
	if execution.Status != constant.CIStatusPending {
		return false, nil
	}
	execution.StartedDate = datetime.NowLocalTime()
	execution.Status = constant.CIStatusRunning
	err := s.ExecutionSvc.Update(ctx, execution)
	if err != nil {
		return false, err
	}
	return true, nil
}
