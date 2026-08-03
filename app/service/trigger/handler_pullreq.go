package trigger

import (
	"context"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

func (s *Service) handleEventPullReqCreated(ctx context.Context, event *events.Event[*pullreqevents.CreatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionPullReqCreated,
		Ref:          fmt.Sprintf("refs/pullreq/%d/head", event.Payload.PullReqID),
		Match:        event.Payload.SourceBranch,
		Commit:       event.Payload.SourceSHA,
		CommitBefore: event.Payload.MergeBaseSHA,
	}
	return s.trigger(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID, hook)
}

func (s *Service) handleEventPullReqReopened(ctx context.Context, event *events.Event[*pullreqevents.ReopenedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionPullReqReopened,
		Ref:          fmt.Sprintf("refs/pullreq/%d/head", event.Payload.PullReqID),
		Match:        event.Payload.SourceBranch,
		Commit:       event.Payload.SourceSHA,
		CommitBefore: event.Payload.MergeBaseSHA,
	}
	return s.trigger(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID, hook)
}

func (s *Service) handleEventPullReqBranchUpdated(ctx context.Context, event *events.Event[*pullreqevents.BranchUpdatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionPullReqBranchUpdated,
		Ref:          fmt.Sprintf("refs/pullreq/%d/head", event.Payload.PullReqID),
		Match:        event.Payload.SourceBranch,
		Commit:       event.Payload.NewSHA,
		CommitBefore: event.Payload.OldSHA,
	}
	return s.trigger(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID, hook)
}

func (s *Service) handleEventPullReqClosed(ctx context.Context, event *events.Event[*pullreqevents.ClosedPayload]) error {
	hook := &triggerer.Hook{
		Trigger: event.Payload.Account,
		Action:  constant.TriggerActionPullReqClosed,
		Ref:     fmt.Sprintf("refs/pullreq/%d/head", event.Payload.PullReqID),
		Match:   event.Payload.SourceBranch,
		Commit:  event.Payload.SourceSHA,
	}
	return s.trigger(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID, hook)
}

func (s *Service) handleEventPullReqMerged(
	ctx context.Context,
	event *events.Event[*pullreqevents.MergedPayload],
) error {
	hook := &triggerer.Hook{
		Trigger: event.Payload.Account,
		Action:  constant.TriggerActionPullReqMerged,
		Ref:     fmt.Sprintf("refs/heads/%s", event.Payload.TargetBranch),
		Match:   event.Payload.TargetBranch,
		Commit:  event.Payload.MergeSHA,
	}
	return s.trigger(ctx, event.Payload.SourceRepoID, event.Payload.PullReqID, hook)
}
