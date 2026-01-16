package provider

import (
	"errors"
	"fmt"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/events"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/stream"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
)

// EventWireSet provides a wire set for this package.
var EventWireSet = wire.NewSet(
	ProvideEventSystem,
	NewEventConfig,
	NewEventRedisClient,
)

func NewEventConfig() events.Config {
	return events.Config{
		Mode:                  events.ModeInMemory,
		Namespace:             "gitfox",
		MaxStreamLength:       1000,
		ApproxMaxStreamLength: true,
	}
}

func NewEventRedisClient() redis.UniversalClient {
	return nil
}

func ProvideEventSystem(config events.Config, redisClient redis.UniversalClient) (*events.System, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("provided config is invalid: %w", err)
	}

	var system *events.System
	var err error
	switch config.Mode {
	case events.ModeInMemory:
		system, err = provideSystemInMemory(config)
	case events.ModeRedis:
		system, err = provideSystemRedis(config, redisClient)
	default:
		return nil, fmt.Errorf("events system mode '%s' is not supported", config.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to setup event system for mode '%s': %w", config.Mode, err)
	}

	return system, nil
}

func provideSystemInMemory(config events.Config) (*events.System, error) {
	broker, err := stream.NewMemoryBroker(config.MaxStreamLength)
	if err != nil {
		return nil, err
	}

	return events.NewSystem(
		newMemoryStreamConsumerFactoryMethod(broker, config.Namespace),
		newMemoryStreamProducer(broker, config.Namespace),
	)
}

func provideSystemRedis(config events.Config, redisClient redis.UniversalClient) (*events.System, error) {
	if redisClient == nil {
		return nil, errors.New("redis client required")
	}

	return events.NewSystem(
		newRedisStreamConsumerFactoryMethod(redisClient, config.Namespace),
		newRedisStreamProducer(redisClient, config.Namespace,
			config.MaxStreamLength, config.ApproxMaxStreamLength),
	)
}

func newMemoryStreamConsumerFactoryMethod(broker *stream.MemoryBroker, namespace string) events.StreamConsumerFactoryFunc {
	return func(groupName string, _ string) (events.StreamConsumer, error) {
		return stream.NewMemoryConsumer(broker, namespace, groupName)
	}
}

func newMemoryStreamProducer(broker *stream.MemoryBroker, namespace string) events.StreamProducer {
	return stream.NewMemoryProducer(broker, namespace)
}

func newRedisStreamConsumerFactoryMethod(
	redisClient redis.UniversalClient,
	namespace string,
) events.StreamConsumerFactoryFunc {
	return func(groupName string, consumerName string) (events.StreamConsumer, error) {
		return stream.NewRedisConsumer(redisClient, namespace, groupName, consumerName)
	}
}

func newRedisStreamProducer(redisClient redis.UniversalClient, namespace string,
	maxStreamLength int64, approxMaxStreamLength bool) events.StreamProducer {
	return stream.NewRedisProducer(redisClient, namespace, maxStreamLength, approxMaxStreamLength)
}
