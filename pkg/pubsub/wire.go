package pubsub

import (
	"time"

	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	ProvidePubSub,
)

func ProvidePubSub() PubSub {
	// TODO: 仅支持内存发布订阅，后续支持 Redis 发布订阅
	// 内存发布订阅默认健康检查间隔 10 秒，发送超时 10 秒，通道大小 1024
	return NewMemory(
		WithApp("gitfox"),
		WithNamespace("default"),
		WithHealthCheckInterval(3*time.Second),
		WithSendTimeout(60*time.Second),
		WithSize(100),
	)
}
