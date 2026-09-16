package events

import (
	"context"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const CreatedEvent events.EventType = "created"

type CreatedPayload struct {
	Base
	SourceBranch string `json:"sourceBranch"`
	TargetBranch string `json:"targetBranch"`
	SourceSHA    string `json:"sourceSHA"`
	MergeBaseSHA string `json:"mergeBaseSHA"`
}

func (r *Reporter) Created(ctx context.Context, payload *CreatedPayload) {

	logger := logger.GetLogger(ctx)

	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, CreatedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send pull request created event")
		return
	}

	logger.Infof("[event]: reported pull request created event with id '%s'", eventID)
}

func (r *Reader) RegisterCreated(
	fn events.HandlerFunc[*CreatedPayload],
	opts ...events.HandlerOption,
) error {
	return events.ReaderRegisterEvent(r.innerReader, CreatedEvent, fn, opts...)
}

const ClosedEvent events.EventType = "closed"

type ClosedPayload struct {
	Base
	SourceBranch string `json:"sourceBranch"`
	SourceSHA    string `json:"sourceSHA"`
}

func (r *Reporter) Closed(ctx context.Context, payload *ClosedPayload) {

	logger := logger.GetLogger(ctx)

	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, ClosedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send pull request closed event")
		return
	}

	logger.Infof("[event]: reported pull request closed event with id '%s'", eventID)
}

func (r *Reader) RegisterClosed(
	fn events.HandlerFunc[*ClosedPayload],
	opts ...events.HandlerOption,
) error {
	return events.ReaderRegisterEvent(r.innerReader, ClosedEvent, fn, opts...)
}

const ReopenedEvent events.EventType = "reopened"

type ReopenedPayload struct {
	Base
	SourceBranch string `json:"sourceBranch"`
	SourceSHA    string `json:"sourceSHA"`
	MergeBaseSHA string `json:"mergeBaseSHA"`
}

func (r *Reporter) Reopened(ctx context.Context, payload *ReopenedPayload) {

	logger := logger.GetLogger(ctx)

	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, ReopenedEvent, payload)
	if err != nil {
		logger.Errorf("[event]: failed to send pull request reopened event")
		return
	}

	logger.Infof("[event]: reported pull request reopened event with id '%s'", eventID)
}

func (r *Reader) RegisterReopened(
	fn events.HandlerFunc[*ReopenedPayload],
	opts ...events.HandlerOption,
) error {
	return events.ReaderRegisterEvent(r.innerReader, ReopenedEvent, fn, opts...)
}

const MergedEvent events.EventType = "merged"

type MergedPayload struct {
	Base
	TargetBranch string                `json:"targetBranch"`
	MergeMethod  repoModel.MergeMethod `json:"mergeMethod"`
	MergeSHA     string                `json:"mergeSHA"`
	TargetSHA    string                `json:"targetSHA"`
	SourceSHA    string                `json:"sourceSHA"`
}

func (r *Reporter) Merged(ctx context.Context, payload *MergedPayload) {

	logger := logger.GetLogger(ctx)

	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, MergedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send pull request merged event")
		return
	}

	logger.Infof("[event]: reported pull request merged event with id '%s'", eventID)
}

func (r *Reader) RegisterMerged(
	fn events.HandlerFunc[*MergedPayload],
	opts ...events.HandlerOption,
) error {
	return events.ReaderRegisterEvent(r.innerReader, MergedEvent, fn, opts...)
}
