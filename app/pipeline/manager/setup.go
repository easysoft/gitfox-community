package manager

import (
	"context"

	"github.com/rs/zerolog/log"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"
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
		log.Error().Err(err).Msg("manager: cannot find the execution")
		return err
	}

	log := log.With().
		Uint("execution.id", execution.ID).
		Uint("stage.id", stage.ID).
		Logger()

	if len(stage.ErrorLog) > 500 {
		stage.ErrorLog = stage.ErrorLog[:500]
	}
	err = s.StageSvc.Update(noContext, &stage.Stage)
	if err != nil {
		log.Error().Err(err).
			Str("stage.status", string(stage.Status)).
			Msg("manager: cannot update the stage")
		return err
	}

	// TODO: create all the steps as part of a single transaction?
	for _, step := range stage.Steps {
		if len(step.ErrorLog) > 500 {
			step.ErrorLog = step.ErrorLog[:500]
		}
		err := s.StepSvc.Create(noContext, step)
		if err != nil {
			log.Error().Err(err).
				Str("stage.status", string(stage.Status)).
				Str("step.name", step.Name).
				Uint("step.id", step.ID).
				Msg("manager: cannot persist the step")
			return err
		}
	}

	_, err = s.updateExecution(noContext, execution)
	if err != nil {
		log.Error().Err(err).Msg("manager: cannot update the execution")
		return err
	}
	pipeline, err := s.PipelineSvc.GetPipelineByID(ctx, int64(execution.PipelineID))
	if err != nil {
		log.Error().Err(err).Msg("manager: cannot find pipeline")
		return err
	}
	stages, err := s.StageSvc.ListWithSteps(noContext, execution.ID)
	if err != nil {
		log.Error().Err(err).Msg("manager: could not list stages with steps")
		return err
	}
	executionDetail := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	err = s.SSEStreamer.Publish(noContext, pipeline.SpaceID, eventModel.SSETypeExecutionRunning, executionDetail)
	if err != nil {
		log.Warn().Err(err).Msg("manager: could not publish execution event")
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
