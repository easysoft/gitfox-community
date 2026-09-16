package trigger

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"slices"
	"time"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	pullreqevents "github.com/easysoft/gitfox-community/app/events/pullreq"
	pipelineModel "github.com/easysoft/gitfox-community/app/model/pipeline"
	scanModel "github.com/easysoft/gitfox-community/app/model/scan"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	pipelineSvc "github.com/easysoft/gitfox-community/app/service/pipeline"
	scanSvc "github.com/easysoft/gitfox-community/app/service/scan"
	repoStore "github.com/easysoft/gitfox-community/app/store/repo"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/stream"
	"github.com/easysoft/gitfox-community/pkg/utils/url"

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
	pipelineSvc    pipelineSvc.Interface
	triggerSvc     triggerer.Triggerer
	repoStore      repoStore.Interface
	urlProvider    url.Provider
	scanTriggerSvc scanSvc.IScanTrigger
	planSvc        scanSvc.IScanPlan
	taskSvc        scanSvc.IScanTask
	reviewSvc      scanSvc.IAIReview
}

func New(
	ctx context.Context,
	config Config,
	pipelineSvc pipelineSvc.Interface,
	triggerSvc triggerer.Triggerer,
	repoStore repoStore.Interface,
	urlProvider url.Provider,
	scanTriggerSvc scanSvc.IScanTrigger,
	planSvc scanSvc.IScanPlan,
	taskSvc scanSvc.IScanTask,
	reviewSvc scanSvc.IAIReview,
	gitReaderFactory *events.ReaderFactory[*gitevents.Reader],
	pullreqEvReaderFactory *events.ReaderFactory[*pullreqevents.Reader],
) (*Service, error) {
	if err := config.Prepare(); err != nil {
		return nil, fmt.Errorf("provided trigger service config is invalid: %w", err)
	}

	service := &Service{
		pipelineSvc:    pipelineSvc,
		triggerSvc:     triggerSvc,
		repoStore:      repoStore,
		urlProvider:    urlProvider,
		scanTriggerSvc: scanTriggerSvc,
		planSvc:        planSvc,
		taskSvc:        taskSvc,
		reviewSvc:      reviewSvc,
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

func (s *Service) trigger(ctx context.Context, repoID, ppmID int64, hook *triggerer.Hook) error {
	var errs error
	if err := s.triggerPipeline(ctx, repoID, hook); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := s.triggerScanPipeline(ctx, repoID, hook); err != nil {
		errs = multierror.Append(errs, err)
	}

	reviewActions := []constant.TriggerAction{
		constant.TriggerActionPullReqCreated,
		constant.TriggerActionPullReqBranchUpdated,
		constant.TriggerActionPullReqReopened,
	}
	if slices.Contains(reviewActions, hook.Action) {
		if err := s.triggerReviewPipeline(ctx, repoID, ppmID, hook); err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (s *Service) triggerPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) error {
	log := logger.GetLogger(ctx)
	log.Infof("[pipeline-trigger] trigger called: repoID=%d, action=%s, ref=%s", repoID, hook.Action, hook.Ref)

	// 根据 repoID 查询 repo 的 gitURL
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return fmt.Errorf("failed to get repo by ID %d: %w", repoID, err)
	}
	gitURL := s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	log.Infof("[pipeline-trigger] repo gitURL=%s", gitURL)

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
		if hook.Action == constant.TriggerActionBranchUpdated {
			if !s.isCommentMatch(hook.CommitMatch, t.Comment) {
				continue
			}
		}
		pipelineInfo, err := s.pipelineSvc.GetPipelineByID(ctx, t.PipelineID)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to get pipeline info %d: %w", t.PipelineID, err))
			continue
		}
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
		hook.Params["gitURL"] = gitURL
		hook.Params["gitRef"] = hook.Ref
		hook.Parent = pipelineInfo.SpaceID
		_, err = s.triggerSvc.Trigger(ctx, uint(t.PipelineID), hook)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
	}
	return errs
}

func (s *Service) triggerScanPipeline(ctx context.Context, repoID int64, hook *triggerer.Hook) error {
	log := logger.GetLogger(ctx)
	log.Infof("[pipeline-trigger-scan] trigger called: repoID=%d, action=%s, ref=%s", repoID, hook.Action, hook.Ref)

	// 根据 repoID 查询 repo 的 gitURL
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return fmt.Errorf("failed to get repo by ID %d: %w", repoID, err)
	}
	gitURL := s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	log.Infof("[pipeline-trigger-scan] repo gitURL=%s", gitURL)
	repo.GitURL = gitURL

	// 查询扫描相关的 trigger
	triggers, err := s.scanTriggerSvc.FindByRepoIDAndActions(ctx, uint(repoID), []constant.TriggerAction{hook.Action})
	if err != nil {
		return fmt.Errorf("failed to list all enabled triggers: %w", err)
	}
	log.Infof("[pipeline-trigger-scan] found %d raw triggers for repoID=%d, action=%s", len(triggers), repoID, hook.Action)
	// 根据 trigger 关联的 planID 查询扫描计划，匹配分支后获取有效的触发器
	var errs error
	for _, t := range triggers {
		plan, err := s.planSvc.GetByID(ctx, repo.ID, uint(t.PlanID))
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to get scan plan info %d: %w", t.PlanID, err))
			continue
		}
		if !s.isBranchMatch(plan.Branches, hook.Match) {
			log.Infof("[pipeline-trigger-scan] skip trigger %d for plan %d due to branch filter not match: filter=%s, ref=%s", t.ID, t.PlanID, plan.Branches, hook.Ref)
			continue
		}

		scanTrigger := &scanModel.ScanTaskTrigger{
			ID:   t.ID,
			Name: t.Name,
			Type: constant.ScanPlanTriggerTypeAction,
		}

		if hook.Params == nil {
			hook.Params = make(map[string]string)
		}
		hook.Parent = int64(repo.SpaceID)
		if _, err = s.taskSvc.Create(ctx, &scanSvc.ScanTaskInput{Plan: plan, Repo: repo, Trigger: scanTrigger, Hook: hook}); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to create scan task for plan %d: %w", t.PlanID, err))
		}
	}
	return errs
}

func (s *Service) triggerReviewPipeline(ctx context.Context, repoID, ppmID int64, hook *triggerer.Hook) error {

	log := logger.GetLogger(ctx)
	log.Infof("[pipeline-trigger-review] trigger called: repoID=%d, action=%s, ref=%s", repoID, hook.Action, hook.Ref)

	// 根据 repoID 查询 repo 的 gitURL
	repo, err := s.repoStore.GetByID(ctx, uint(repoID))
	if err != nil {
		return fmt.Errorf("failed to get repo by ID %d: %w", repoID, err)
	}
	gitURL := s.urlProvider.GenerateGITCloneURL(ctx, repo.Path)
	log.Infof("[pipeline-trigger-review] repo gitURL=%s", gitURL)
	repo.GitURL = gitURL

	var errs error
	if hook.Params == nil {
		hook.Params = make(map[string]string)
	}

	// 创建 or 获取 review pipeline
	hook.Params["gitURL"] = gitURL
	hook.Params["gitRef"] = hook.Ref
	hook.Parent = int64(repo.SpaceID)

	if _, err := s.reviewSvc.TriggerCodeReview(ctx, &scanSvc.CodeReviewInput{
		PullReqID: ppmID,
		Repo:      repo,
		Hook:      hook,
	}); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("failed to get review pipeline for repo %d: %w", repoID, err))
		return errs
	}

	return errs
}

// isBranchMatch checks whether ref matches any of the include patterns in branches.
// Returns true if at least one pattern matches, false if none match.
// If branches is nil or has no include patterns, it returns true (no filter).
func (s *Service) isBranchMatch(branches *scanModel.Pattern, ref string) bool {
	if branches == nil || len(branches.Include) == 0 {
		return true
	}
	for _, pattern := range branches.Include {
		if ok, _ := doublestar.Match(pattern, ref); ok {
			return true
		}
	}
	return false
}

func (s *Service) isCommentMatch(commitMessage, comment string) bool {
	if comment == "" {
		return true
	}
	commentList := strings.Split(comment, ",")
	for _, c := range commentList {
		if strings.Contains(commitMessage, c) {
			return true
		}
	}
	return false
}
