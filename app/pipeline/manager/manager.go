package manager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	events "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pipeline"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"

	executionModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/execution"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	stageModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/stage"
	stepModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/scheduler"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/execution"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pullrequest"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/stage"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/step"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/livelog"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	urlprovider "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/url"
)

var noContext = context.Background()

var _ ExecutionManager = (*Manager)(nil)

type (
	// Request provides filters when requesting a pending
	// build from the queue. This allows an agent, for example,
	// to request a build that matches its architecture and kernel.
	Request struct {
		Token    string            `json:"token"`
		Type     string            `json:"type"`
		OS       string            `json:"os"`
		Arch     string            `json:"arch"`
		Labels   map[string]string `json:"labels,omitempty"`
		Metadata map[string]string `json:"metadata,omitempty"`
	}

	// Config represents a pipeline config file.
	Config struct {
		Data string `json:"data"`
	}

	// Netrc contains login and initialization information used
	// by an automated login process.
	Netrc struct {
		Machine  string `json:"machine"`
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	// ExecutionContext represents the minimum amount of information
	// required by the runner to execute a build.
	ExecutionContext struct {
		Execution *executionModel.ExecutionWithStages `json:"build"`
		Stage     *stageModel.StageWithSteps          `json:"stage"`
		Config    *pipelineModel.PipelineFile         `json:"config"`
		Netrc     *Netrc                              `json:"netrc"`
		// Secrets   []*types.Secret           `json:"secrets"`
	}

	// ExecutionManager encapsulates complex build operations and provides
	// a simplified interface for build runners.
	ExecutionManager interface {
		// Request requests the next available build stage for execution.
		Request(ctx context.Context, args *Request) (*stageModel.Stage, error)

		// Watch watches for build cancellation requests.
		Watch(ctx context.Context, executionID int64) (bool, error)

		// Accept accepts the build stage for execution.
		Accept(ctx context.Context, stage int64, machine string) (*stageModel.Stage, error)

		// Write writes a line to the build logs.
		Write(ctx context.Context, step int64, line *livelog.Line) error

		// StepStartTime returns the time the step started, used as the base for
		// converting the relative log timestamps reported by runners.
		// It returns the zero time when the base is unavailable.
		StepStartTime(ctx context.Context, stepID int64) time.Time

		// Details returns details about stage.
		Details(ctx context.Context, stageID int64) (*ExecutionContext, error)

		// UploadLogs uploads the full logs.
		UploadLogs(ctx context.Context, stepID int64, r io.Reader) error

		// BeforeStep signals the build step is about to start.
		BeforeStep(ctx context.Context, step *stepModel.Step) error

		// AfterStep signals the build step is complete.
		AfterStep(ctx context.Context, step *stepModel.Step) error

		// BeforeStage signals the build stage is about to start.
		BeforeStage(ctx context.Context, stage *stageModel.StageWithSteps) error

		// AfterStage signals the build stage is complete.
		AfterStage(ctx context.Context, stage *stageModel.StageWithSteps) error
	}
)

// Manager provides a simplified interface to the build runner so that it
// can more easily interact with the server.
type Manager struct {
	Config       *config.Config
	urlProvider  urlprovider.Provider
	reporter     events.Reporter
	prReporter   *pullreqevents.Reporter
	SSEStreamer  sse.Streamer
	pullReqSvc   pullrequest.Interface
	PipelineSvc  pipeline.Interface
	ExecutionSvc execution.Interface
	StageSvc     stage.Interface
	StepSvc      step.Interface
	Scheduler    scheduler.Scheduler
	Logz         livelog.LogStream
}

func New(
	config *config.Config,
	executionSvc execution.Interface,
	pipelineSvc pipeline.Interface,
	urlProvider urlprovider.Provider,
	sseStreamer sse.Streamer,
	scheduler scheduler.Scheduler,
	stageSvc stage.Interface,
	stepSvc step.Interface,
	reporter events.Reporter,
	prReporter *pullreqevents.Reporter,
	pullReqSvc pullrequest.Interface,
	logz livelog.LogStream,
) *Manager {
	return &Manager{
		Config:       config,
		ExecutionSvc: executionSvc,
		PipelineSvc:  pipelineSvc,
		urlProvider:  urlProvider,
		SSEStreamer:  sseStreamer,
		Scheduler:    scheduler,
		StageSvc:     stageSvc,
		StepSvc:      stepSvc,
		reporter:     reporter,
		prReporter:   prReporter,
		pullReqSvc:   pullReqSvc,
		Logz:         logz,
	}
}

// Request requests the next available build stage for execution.
func (m *Manager) Request(ctx context.Context, args *Request) (*stageModel.Stage, error) {
	log := logger.GetLogger(ctx).With(
		"token", args.Token,
		"type", args.Type,
		"os", args.OS,
		"arch", args.Arch,
	)
	log.Debug("manager: request queue item")

	stage, err := m.Scheduler.Request(ctx, scheduler.Filter{
		Token:    args.Token,
		Type:     args.Type,
		OS:       args.OS,
		Arch:     args.Arch,
		Labels:   args.Labels,
		Metadata: args.Metadata,
	})
	if err != nil && ctx.Err() != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return nil, ctx.Err()
		}
		log.Debugw("manager: context canceled", "error", err)
		return nil, err
	}
	if err != nil {
		log.Warnw("manager: request queue item error", "error", err)
		return nil, err
	}
	return stage, nil
}

// Accept accepts the build stage for execution. It is possible for multiple
// agents to pull the same stage from the queue.
func (m *Manager) Accept(_ context.Context, id int64, machine string) (*stageModel.Stage, error) {
	log := logger.GetLogger(noContext).With(
		"stage-id", id,
		"machine", machine,
	)
	log.Debug("manager: accept stage")

	stage, err := m.StageSvc.GetByID(noContext, uint(id))
	if err != nil {
		log.Warnw("manager: cannot find stage", "error", err)
		return nil, err
	}
	if stage.Host != "" {
		log.Debug("manager: stage already assigned. abort.")
		return nil, fmt.Errorf("stage already assigned, abort")
	}

	//TODO: 成功解析前先对machine不赋值
	stage.Host = machine
	stage.Status = constant.CIStatusPending
	if err := m.StageSvc.Update(noContext, stage); err != nil {
		log.Debugw("manager: cannot update stage", "error", err)
	} else {
		log.Info("manager: stage accepted")
	}
	return stage, err
}

// Write writes a line to the build logs.
func (m *Manager) Write(ctx context.Context, step int64, line *livelog.Line) error {
	err := m.Logz.Write(ctx, step, line)
	if err != nil {
		logger.Warnw("manager: cannot write to log stream",
			"step-id", step,
			"error", err,
		)
		return err
	}
	return nil
}

// StepStartTime returns the time the step started, used as the base for
// converting the relative log timestamps reported by runners.
// 取不到时（step 不存在、开始时间未知或查询失败）返回零值时间，
// 由 ResolveLineTimestamp 退回当前时间，不影响日志写入本身。
func (m *Manager) StepStartTime(ctx context.Context, stepID int64) time.Time {
	step, err := m.StepSvc.GetByID(ctx, uint(stepID))
	if err != nil {
		logger.Warnw("manager: cannot resolve step start time",
			"step-id", stepID,
			"error", err,
		)
		return time.Time{}
	}
	if step.Started == nil {
		return time.Time{}
	}
	return time.Unix(toUnix(step.Started), 0)
}

// UploadLogs uploads the full logs.
func (m *Manager) UploadLogs(ctx context.Context, stepID int64, r io.Reader) error {
	step, err := m.StepSvc.GetByID(ctx, uint(stepID))
	if err != nil {
		logger.Errorw("manager: cannot upload complete logs",
			"error", err,
			"step-id", stepID,
		)
		return err
	}
	stage, err := m.StageSvc.GetByID(ctx, step.StageID)
	if err != nil {
		logger.Errorw("manager: cannot upload complete logs",
			"error", err,
			"step-id", stepID,
		)
		return err
	}
	execution, err := m.ExecutionSvc.GetByID(ctx, stage.ExecutionID)
	if err != nil {
		logger.Errorw("manager: cannot upload complete logs",
			"error", err,
			"step-id", stepID,
		)
		return err
	}
	if err := m.PipelineSvc.CreateLogs(ctx, &pipelineModel.S3Logs{
		PipelineID:  fmt.Sprint(execution.PipelineID),
		ExecutionID: fmt.Sprint(stage.ExecutionID),
		StageID:     fmt.Sprint(step.StageID),
		StepID:      fmt.Sprint(stepID),
	}, r); err != nil {
		logger.Errorw("manager: cannot upload complete logs",
			"error", err,
			"step-id", stepID,
		)
		return err
	}
	return nil
}

// Details provides details about the stage.
func (m *Manager) Details(ctx context.Context, stageID int64) (*ExecutionContext, error) {
	log := logger.GetLogger(ctx).With(
		"stage-id", stageID,
	)
	log.Debug("manager: fetching stage details")

	stage, err := m.StageSvc.GetByIDWithSteps(noContext, uint(stageID))
	if err != nil {
		log.Warnw("manager: cannot find stage", "error", err)
		return nil, err
	}

	execution, err := m.ExecutionSvc.GetByID(noContext, stage.ExecutionID)
	if err != nil {
		log.Warnw("manager: cannot find build", "error", err)
		return nil, err
	}
	stages, err := m.StageSvc.ListWithSteps(noContext, stage.ExecutionID)
	if err != nil {
		log.Warnw("manager: cannot list stages", "error", err)
		return nil, err
	}
	executionDetails := executionModel.ExecutionWithStages{
		Execution: *execution,
		Stages:    stages,
	}
	log = log.With(
		"build", int64(execution.ID),
	)

	// TODO: Currently we fetch all the secrets from the same space.
	// This logic can be updated when needed.
	// secrets, err := m.Secrets.ListAll(noContext, repo.ParentID)
	// if err != nil {
	// 	log.Warn().Err(err).Msg("manager: cannot list secrets")
	// 	return nil, err
	// }

	pipeline, err := m.PipelineSvc.GetPipelineByID(noContext, int64(execution.PipelineID))
	if err != nil {
		log.Warnw("manager: cannot find pipeline", "error", err)
		return nil, err
	}
	// Fetch contents of YAML from the execution ref at the pipeline config path.
	file, err := m.PipelineSvc.GetPipelineContent(noContext, pipeline.ID)
	if err != nil {
		log.Warnw("manager: cannot fetch file", "error", err)
		return nil, err
	}

	netrc, err := m.createNetrc(ctx)
	if err != nil {
		log.Warnw("manager: failed to create netrc", "error", err)
		return nil, err
	}

	return &ExecutionContext{
		Execution: &executionDetails,
		Stage:     stage,
		Config:    file,
		Netrc:     netrc,
		// Secrets:   secrets,
	}, nil
}

// Before signals the build step is about to start.
func (m *Manager) BeforeStep(_ context.Context, step *stepModel.Step) error {
	log := logger.GetLogger(noContext).With(
		"step.status", string(step.Status),
		"step.name", step.Name,
		"step.id", int64(step.ID),
	)

	log.Debug("manager: updating step status")

	err := m.Logz.Create(noContext, int64(step.ID))
	if err != nil {
		log.Warnw("manager: cannot create log stream", "error", err)
		return err
	}
	updater := &updater{
		ExecutionSvc: m.ExecutionSvc,
		SSEStreamer:  m.SSEStreamer,
		StepSvc:      m.StepSvc,
		StageSvc:     m.StageSvc,
		PipelineSvc:  m.PipelineSvc,
	}
	return updater.do(noContext, step)
}

// After signals the build step is complete.
func (m *Manager) AfterStep(_ context.Context, step *stepModel.Step) error {
	log := logger.GetLogger(noContext).With(
		"step.status", string(step.Status),
		"step.name", step.Name,
		"step.id", int64(step.ID),
	)
	log.Debug("manager: updating step status")

	var retErr error
	updater := &updater{
		ExecutionSvc: m.ExecutionSvc,
		SSEStreamer:  m.SSEStreamer,
		StepSvc:      m.StepSvc,
		StageSvc:     m.StageSvc,
		PipelineSvc:  m.PipelineSvc,
	}

	if err := updater.do(noContext, step); err != nil {
		retErr = err
		log.Warnw("manager: cannot update step", "error", err)
	}

	if err := m.Logz.Delete(noContext, int64(step.ID)); err != nil && !errors.Is(err, livelog.ErrStreamNotFound) {
		log.Warnw("manager: cannot teardown log stream", "error", err)
	}
	return retErr
}

// BeforeAll signals the build stage is about to start.
func (m *Manager) BeforeStage(_ context.Context, stage *stageModel.StageWithSteps) error {
	s := &setup{
		ExecutionSvc: m.ExecutionSvc,
		PipelineSvc:  m.PipelineSvc,
		SSEStreamer:  m.SSEStreamer,
		StepSvc:      m.StepSvc,
		StageSvc:     m.StageSvc,
	}

	return s.do(noContext, stage)
}

// AfterAll signals the build stage is complete.
func (m *Manager) AfterStage(_ context.Context, stage *stageModel.StageWithSteps) error {
	t := &teardown{
		ExecutionSvc: m.ExecutionSvc,
		PipelineSvc:  m.PipelineSvc,
		SSEStreamer:  m.SSEStreamer,
		Logz:         m.Logz,
		Scheduler:    m.Scheduler,
		StepSvc:      m.StepSvc,
		StageSvc:     m.StageSvc,
		Reporter:     m.reporter,
		PRReporter:   m.prReporter,
		PullReqSvc:   m.pullReqSvc,
	}
	return t.do(noContext, stage)
}

// Watch watches for build cancellation requests.
func (m *Manager) Watch(ctx context.Context, executionID int64) (bool, error) {
	ok, err := m.Scheduler.Cancelled(ctx, executionID)
	// we expect a context cancel error here which
	// indicates a polling timeout. The subscribing
	// client should look for the context cancel error
	// and resume polling.
	if err != nil {
		return ok, err
	}

	// // TODO: we should be able to return
	// // immediately if Cancelled returns true. This requires
	// // some more testing but would avoid the extra database
	// // call.
	// if ok {
	// 	return ok, err
	// }

	// if no error is returned we should check
	// the database to see if the build is complete. If
	// complete, return true.
	execution, err := m.ExecutionSvc.GetByID(ctx, uint(executionID))
	if err != nil {
		log := logger.GetLogger(ctx).With(
			"execution.id", executionID,
		)
		log.Warn("manager: cannot find build")
		return ok, fmt.Errorf("could not find build for cancellation: %w", err)
	}
	return execution.Status.IsDone(), nil
}

func (m *Manager) createNetrc(ctx context.Context) (*Netrc, error) {
	return &Netrc{
		Machine:  m.urlProvider.GetGITHostname(ctx),
		Login:    m.Config.PipelineAuth.UserName,
		Password: m.Config.PipelineAuth.Password,
	}, nil
}
