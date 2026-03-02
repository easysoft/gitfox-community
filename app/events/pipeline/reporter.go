package events

import (
	"errors"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

// Reporter is the event reporter for this package.
// It exposes typesafe send methods for all events of this package.
// NOTE: Event send methods are in the event's dedicated file.
type Reporter struct {
	innerReporter *events.GenericReporter
}

func NewReporter(eventsSystem *events.System) (*Reporter, error) {
	innerReporter, err := events.NewReporter(eventsSystem, category)
	if err != nil {
		return nil, errors.New("failed to create new GenericReporter from event system")
	}

	return &Reporter{
		innerReporter: innerReporter,
	}, nil
}
