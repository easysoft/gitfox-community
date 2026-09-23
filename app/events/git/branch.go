// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const BranchCreatedEvent events.EventType = "branch-created"

type BranchCreatedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	SHA     string `json:"sha"`
}

func (r *Reporter) BranchCreated(ctx context.Context, payload *BranchCreatedPayload) {

	logger := logger.GetLogger(ctx)

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, BranchCreatedEvent, payload)
	if err != nil {
		logger.Error("failed to send branch created event")
		return
	}

	logger.Infof("reported branch created event with id '%s'", eventID)
}

func (r *Reader) RegisterBranchCreated(fn events.HandlerFunc[*BranchCreatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, BranchCreatedEvent, fn, opts...)
}

const BranchUpdatedEvent events.EventType = "branch-updated"

type BranchUpdatedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	OldSHA  string `json:"oldSHA"`
	NewSHA  string `json:"newSHA"`
	Forced  bool   `json:"forced"`
	// CommitMessage is the full message of the commit NewSHA points to (best effort, may be empty).
	CommitMessage string `json:"commitMessage,omitempty"`
}

func (r *Reporter) BranchUpdated(ctx context.Context, payload *BranchUpdatedPayload) {

	logger := logger.GetLogger(ctx)
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, BranchUpdatedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send branch updated event")
		return
	}

	logger.Infof("[event]: reported branch updated event with id '%s'", eventID)
}

func (r *Reader) RegisterBranchUpdated(fn events.HandlerFunc[*BranchUpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, BranchUpdatedEvent, fn, opts...)
}

const BranchDeletedEvent events.EventType = "branch-deleted"

type BranchDeletedPayload struct {
	RepoID  int64  `json:"repoID"`
	Account string `json:"account"`
	Ref     string `json:"ref"`
	SHA     string `json:"sha"`
}

func (r *Reporter) BranchDeleted(ctx context.Context, payload *BranchDeletedPayload) {
	logger := logger.GetLogger(ctx)
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, BranchDeletedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send branch deleted event")
		return
	}

	logger.Infof("[event]: reported branch deleted event with id '%s'", eventID)
}

func (r *Reader) RegisterBranchDeleted(fn events.HandlerFunc[*BranchDeletedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, BranchDeletedEvent, fn, opts...)
}
