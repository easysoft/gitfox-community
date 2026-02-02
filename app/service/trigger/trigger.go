package trigger

import (
	"context"
	"errors"
	"fmt"
	"time"

	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	pipelineModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	pipelineSvc "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/stream"

	"github.com/hashicorp/go-multierror"
)

const (
	eventsReaderGroupName = "gitfox:trigger"
)

type Config struct {
	EventReaderName string
	Concurrency     int
	MaxRetries      int
}

func (c *Config) Prepare() error {
	if c == nil {
		return errors.New("config is required")
	}
	if c.EventReaderName == "" {
		return errors.New("config.EventReaderName is required")
	}
	if c.Concurrency < 1 {
		return errors.New("config.Concurrency has to be a positive number")
	}
	if c.MaxRetries < 0 {
		return errors.New("config.MaxRetries can't be negative")
	}

	return nil
}

type Service struct {
	pipelineSvc pipelineSvc.Interface
	triggerSvc  triggerer.Triggerer
}

func New(
	ctx context.Context,
	config Config,
	pipelineSvc pipelineSvc.Interface,
	triggerSvc triggerer.Triggerer,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullreqEvReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
) (*Service, error) {
	if err := config.Prepare(); err != nil {
		return nil, fmt.Errorf("provided trigger service config is invalid: %w", err)
	}

	service := &Service{
		pipelineSvc: pipelineSvc,
		triggerSvc:  triggerSvc,
	}

	_, err := gitReaderFactory.Launch(ctx, eventsReaderGroupName, config.EventReaderName,
		func(r *gitevents.Reader) error {
			const idleTimeout = 1 * time.Minute
			r.Configure(
				stream.WithConcurrency(config.Concurrency),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					stream.WithMaxRetries(config.MaxRetries),
				))

			_ = r.RegisterBranchCreated(service.handleEventBranchCreated)
			_ = r.RegisterBranchUpdated(service.handleEventBranchUpdated)

			_ = r.RegisterTagCreated(service.handleEventTagCreated)
			_ = r.RegisterTagUpdated(service.handleEventTagUpdated)

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to launch git events reader: %w", err)
	}

	_, err = pullreqEvReaderFactory.Launch(ctx, eventsReaderGroupName, config.EventReaderName,
		func(r *pullreqevents.Reader) error {
			const idleTimeout = 1 * time.Minute
			r.Configure(
				stream.WithConcurrency(config.Concurrency),
				stream.WithHandlerOptions(
					stream.WithIdleTimeout(idleTimeout),
					// retries not needed for builds which failed to trigger, can be adjusted when needed
					stream.WithMaxRetries(0),
				))

			_ = r.RegisterCreated(service.handleEventPullReqCreated)
			_ = r.RegisterBranchUpdated(service.handleEventPullReqBranchUpdated)
			_ = r.RegisterReopened(service.handleEventPullReqReopened)
			_ = r.RegisterClosed(service.handleEventPullReqClosed)
			_ = r.RegisterMerged(service.handleEventPullReqMerged)

			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("failed to launch pr events reader: %w", err)
	}

	return service, nil
}

func (s *Service) trigger(ctx context.Context, repoID int64, hook *triggerer.Hook) error {
	log := logger.GetLogger(ctx)
	log.Infof("[pipeline-trigger] trigger called: repoID=%d, action=%s, ref=%s", repoID, hook.Action, hook.Ref)

	// Get all enabled triggers for a repo or space.
	filter := pipelineModel.TriggerFilter{
		RepoID:  repoID,
		Trigger: []string{string(hook.Action)},
	}
	triggers, err := s.pipelineSvc.GetTriggers(ctx, &filter)
	if err != nil {
		return fmt.Errorf("failed to list all enabled triggers: %w", err)
	}

	log.Infof("[pipeline-trigger] found %d triggers for repoID=%d, action=%s", len(triggers), repoID, hook.Action)

	var errs error
	for _, t := range triggers {
		params, err := s.pipelineSvc.GetPipelineVariables(ctx, t.PipelineID)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to get pipeline variables %d: %w", t.PipelineID, err))
			continue
		}
		// 合并 params 到 hook.Params
		if hook.Params == nil {
			hook.Params = make(map[string]string)
		}
		for k, v := range params {
			hook.Params[k] = v
		}
		_, err = s.triggerSvc.Trigger(ctx, uint(t.PipelineID), hook)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}
