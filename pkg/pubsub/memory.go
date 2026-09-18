package pubsub

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/easysoft/gitfox-community/pkg/logger"
)

var (
	ErrClosed = errors.New("pubsub: subscriber is closed")
)

type Memory struct {
	config   Config
	mutex    sync.Mutex
	registry []*MemorySubscriber
}

// NewMemory create an instance of memory pubsub implementation.
func NewMemory(options ...Option) *Memory {
	config := Config{
		App:            "app",
		Namespace:      "default",
		HealthInterval: 3 * time.Second,
		SendTimeout:    60,
		ChannelSize:    100,
	}

	for _, f := range options {
		f.Apply(&config)
	}
	return &Memory{
		config:   config,
		registry: make([]*MemorySubscriber, 0, 16),
	}
}

// Subscribe consumer to process the event with payload.
func (r *Memory) Subscribe(
	ctx context.Context,
	topic string,
	handler func(payload []byte) error,
	options ...SubscribeOption,
) Consumer {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	config := SubscribeConfig{
		topics:      make([]string, 0, 8),
		app:         r.config.App,
		namespace:   r.config.Namespace,
		sendTimeout: r.config.SendTimeout,
		channelSize: r.config.ChannelSize,
	}

	for _, f := range options {
		f.Apply(&config)
	}

	// create subscriber and map it to the registry
	subscriber := &MemorySubscriber{
		config:  &config,
		handler: handler,
	}

	config.topics = append(config.topics, topic)
	subscriber.topics = subscriber.formatTopics(config.topics...)

	// start subscriber
	go subscriber.start(ctx)

	// register subscriber
	r.registry = append(r.registry, subscriber)

	return subscriber
}

// Publish event to message broker with payload.
func (r *Memory) Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error {
	logger := logger.GetLogger(ctx)
	if len(r.registry) == 0 {
		logger.Warn("in pubsub Publish: no subscribers registered")
		return nil
	}
	pubConfig := PublishConfig{
		app:       r.config.App,
		namespace: r.config.Namespace,
	}
	for _, f := range opts {
		f.Apply(&pubConfig)
	}

	topic = formatTopic(pubConfig.app, pubConfig.namespace, topic)
	wg := sync.WaitGroup{}
	for _, sub := range r.registry {
		if slices.Contains(sub.topics, topic) && !sub.isClosed() {
			wg.Add(1)
			go func(subscriber *MemorySubscriber) {
				defer wg.Done()
				// timer is based on subscriber data
				t := time.NewTimer(subscriber.config.sendTimeout)
				defer t.Stop()
				select {
				case <-ctx.Done():
					return
				case subscriber.channel <- payload:
					logger.Debugf("in pubsub Publish: message %v sent to topic %s", string(payload), topic)
				case <-t.C:
					// channel is full for topic (message is dropped)
					logger.Warnf("in pubsub Publish: %s topic is full for %s (message is dropped)",
						topic, subscriber.config.sendTimeout)
				}
			}(sub)
		}
	}

	// Wait for all subscribers to complete
	// Otherwise, we might fail notifying some subscribers due to context completion.
	wg.Wait()

	return nil
}

func (r *Memory) Close(_ context.Context) error {
	for _, subscriber := range r.registry {
		if err := subscriber.Close(); err != nil {
			return err
		}
	}
	return nil
}

type MemorySubscriber struct {
	config  *SubscribeConfig
	handler func([]byte) error
	channel chan []byte
	once    sync.Once
	mutex   sync.RWMutex
	topics  []string
	closed  bool
}

func (s *MemorySubscriber) start(ctx context.Context) {
	logger := logger.GetLogger(ctx)
	s.channel = make(chan []byte, s.config.channelSize)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-s.channel:
			if !ok {
				return
			}
			if err := s.handler(msg); err != nil {
				// TODO: bump err to caller
				logger.Error("in pubsub start: error while running handler for topic")
			}
		}
	}
}

func (s *MemorySubscriber) Subscribe(_ context.Context, topics ...string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	topics = s.formatTopics(topics...)
	for _, ch := range topics {
		if slices.Contains(s.topics, ch) {
			continue
		}
		s.topics = append(s.topics, ch)
	}
	return nil
}

func (s *MemorySubscriber) Unsubscribe(_ context.Context, topics ...string) error {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	topics = s.formatTopics(topics...)
	for i, ch := range topics {
		if slices.Contains(s.topics, ch) {
			s.topics[i] = s.topics[len(s.topics)-1]
			s.topics = s.topics[:len(s.topics)-1]
		}
	}
	return nil
}

func (s *MemorySubscriber) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.closed {
		return ErrClosed
	}
	s.closed = true
	s.once.Do(func() {
		close(s.channel)
	})
	return nil
}

func (s *MemorySubscriber) isClosed() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.closed
}

func (s *MemorySubscriber) formatTopics(topics ...string) []string {
	result := make([]string, len(topics))
	for i, topic := range topics {
		result[i] = formatTopic(s.config.app, s.config.namespace, topic)
	}
	return result
}
