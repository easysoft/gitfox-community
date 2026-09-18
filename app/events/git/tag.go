package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const TagCreatedEvent events.EventType = "tag-created"

type TagCreatedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	SHA     string `json:"sha"`
}

func (r *Reporter) TagCreated(ctx context.Context, payload *TagCreatedPayload) {
	logger := logger.GetLogger(ctx)
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, TagCreatedEvent, payload)
	if err != nil {
		logger.Error("failed to send tag created event")
		return
	}

	logger.Infof("reported tag created event with id '%s'", eventID)
}

func (r *Reader) RegisterTagCreated(fn events.HandlerFunc[*TagCreatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, TagCreatedEvent, fn, opts...)
}

const TagUpdatedEvent events.EventType = "tag-updated"

type TagUpdatedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	OldSHA  string `json:"oldSHA"`
	NewSHA  string `json:"newSHA"`
	Forced  bool   `json:"forced"`
}

func (r *Reporter) TagUpdated(ctx context.Context, payload *TagUpdatedPayload) {
	logger := logger.GetLogger(ctx)
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, TagUpdatedEvent, payload)
	if err != nil {
		logger.Error("failed to send tag updated event")
		return
	}

	logger.Infof("reported tag updated event with id '%s'", eventID)
}

func (r *Reader) RegisterTagUpdated(fn events.HandlerFunc[*TagUpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, TagUpdatedEvent, fn, opts...)
}

const TagDeletedEvent events.EventType = "tag-deleted"

type TagDeletedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	SHA     string `json:"sha"`
}

func (r *Reporter) TagDeleted(ctx context.Context, payload *TagDeletedPayload) {
	logger := logger.GetLogger(ctx)
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, TagDeletedEvent, payload)
	if err != nil {
		logger.Error("failed to send tag deleted event")
		return
	}

	logger.Infof("reported tag deleted event with id '%s'", eventID)
}

func (r *Reader) RegisterTagDeleted(fn events.HandlerFunc[*TagDeletedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, TagDeletedEvent, fn, opts...)
}
