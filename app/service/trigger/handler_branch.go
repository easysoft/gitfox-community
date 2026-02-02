package trigger

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/constant"
	gitevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/git"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/pipeline/triggerer"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

func (s *Service) handleEventBranchCreated(ctx context.Context, event *events.Event[*gitevents.BranchCreatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger: event.Payload.Account,
		Action:  constant.TriggerActionBranchCreated,
		Ref:     event.Payload.Ref,
	}
	return s.trigger(ctx, event.Payload.RepoID, hook)
}

func (s *Service) handleEventBranchUpdated(ctx context.Context,
	event *events.Event[*gitevents.BranchUpdatedPayload]) error {
	hook := &triggerer.Hook{
		Trigger: event.Payload.Account,
		Action:  constant.TriggerActionBranchUpdated,
		Ref:     event.Payload.Ref,
	}
	return s.trigger(ctx, event.Payload.RepoID, hook)
}
