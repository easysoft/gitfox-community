package lock

import (
	"time"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideMutexManager,
)

func ProvideMutexManager() MutexManager {
	return NewInMemory(Config{
		App:           "gitfox",
		Namespace:     "default",
		Provider:      MemoryProvider,
		Expiry:        8 * time.Second,
		Tries:         8,
		RetryDelay:    250 * time.Millisecond,
		DriftFactor:   0.01,
		TimeoutFactor: 0.25,
	})
}
