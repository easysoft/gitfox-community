// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package events

import (
	"context"

	"github.com/easysoft/gitfox-community/app/constant"
	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

const ExecutedEvent events.EventType = "executed"

type ExecutedPayload struct {
	PipelineID   uint              `json:"pipelineID"`
	ExecutionNum uint              `json:"executionID"`
	Status       constant.CIStatus `json:"status"`
}

func (r *Reporter) Executed(ctx context.Context, payload *ExecutedPayload) {
	eventID, err := events.ReporterSendEvent(r.innerReporter, ctx, ExecutedEvent, payload)
	if err != nil {
		logger.GetLogger(ctx).Errorw("failed to send pipeline executed event", "error", err)
		return
	}

	logger.GetLogger(ctx).Debugf("reported pipeline executed event with id '%s'", eventID)
}

func (r *Reader) RegisterExecuted(fn events.HandlerFunc[*ExecutedPayload],
	opts ...events.HandlerOption) error {
	return events.ReaderRegisterEvent(r.innerReader, ExecutedEvent, fn, opts...)
}
