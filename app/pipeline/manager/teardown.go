// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package manager

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/easysoft/gitfox-community/app/constant"
	events "github.com/easysoft/gitfox-community/app/events/pipeline"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	"github.com/easysoft/gitfox-community/pkg/livelog"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"

	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	executionModel "github.com/easysoft/gitfox-community/app/model/execution"
	stageModel "github.com/easysoft/gitfox-community/app/model/stage"
	"github.com/easysoft/gitfox-community/app/pipeline/scheduler"
	"github.com/easysoft/gitfox-community/app/service/execution"
	"github.com/easysoft/gitfox-community/app/service/pipeline"
	"github.com/easysoft/gitfox-community/app/service/pullrequest"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/app/service/stage"
	"github.com/easysoft/gitfox-community/app/service/step"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

// ErrStageTerminal is returned when a stage update is refused because the
// stage is already in a terminal state (e.g. a state written by the
// canceler). The API layer maps this sentinel to an HTTP 409 so the agent
// stops updating the stage instead of overwriting the server's terminal
// state.
var ErrStageTerminal = errors.New("stage already in terminal state")

type teardown struct {
	PipelineSvc  pipeline.Interface
	ExecutionSvc execution.Interface
	StageSvc     stage.Interface
	StepSvc      step.Interface
	PullReqSvc   pullrequest.Interface
	Scheduler    scheduler.Scheduler
	SSEStreamer  sse.Streamer
	Reporter     events.Reporter
	PRReporter   *pullreqevents.Reporter
	Logz         livelog.LogStream
}

//nolint:gocognit // refactor if needed.
func (t *teardown) do(ctx context.Context, stage *stageModel.StageWithSteps) error {
	log := logger.GetLogger(ctx).With(
		"stage.id", stage.ID,
	)
	log.Debug("manager: stage is complete. teardown")

	execution, err := t.ExecutionSvc.GetByID(noContext, stage.ExecutionID)
	if err != nil {
		log.Errorw("manager: cannot find the execution", "error", err)
		return err
	}

	// The canceler may have already written a terminal state (killed/
	// skipped) for this stage while it was executing — or the execution
	// was cancelled while the canceler's stage write was in flight. In
	// both cases the agent's completion report is stale and must not
	// overwrite the server's state: refuse with a conflict error so the
	// agent abandons silently. Check the authoritative DB state, not the
	// agent's request payload.
	existing, err := t.StageSvc.GetByID(noContext, stage.ID)
	if err != nil {
		log.Errorw("manager: cannot find stage", "error", err)
		return err
	}
	if existing.Status.IsDone() || execution.Status == constant.CIStatusKilled {
		// still release the log streams for the steps the agent ran;
		// the terminal stage would otherwise keep them in memory until
		// the server restarts.
		for _, step := range stage.Steps {
			if err := t.Logz.Delete(noContext, int64(step.ID)); err != nil && !errors.Is(err, livelog.ErrStreamNotFound) {
				log.Warnw("manager: cannot delete log stream for terminal stage step",
					"step.id", step.ID, "error", err)
			}
		}
		return fmt.Errorf("%w: stage %d status %s", ErrStageTerminal, stage.ID, existing.Status)
	}

	pipeline, err := t.PipelineSvc.GetPipelineByID(ctx, int64(execution.PipelineID))
	if err != nil {
		log.Errorw("manager: cannot find pipeline", "error", err)
		return err
	}

	log = log.With(
		"execution.id", execution.ID,
		"stage.status", string(stage.Status),
	)

	for _, step := range stage.Steps {
		if len(step.ErrorLog) > 500 {
			step.ErrorLog = step.ErrorLog[:500]
		}
		err := t.StepSvc.Update(noContext, step)
		if err != nil {
			log = log.With(
				"step.name", step.Name,
				"step.id", step.ID,
				"error", err,
			)

			log.Error("manager: cannot persist the step")
			return err
		}
	}

	if len(stage.ErrorLog) > 500 {
		stage.ErrorLog = stage.ErrorLog[:500]
	}

	err = t.StageSvc.Update(noContext, &stage.Stage)
	if err != nil {
		log.Errorw("manager: cannot update the stage",
			"error", err,
		)
		return err
	}

	// TODO: 等 live log
	for _, step := range stage.Steps {
		err = t.Logz.Delete(noContext, int64(step.ID))
		if err != nil && !errors.Is(err, livelog.ErrStreamNotFound) {
			log.Warnw(fmt.Sprintf("failed to delete log stream for step %d", step.ID), "error", err)
		}
	}

	stages, err := t.StageSvc.ListWithSteps(noContext, execution.ID)
	if err != nil {
		log.Warnw("manager: cannot get stages",
			"error", err,
		)
		return err
	}

	err = t.cancelDownstream(ctx, stages)
	if err != nil {
		log.Errorw("manager: cannot cancel downstream builds",
			"error", err,
		)
		return err
	}

	err = t.scheduleDownstream(ctx, stages)
	if err != nil {
		log.Errorw("manager: cannot schedule downstream builds",
			"error", err,
		)
		return err
	}

	if !isexecutionComplete(stages) {
		log.Warnw("manager: execution pending completion of additional stages",
			"error", err,
		)
		return nil
	}

	log.Info("manager: execution is finished, teardown")

	execution.Status = constant.CIStatusSuccess
	execution.FinishedDate = datetime.NowLocalTime()
	for _, sibling := range stages {
		if sibling.Status == constant.CIStatusKilled {
			execution.Status = constant.CIStatusKilled
			break
		}
		if sibling.Status == constant.CIStatusFailure {
			execution.Status = constant.CIStatusFailure
			break
		}
		if sibling.Status == constant.CIStatusError {
			execution.Status = constant.CIStatusError
			break
		}
	}
	if execution.StartedDate == nil {
		execution.StartedDate = execution.FinishedDate
	}

	err = t.ExecutionSvc.Update(noContext, execution)
	if err != nil {
		log.Warnw("manager: cannot update the execution",
			"error", err,
		)
		return err
	}

	pipeline.LastExec = execution.FinishedDate
	pipeline.LastResult = string(execution.Status)
	err = t.PipelineSvc.UpdatePipeline(noContext, pipeline)
	if err != nil {
		log.Warnw("manager: cannot update the pipeline",
			"error", err,
		)
	}

	executionDetail := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	err = t.SSEStreamer.Publish(noContext, pipeline.SpaceID, eventModel.SSETypeExecutionCompleted, executionDetail)
	if err != nil {
		log.Warnw("manager: could not publish execution completed event",
			"error", err,
		)
	}

	// send pipeline execution status
	t.reportExecutionCompleted(ctx, execution)

	return nil
}

// cancelDownstream is a helper function that tests for
// downstream stages and cancels them based on the overall
// pipeline state.
//
//nolint:gocognit // refactor if needed
func (t *teardown) cancelDownstream(
	ctx context.Context,
	stages []*stageModel.StageWithSteps,
) error {
	failed := false
	for _, s := range stages {
		// check pipeline state
		if s.Status.IsFailed() {
			failed = true
		}
	}

	var errs error
	for _, s := range stages {
		if s.Status != constant.CIStatusWaitingOnDeps {
			continue
		}

		var skip bool
		if failed && s.RunPolicy != string(constant.RunPolicyOnFailure) {
			skip = true
		}
		if !failed && s.RunPolicy != string(constant.RunPolicyOnSuccess) {
			skip = true
		}
		if !skip {
			continue
		}

		if !areDepsComplete(&s.Stage, stages) {
			continue
		}

		log := logger.GetLogger(ctx).With(
			"stage.id", s.ID,
			"stage.run_policy", s.RunPolicy,
			"failed", failed,
			"stage.depends_on", strings.Join(s.Depends, ","),
		)

		log.Debug("manager: skipping step")

		s.Status = constant.CIStatusSkipped
		s.Started = datetime.NowLocalTime()
		s.Finished = datetime.NowLocalTime()
		err := t.StageSvc.Update(noContext, &s.Stage)
		if err != nil {
			log.Errorw("manager: cannot update stage status",
				"error", err,
			)
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func isexecutionComplete(stages []*stageModel.StageWithSteps) bool {
	for _, stage := range stages {
		if stage.Status == constant.CIStatusPending ||
			stage.Status == constant.CIStatusRunning ||
			stage.Status == constant.CIStatusWaitingOnDeps ||
			stage.Status == constant.CIStatusDeclined ||
			stage.Status == constant.CIStatusBlocked {
			return false
		}
	}
	return true
}

func areDepsComplete(stage *stageModel.Stage, stages []*stageModel.StageWithSteps) bool {
	deps := map[string]struct{}{}
	for _, dep := range stage.Depends {
		deps[dep] = struct{}{}
	}
	for _, sibling := range stages {
		if _, ok := deps[sibling.Code]; !ok {
			continue
		}
		if !sibling.Status.IsDone() {
			return false
		}
	}
	return true
}

// scheduleDownstream is a helper function that tests for
// downstream stages and schedules stages if all dependencies
// and execution requirements are met.
func (t *teardown) scheduleDownstream(
	ctx context.Context,
	stages []*stageModel.StageWithSteps,
) error {
	var errs error
	for _, sibling := range stages {
		if sibling.Status != constant.CIStatusWaitingOnDeps {
			continue
		}

		if len(sibling.Depends) == 0 {
			continue
		}

		// PROBLEM: isDep only checks the direct parent
		// i think ....
		// if isDep(stage, sibling) == false {
		// 	continue
		// }
		if !areDepsComplete(&sibling.Stage, stages) {
			continue
		}
		// if isLastDep(stage, sibling, stages) == false {
		// 	continue
		// }

		log := logger.GetLogger(ctx).With(
			"stage.id", sibling.ID,
			"stage.name", sibling.Name,
			"stage.code", sibling.Code,
			"stage.depends_on", strings.Join(sibling.Depends, ","),
		)

		log.Debug("manager: schedule next stage")

		sibling.Status = constant.CIStatusPending
		err := t.StageSvc.Update(noContext, &sibling.Stage)
		if err != nil {
			log.Errorw("manager: cannot update stage status",
				"error", err,
			)
			errs = multierror.Append(errs, err)
		}

		err = t.Scheduler.Schedule()
		if err != nil {
			log.Errorw("manager: cannot schedule stage",
				"error", err,
			)
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

// resync updates the stage from the database. Note that it does
// not update the Version field. This is by design. It prevents
// the current go routine from updating a stage that has been
// updated by another go routine.
func (t *teardown) resync(ctx context.Context, stage *stageModel.Stage) error {
	updated, err := t.StageSvc.GetByID(ctx, stage.ID)
	if err != nil {
		return err
	}
	stage.Status = updated.Status
	stage.ErrorLog = updated.ErrorLog
	stage.Host = updated.Host
	stage.Started = updated.Started
	stage.Finished = updated.Finished
	return nil
}

func (t *teardown) reportExecutionCompleted(ctx context.Context, execution *executionModel.Execution) {
	t.Reporter.Executed(ctx, &events.ExecutedPayload{
		PipelineID:   execution.PipelineID,
		ExecutionNum: execution.ID,
	})
}
