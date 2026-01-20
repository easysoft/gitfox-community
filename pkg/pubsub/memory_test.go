package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemory_BasicPubSub(t *testing.T) {
	ctx := context.Background()
	ps := NewMemory(WithSendTimeout(time.Second))

	topic := "test-topic"
	payload := []byte("hello world")
	received := make(chan []byte, 1)

	// Subscribe
	sub := ps.Subscribe(ctx, topic, func(p []byte) error {
		received <- p
		return nil
	})
	time.Sleep(50 * time.Millisecond)
	defer sub.Close()

	// Publish
	err := ps.Publish(ctx, topic, payload)
	assert.NoError(t, err)

	// Verify
	select {
	case p := <-received:
		assert.Equal(t, payload, p)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestMemory_MultipleSubscribers(t *testing.T) {
	ctx := context.Background()
	ps := NewMemory(WithSendTimeout(time.Second))

	topic := "broadcast"
	payload := []byte("data")

	count := 3
	wg := make(chan bool, count)

	for i := 0; i < count; i++ {
		ps.Subscribe(ctx, topic, func(p []byte) error {
			if string(p) == string(payload) {
				wg <- true
			}
			return nil
		})
	}
	time.Sleep(50 * time.Millisecond)

	err := ps.Publish(ctx, topic, payload)
	assert.NoError(t, err)

	for i := 0; i < count; i++ {
		select {
		case <-wg:
		case <-time.After(time.Second):
			t.Fatal("timeout waiting for subscribers")
		}
	}
}

func TestMemory_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	ps := NewMemory(WithSendTimeout(time.Second))

	topic := "dynamic"
	received := make(chan []byte, 1)

	sub := ps.Subscribe(ctx, topic, func(p []byte) error {
		received <- p
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	// Publish 1
	ps.Publish(ctx, topic, []byte("1"))
	select {
	case <-received:
	case <-time.After(time.Second):
		t.Fatal("timeout 1")
	}

	// Unsubscribe
	err := sub.Unsubscribe(ctx, topic)
	assert.NoError(t, err)

	// Publish 2
	ps.Publish(ctx, topic, []byte("2"))
	select {
	case <-received:
		t.Fatal("received message after unsubscribe")
	case <-time.After(100 * time.Millisecond):
		// Expected timeout
	}
}

func TestMemory_NamespaceIsolation(t *testing.T) {
	ctx := context.Background()
	ps := NewMemory(WithSendTimeout(time.Second))

	topic := "secret"
	received := make(chan []byte, 1)

	// Subscribe to default namespace
	ps.Subscribe(ctx, topic, func(p []byte) error {
		received <- p
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	// Publish to different namespace
	ps.Publish(ctx, topic, []byte("data"), WithPublishNamespace("other"))

	select {
	case <-received:
		t.Fatal("should not receive message from other namespace")
	case <-time.After(100 * time.Millisecond):
		// OK
	}

	// Publish to default namespace
	ps.Publish(ctx, topic, []byte("data"))
	select {
	case <-received:
		// OK
	case <-time.After(time.Second):
		t.Fatal("should receive message from default namespace")
	}
}

func TestMemory_Close(t *testing.T) {
	ctx := context.Background()
	ps := NewMemory(WithSendTimeout(time.Second))

	topic := "topic"
	received := make(chan []byte, 1)

	ps.Subscribe(ctx, topic, func(p []byte) error {
		received <- p
		return nil
	})
	time.Sleep(50 * time.Millisecond)

	ps.Close(ctx)

	// Publish should likely fail silently or return no error but message won't be delivered if closed
	// Actually implementation checks !sub.isClosed()
	err := ps.Publish(ctx, topic, []byte("msg"))
	assert.NoError(t, err)

	select {
	case <-received:
		t.Fatal("should not receive message after close")
	case <-time.After(100 * time.Millisecond):
		// OK
	}
}
