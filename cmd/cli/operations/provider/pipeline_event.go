package provider

import (
	pipelineevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pipeline"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"

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

func ProvidePipelineEventReporter(eventsSystem *events.System) (*pipelineevents.Reporter, error) {
	return pipelineevents.NewReporter(eventsSystem)
}
