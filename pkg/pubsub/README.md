# PubSub 模块

`pubsub` 是一个轻量级的发布/订阅（Publish/Subscribe）系统接口封装，旨在为应用程序提供解耦的消息传递机制。目前默认提供基于内存（In-Memory）的实现，并为未来扩展支持 Redis 等其他中间件预留了接口。

## 功能特性

*   **多后端支持**：定义了通用的 `PubSub`、`Publisher` 和 `Consumer` 接口，方便切换底层实现（目前支持 Memory）。
*   **内存实现**：内置高性能的内存消息队列，支持并发读写。
*   **命名空间隔离**：支持 Namespace，不同命名空间下的 Topic 相互隔离，互不干扰。
*   **灵活配置**：支持自定义缓冲区大小、发送超时时间、健康检查间隔等。
*   **广播机制**：支持多个消费者订阅同一个 Topic，消息会广播给所有订阅者。
*   **Wire 集成**：提供 `WireSet`，方便使用 Google Wire 进行依赖注入。

## 接口说明

### 核心接口

*   `Publisher`: 定义发布消息的方法 `Publish`。
*   `Consumer`: 定义订阅者行为，包括 `Subscribe`、`Unsubscribe` 和 `Close`。
*   `PubSub`: 组合接口，包含 `Publisher` 和创建 `Consumer` 的 `Subscribe` 方法。

### 配置项

可以通过 `Option` 函数在创建实例时进行配置：

*   `WithApp(name string)`: 设置应用名称。
*   `WithNamespace(name string)`: 设置默认命名空间。
*   `WithSendTimeout(d time.Duration)`: 设置发送超时时间。
*   `WithSize(size int)`: 设置内部 Channel 缓冲区大小。

## 使用示例

### 1. 初始化

使用 `NewMemory` 创建一个内存实现的 PubSub 实例：

```go
package main

import (
	"context"
	"time"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

func main() {
	// 创建 PubSub 实例
	ps := pubsub.NewMemory(
		pubsub.WithApp("my-app"),
		pubsub.WithNamespace("default"),
		pubsub.WithSendTimeout(time.Second),
		pubsub.WithSize(1024),
	)
	
	// ... 使用 ps
}
```

或者使用 Wire 注入：

```go
// wire.go
var Set = wire.NewSet(
    pubsub.WireSet,
    // ...
)
```

### 2. 订阅消息 (Subscribe)

订阅者通过回调函数处理接收到的消息。`Subscribe` 方法返回一个 `Consumer` 对象，用于后续管理订阅状态。

```go
ctx := context.Background()
topic := "user.created"

// 订阅消息
consumer := ps.Subscribe(ctx, topic, func(payload []byte) error {
    println("Received message:", string(payload))
    return nil
})

// 程序退出时关闭消费者
defer consumer.Close()
```

### 3. 发布消息 (Publish)

发布者将消息发送到指定的 Topic。

```go
ctx := context.Background()
topic := "user.created"
payload := []byte(`{"id": 1, "name": "Alice"}`)

// 发布消息
err := ps.Publish(ctx, topic, payload)
if err != nil {
    panic(err)
}
```

### 4. 命名空间隔离

可以在发布和订阅时指定特定的命名空间，实现消息隔离。

```go
// 发布到 "system" 命名空间
ps.Publish(ctx, "alerts", []byte("CPU High"), pubsub.WithPublishNamespace("system"))

// 订阅 "system" 命名空间
ps.Subscribe(ctx, "alerts", handler, pubsub.WithChannelNamespace("system"))
```