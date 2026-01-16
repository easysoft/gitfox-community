package events

import (
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
)

func NewReaderFactory(eventsSystem *events.System) (*events.ReaderFactory[*Reader], error) {
	readerFactoryFunc := func(innerReader *events.GenericReader) (*Reader, error) {
		return &Reader{
			innerReader: innerReader,
		}, nil
	}

	return events.NewReaderFactory(eventsSystem, category, readerFactoryFunc)
}

// Reader is the event reader for this package.
// It exposes typesafe event registration methods for all events by this package.
// NOTE: Event registration methods are in the event's dedicated file.
type Reader struct {
	innerReader *events.GenericReader
}

func (r *Reader) Configure(opts ...events.ReaderOption) {
	r.innerReader.Configure(opts...)
}
