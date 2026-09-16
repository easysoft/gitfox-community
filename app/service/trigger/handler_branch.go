package trigger

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	"github.com/easysoft/gitfox-community/pkg/events"
)

func (s *Service) handleEventBranchCreated(ctx context.Context, event *events.Event[*gitevents.BranchCreatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionBranchCreated,
		Ref:          event.Payload.Ref,
		Match:        strings.TrimPrefix(event.Payload.Ref, "refs/heads/"),
		Commit:       event.Payload.SHA,
		CommitBefore: event.Payload.SHA,
	}
	return s.trigger(ctx, event.Payload.RepoID, 0, hook)
}

func (s *Service) handleEventBranchUpdated(ctx context.Context,
	event *events.Event[*gitevents.BranchUpdatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionBranchUpdated,
		Ref:          event.Payload.Ref,
		Match:        strings.TrimPrefix(event.Payload.Ref, "refs/heads/"),
		Commit:       event.Payload.NewSHA,
		CommitBefore: event.Payload.OldSHA,
		CommitMatch:  event.Payload.CommitMessage,
	}
	return s.trigger(ctx, event.Payload.RepoID, 0, hook)
}
