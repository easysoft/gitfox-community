// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/pkg/logger"

	"github.com/easysoft/gitfox-community/pkg/events"
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
