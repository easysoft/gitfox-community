package events

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"

	"github.com/rs/zerolog/log"
)

const BranchUpdatedEvent events.EventType = "branch-updated"

type BranchUpdatedPayload struct {
	Base
	OldSHA          string `json:"oldSHA"`
	NewSHA          string `json:"newSHA"`
	OldMergeBaseSHA string `json:"oldMergeBaseSHA"`
	NewMergeBaseSHA string `json:"newMergeBaseSHA"`
	Forced          bool   `json:"forced"`
}

func (r *Reporter) BranchUpdated(ctx context.Context, payload *BranchUpdatedPayload) {
	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, BranchUpdatedEvent, payload)
	if err != nil {
		log.Ctx(ctx).Err(err).Msgf("failed to send pull request branch updated event")
		return
	}

	log.Ctx(ctx).Debug().Msgf("reported pull request branch updated event with id '%s'", eventID)
}

func (r *Reader) RegisterBranchUpdated(fn events.HandlerFunc[*BranchUpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, BranchUpdatedEvent, fn, opts...)
}
