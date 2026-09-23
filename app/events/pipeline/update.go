// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

const UpdatedEvent events.EventType = "updated"

type UpdatedPayload struct {
	PipelineID int64 `json:"pipelineID"`
}

func (r *Reporter) Updated(ctx context.Context, payload *UpdatedPayload) {
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, UpdatedEvent, payload)
	if err != nil {
		logger.GetLogger(ctx).Errorw("failed to send pipeline updated event", "error", err)
		return
	}

	logger.GetLogger(ctx).Debugf("reported pipeline update event with id '%s'", eventID)
}

func (r *Reader) RegisterUpdated(fn events.HandlerFunc[*UpdatedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, UpdatedEvent, fn, opts...)
}
