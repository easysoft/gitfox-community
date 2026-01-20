package provider

import (
	pullreqevents "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/events/pullreq"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"

	"github.com/google/wire"
)

// PullReqEventWireSet provides a wire set for this package.
var PullReqEventWireSet = wire.NewSet(
	ProvidePullReqReaderFactory,
	ProvidePullReqReporter,
)

func ProvidePullReqReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*pullreqevents.Reader], error) {
	return pullreqevents.NewReaderFactory(eventsSystem)
}

func ProvidePullReqReporter(eventsSystem *events.System) (*pullreqevents.Reporter, error) {
	return pullreqevents.NewReporter(eventsSystem)
}
