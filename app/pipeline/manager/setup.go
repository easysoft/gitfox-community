// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package manager

import (
	"context"
	"errors"
	"fmt"

	"github.com/easysoft/gitfox-community/app/constant"
	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
	"github.com/easysoft/gitfox-community/app/service/execution"
	"github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/app/service/stage"
	"github.com/easysoft/gitfox-community/app/service/step"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// ErrStageCancelled is returned when a stage is refused to start because
// the execution was cancelled while the stage was in-flight from the
// queue. The canceler has already written a terminal state (killed/
// skipped) for the stage, so the runner must not overwrite it — the API
// layer maps this sentinel to an HTTP 409 so remote agents recognise the
// stage is no longer theirs.
var ErrStageCancelled = errors.New("stage refused: execution was cancelled")

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

	// do not start the stage if the execution was cancelled while the
	// stage was in-flight from the queue. The canceler has already
	// marked the stage as killed/skipped, so refuse to start and return
	// an error to interrupt the runner.
	if execution.Status == constant.CIStatusKilled {
		log.Warnw("manager: execution was cancelled, refuse to start stage",
			"stage.status", string(stage.Status),
		)
		stage.Status = constant.CIStatusSkipped
		if err := s.StageSvc.Update(noContext, &stage.Stage); err != nil {
			log.Errorw("manager: cannot update stage to skipped", "error", err)
		}
		return fmt.Errorf("%w: execution %d", ErrStageCancelled, execution.ID)
	}

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
