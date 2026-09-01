package events

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

const CreatedEvent events.EventType = "created"

type CreatedPayload struct {
	PipelineID int64 `json:"pipelineID"`
}

func (r *Reporter) Created(ctx context.Context, payload *CreatedPayload) {
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, CreatedEvent, payload)
	if err != nil {
		logger.GetLogger(ctx).Errorw("failed to send pipeline created event", "error", err)
		return
	}

	logger.GetLogger(ctx).Debugf("reported pipeline created event with id '%s'", eventID)
}

func (r *Reader) RegisterCreated(fn events.HandlerFunc[*CreatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, CreatedEvent, fn, opts...)
}
