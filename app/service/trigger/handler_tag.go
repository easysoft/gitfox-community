package trigger

import (
	"context"
	"strings"

	"github.com/easysoft/gitfox-community/app/constant"
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	"github.com/easysoft/gitfox-community/app/pipeline/triggerer"
	"github.com/easysoft/gitfox-community/pkg/events"
)

func (s *Service) handleEventTagCreated(ctx context.Context,
	event *events.Event[*gitevents.TagCreatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger:      event.Payload.Account,
		Action:       constant.TriggerActionTagCreated,
		Ref:          event.Payload.Ref,
		Match:        strings.TrimPrefix(event.Payload.Ref, "refs/tags/"),
		Commit:       event.Payload.SHA,
		CommitBefore: event.Payload.SHA,
	}
	return s.trigger(ctx, event.Payload.RepoID, 0, hook)
}

func (s *Service) handleEventTagUpdated(ctx context.Context,
	event *events.Event[*gitevents.TagUpdatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger: event.Payload.Account,
		Action:  constant.TriggerActionTagUpdated,
		Ref:     event.Payload.Ref,
		Match:   strings.TrimPrefix(event.Payload.Ref, "refs/tags/"),
		Commit:  event.Payload.NewSHA,
	}
	return s.trigger(ctx, event.Payload.RepoID, 0, hook)
}
