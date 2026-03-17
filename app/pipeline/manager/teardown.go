package manager

import (
	"context"
	"errors"
	"strings"

	// "github.com/easysoft/gitfox/livelog"

	"github.com/hashicorp/go-multierror"
	"github.com/rs/zerolog/log"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	events "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pipeline"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"

	eventModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"
	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/scheduler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pullrequest"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"
)

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
	log := log.With().
		Uint("stage.id", stage.ID).
		Logger()
	log.Debug().Msg("manager: stage is complete. teardown")

	execution, err := t.ExecutionSvc.GetByID(noContext, stage.ExecutionID)
	if err != nil {
		log.Error().Err(err).Msg("manager: cannot find the execution")
		return err
	}

	pipeline, err := t.PipelineSvc.GetPipelineByID(ctx, int64(execution.PipelineID))
	if err != nil {
		log.Error().Err(err).Msg("manager: cannot find pipeline")
		return err
	}

	log = log.With().
		Uint("execution.id", execution.ID).
		Str("stage.status", string(stage.Status)).
		Logger()

	for _, step := range stage.Steps {
		if len(step.ErrorLog) > 500 {
			step.ErrorLog = step.ErrorLog[:500]
		}
		err := t.StepSvc.Update(noContext, step)
		if err != nil {
			log = log.With().
				Str("step.name", step.Name).
				Uint("step.id", step.ID).
				Err(err).
				Logger()

			log.Error().Msg("manager: cannot persist the step")
			return err
		}
	}

	if len(stage.ErrorLog) > 500 {
		stage.ErrorLog = stage.ErrorLog[:500]
	}

	err = t.StageSvc.Update(noContext, &stage.Stage)
	if err != nil {
		log.Error().Err(err).
			Msg("manager: cannot update the stage")
		return err
	}

	// TODO: 等 live log
	for _, step := range stage.Steps {
		err = t.Logz.Delete(noContext, int64(step.ID))
		if err != nil && !errors.Is(err, livelog.ErrStreamNotFound) {
			log.Warn().Err(err).Msgf("failed to delete log stream for step %d", step.ID)
		}
	}

	stages, err := t.StageSvc.ListWithSteps(noContext, execution.ID)
	if err != nil {
		log.Warn().Err(err).
			Msg("manager: cannot get stages")
		return err
	}

	err = t.cancelDownstream(ctx, stages)
	if err != nil {
		log.Error().Err(err).
			Msg("manager: cannot cancel downstream builds")
		return err
	}

	err = t.scheduleDownstream(ctx, stages)
	if err != nil {
		log.Error().Err(err).
			Msg("manager: cannot schedule downstream builds")
		return err
	}

	if !isexecutionComplete(stages) {
		log.Warn().Err(err).
			Msg("manager: execution pending completion of additional stages")
		return nil
	}

	log.Info().Msg("manager: execution is finished, teardown")

	execution.Status = constant.CIStatusSuccess
	execution.Finished = datetime.NowLocalTime()
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
	if execution.Started == nil {
		execution.Started = execution.Finished
	}

	err = t.ExecutionSvc.Update(noContext, execution)
	if err != nil {
		log.Warn().Err(err).
			Msg("manager: cannot update the execution")
		return err
	}

	executionDetail := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	err = t.SSEStreamer.Publish(noContext, pipeline.SpaceID, eventModel.SSETypeExecutionCompleted, executionDetail)
	if err != nil {
		log.Warn().Err(err).
			Msg("manager: could not publish execution completed event")
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

		log := log.With().
			Uint("stage.id", s.ID).
			Str("stage.run_policy", s.RunPolicy).
			Bool("failed", failed).
			Str("stage.depends_on", strings.Join(s.Depends, ",")).
			Logger()

		log.Debug().Msg("manager: skipping step")

		s.Status = constant.CIStatusSkipped
		s.Started = datetime.NowLocalTime()
		s.Finished = datetime.NowLocalTime()
		err := t.StageSvc.Update(noContext, &s.Stage)
		if err != nil {
			log.Error().Err(err).
				Msg("manager: cannot update stage status")
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
		if _, ok := deps[sibling.Name]; !ok {
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

		log := log.With().
			Uint("stage.id", sibling.ID).
			Str("stage.name", sibling.Name).
			Str("stage.depends_on", strings.Join(sibling.Depends, ",")).
			Logger()

		log.Debug().Msg("manager: schedule next stage")

		sibling.Status = constant.CIStatusPending
		err := t.StageSvc.Update(noContext, &sibling.Stage)
		if err != nil {
			log.Error().Err(err).
				Msg("manager: cannot update stage status")
			errs = multierror.Append(errs, err)
		}

		err = t.Scheduler.Schedule()
		if err != nil {
			log.Error().Err(err).
				Msg("manager: cannot schedule stage")
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
