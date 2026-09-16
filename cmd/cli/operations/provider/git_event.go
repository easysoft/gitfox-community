package provider

import (
	gitevents "github.com/easysoft/gitfox-community/app/events/git"
	"github.com/easysoft/gitfox-community/pkg/events"

	"github.com/google/wire"
)

// GitEventWireSet provides a wire set for this package.
var GitEventWireSet = wire.NewSet(
	ProvideGitEventReaderFactory,
	ProvideGitEventReporter,
)

func ProvideGitEventReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*gitevents.Reader], error) {
	return gitevents.NewReaderFactory(eventsSystem)
}

func ProvideGitEventReporter(eventsSystem *events.System) (*gitevents.Reporter, error) {
	return gitevents.NewReporter(eventsSystem)
}
