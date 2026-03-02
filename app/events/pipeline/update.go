package events

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"

	"github.com/rs/zerolog/log"
)

const UpdatedEvent events.EventType = "updated"

type UpdatedPayload struct {
	PipelineID int64 `json:"pipelineID"`
}

func (r *Reporter) Updated(ctx context.Context, payload *UpdatedPayload) {
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, UpdatedEvent, payload)
	if err != nil {
		log.Ctx(ctx).Err(err).Msgf("failed to send pipeline updated event")
		return
	}

	log.Ctx(ctx).Debug().Msgf("reported pipeline update event with id '%s'", eventID)
}

func (r *Reader) RegisterUpdated(fn events.HandlerFunc[*UpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, UpdatedEvent, fn, opts...)
}
