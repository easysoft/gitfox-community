package webhook

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	webhookModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/webhook"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
	"go.uber.org/multierr"
)

// triggerForEventWithRepo triggers all webhooks for the given repo and triggerType
// using the eventID to generate a deterministic triggerID and using the output of bodyFn as payload.
// The method tries to find the repository and principal and provides both to the bodyFn to generate the body.
// NOTE: technically we could avoid this call if we send the data via the event (though then events will get big).
func (s *Service) triggerForEventWithRepo(
	ctx context.Context,
	triggerType constant.WebhookTrigger,
	eventID string,
	principal string,
	repoID int64,
	createBodyFn func(string, *repoModel.Repo) (any, error),
) error {
	repo, err := s.repoSvc.GetByID(ctx, nil, uint(repoID))
	if err != nil {
		return err
	}

	// create body
	body, err := createBodyFn(principal, repo)
	if err != nil {
		return fmt.Errorf("body creation function failed: %w", err)
	}

	parents, err := s.getParentInfoRepo(repo.ID)
	if err != nil {
		return fmt.Errorf("failed to get webhook parent info for parents: %w", err)
	}

	return s.triggerForEvent(ctx, eventID, parents, triggerType, &webhookModel.ReqForList{}, body)
}

// triggerForEventWithPullReq triggers all webhooks for the given repo and triggerType
// using the eventID to generate a deterministic triggerID and using the output of bodyFn as payload.
// The method tries to find the pullreq, principal, target repo, and source repo
// and provides all to the bodyFn to generate the body.
// NOTE: technically we could avoid this call if we send the data via the event (though then events will get big).
func (s *Service) triggerForEventWithPullReq(ctx context.Context,
	triggerType constant.WebhookTrigger, eventID string, principal string, targetRepoID, prID int64,
	createBodyFn func(principal string, pr *repoModel.PullRequest, targetRepo, sourceRepo *repoModel.Repo) (any, error)) error {

	pr, err := s.pullreqStore.GetByID(ctx, uint(targetRepoID), uint(prID))
	if err != nil {
		return err
	}

	targetRepo, err := s.repoSvc.GetByID(ctx, nil, pr.TargetRepoID)
	if err != nil {
		return fmt.Errorf("failed to get pr target repo: %w", err)
	}

	sourceRepo := targetRepo
	if pr.SourceRepoID != pr.TargetRepoID {
		sourceRepo, err = s.repoSvc.GetByID(ctx, nil, pr.SourceRepoID)
		if err != nil {
			return fmt.Errorf("failed to get pr source repo: %w", err)
		}
	}

	// create body
	body, err := createBodyFn(principal, pr, targetRepo, sourceRepo)
	if err != nil {
		return fmt.Errorf("body creation function failed: %w", err)
	}

	parents, err := s.getParentInfoRepo(targetRepo.ID)
	if err != nil {
		return fmt.Errorf("failed to get webhook parent info: %w", err)
	}

	return s.triggerForEvent(ctx, eventID, parents, triggerType, &webhookModel.ReqForList{}, body)
}

func (s *Service) getParentInfoRepo(repoID uint) ([]webhookModel.ParentInfo, error) {
	var parents []webhookModel.ParentInfo

	parents = append(parents, webhookModel.ParentInfo{
		ID:   repoID,
		Type: "repo",
	})

	return parents, nil
}

// triggerForEvent triggers all webhooks for the given parentType/ID and triggerType
// using the eventID to generate a deterministic triggerID and sending the provided body as payload.
func (s *Service) triggerForEvent(
	ctx context.Context,
	eventID string,
	parents []webhookModel.ParentInfo,
	triggerType constant.WebhookTrigger,
	filter *webhookModel.ReqForList,
	body any,
) error {
	log := logger.GetLogger(ctx)
	triggerID := eventID

	results, err := s.triggerWebhooksFor(ctx, parents, triggerID, triggerType, filter, body)

	// return all errors and force the event to be reprocessed (it's not webhook execution specific!)
	if err != nil {
		return fmt.Errorf(
			"failed to trigger %s (id: '%s') for webhooks %#v: %w",
			triggerType, triggerID, parents, err,
		)
	}

	// go through all events and figure out if we need to retry the event.
	// Combine all errors into a single error to log (to reduce number of logs)
	retryRequired := false
	var errs error
	for _, result := range results {
		if result.Skipped() {
			continue
		}

		// combine errors of non-successful executions
		if result.Execution.Result != constant.WebhookExecutionResultSuccess {
			errs = multierr.Append(errs,
				fmt.Errorf("execution %d of webhook %d resulted in %s: %w",
					result.Execution.ID, result.Webhook.ID, result.Execution.Result, result.Err))
		}

		if result.Execution.Result == constant.WebhookExecutionResultRetriableError {
			retryRequired = true
		}
	}

	// in case there was at least one error, log error details in single log to reduce log flooding
	if errs != nil {
		log.Warnf("webhook execution for %#v had errors", parents)
	}

	// in case at least one webhook has to be retried, return an error to the event framework to have it reprocessed
	if retryRequired {
		return fmt.Errorf("at least one webhook execution resulted in a retry for %#v", parents)
	}

	return nil
}
