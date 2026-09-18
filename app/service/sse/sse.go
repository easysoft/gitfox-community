package sse

import (
	"context"
	"encoding/json"
	"strconv"

	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	"github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
)

// Event is a server sent event.
type Event struct {
	Type eventModel.SSEType `json:"type"`
	Data json.RawMessage    `json:"data"`
}

type Streamer interface {
	// Publish publishes an event to a given space ID.
	Publish(ctx context.Context, spaceID int64, eventType eventModel.SSEType, data any) error

	// Stream streams the events on a space ID.
	Stream(ctx context.Context, spaceID int64) (<-chan *Event, <-chan error, func(context.Context) error)
}

type pubsubStreamer struct {
	pubsub    pubsub.PubSub
	namespace string
}

func NewSSE(pubsub pubsub.PubSub, namespace string) Streamer {
	return &pubsubStreamer{
		pubsub:    pubsub,
		namespace: namespace,
	}
}

func (e *pubsubStreamer) Publish(ctx context.Context, spaceID int64, eventType eventModel.SSEType, data any) error {
	dataSerialized, err := json.Marshal(data)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize data")
	}
	event := Event{
		Type: eventType,
		Data: dataSerialized,
	}
	serializedEvent, err := json.Marshal(event)
	if err != nil {
		return errors.Wrapf(err, "failed to serialize event")
	}
	namespaceOption := pubsub.WithPublishNamespace(e.namespace)
	topic := getSpaceTopic(spaceID)
	err = e.pubsub.Publish(ctx, topic, serializedEvent, namespaceOption)
	if err != nil {
		return errors.Wrapf(err, "failed to publish event on pubsub")
	}

	return nil
}

func (e *pubsubStreamer) Stream(
	ctx context.Context,
	spaceID int64,
) (<-chan *Event, <-chan error, func(context.Context) error) {
	chEvent := make(chan *Event, 100) // TODO: check best size here
	chErr := make(chan error)
	g := func(payload []byte) error {
		event := &Event{}
		err := json.Unmarshal(payload, event)
		if err != nil {
			// This should never happen
			return err
		}
		select {
		case chEvent <- event:
		default:
		}

		return nil
	}
	namespaceOption := pubsub.WithChannelNamespace(e.namespace)
	topic := getSpaceTopic(spaceID)
	consumer := e.pubsub.Subscribe(ctx, topic, g, namespaceOption)
	cleanupFN := func(_ context.Context) error {
		return consumer.Close()
	}

	return chEvent, chErr, cleanupFN
}

// getSpaceTopic creates the namespace name which will be `spaces:<id>`.
func getSpaceTopic(spaceID int64) string {
	return "spaces:" + strconv.Itoa(int(spaceID))
}
