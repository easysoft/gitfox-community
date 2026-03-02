package events

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"

	"github.com/rs/zerolog/log"
)

const CreatedEvent events.EventType = "created"

type CreatedPayload struct {
	PipelineID int64 `json:"pipelineID"`
}

func (r *Reporter) Created(ctx context.Context, payload *CreatedPayload) {
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, CreatedEvent, payload)
	if err != nil {
		log.Ctx(ctx).Err(err).Msgf("failed to send pipeline created event")
		return
	}

	log.Ctx(ctx).Debug().Msgf("reported pipeline created event with id '%s'", eventID)
}

func (r *Reader) RegisterCreated(fn events.HandlerFunc[*CreatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, CreatedEvent, fn, opts...)
}
