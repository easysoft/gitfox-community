package events

import (
	"context"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/stream"
)

// StreamProducer is an abstraction of a producer from the streams package.
type StreamProducer interface {
	Send(ctx context.Context, streamID string, payload map[string]interface{}) (string, error)
}

// StreamConsumer is an abstraction of a consumer from the streams package.
type StreamConsumer interface {
	Register(streamID string, handler stream.HandlerFunc, opts ...stream.HandlerOption) error
	Configure(opts ...stream.ConsumerOption)
	Start(ctx context.Context) error
	Errors() <-chan error
	Infos() <-chan string
}

// StreamConsumerFactoryFunc is an abstraction of a factory method for stream consumers.
type StreamConsumerFactoryFunc func(groupName string, consumerName string) (StreamConsumer, error)
