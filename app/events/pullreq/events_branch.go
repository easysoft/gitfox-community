// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/events"
	"github.com/easysoft/gitfox-community/pkg/logger"
)

const BranchUpdatedEvent events.EventType = "branch-updated"

type BranchUpdatedPayload struct {
	Base
	SourceBranch    string `json:"sourceBranch"`
	OldSHA          string `json:"oldSHA"`
	NewSHA          string `json:"newSHA"`
	OldMergeBaseSHA string `json:"oldMergeBaseSHA"`
	NewMergeBaseSHA string `json:"newMergeBaseSHA"`
	Forced          bool   `json:"forced"`
}

func (r *Reporter) BranchUpdated(ctx context.Context, payload *BranchUpdatedPayload) {
	logger := logger.GetLogger(ctx)
	if payload == nil {
		return
	}

	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, BranchUpdatedEvent, payload)
	if err != nil {
		logger.Error("[event]: failed to send pull request branch updated event")
		return
	}

	logger.Infof("[event]: reported pull request branch updated event with id '%s'", eventID)
}

func (r *Reader) RegisterBranchUpdated(fn events.HandlerFunc[*BranchUpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, BranchUpdatedEvent, fn, opts...)
}
