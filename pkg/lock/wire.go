package lock

import (
	"time"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvideMutexManager,
)

func ProvideMutexManager() MutexManager {
	// TODO: 仅支持内存锁，后续支持 Redis 锁
	// 内存锁默认过期时间 8 秒，重试次数 8 次，重试延迟 250 毫秒，漂移因子 0.01，超时因子 0.25
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
