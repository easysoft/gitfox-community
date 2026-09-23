// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	pipelineevents "github.com/easysoft/gitfox-community/app/events/pipeline"
	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/google/wire"
)

// PipelineEventWireSet provides a wire set for this package.
var PipelineEventWireSet = wire.NewSet(
	ProvidePipelineEventReaderFactory,
	ProvidePipelineEventReporter,
)

func ProvidePipelineEventReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*pipelineevents.Reader], error) {
	return pipelineevents.NewReaderFactory(eventsSystem)
}

func ProvidePipelineEventReporter(eventsSystem *events.System) (pipelineevents.Reporter, error) {
	r, err := pipelineevents.NewReporter(eventsSystem)
	if err != nil {
		return pipelineevents.Reporter{}, err
	}
	return *r, nil
}
