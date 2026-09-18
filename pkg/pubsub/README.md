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
	"github.com/easysoft/gitfox-community/pkg/pubsub"
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

## 技术实现
          
在 `pkg/pubsub` 模块中，内存（Memory）发布/订阅的实现逻辑非常直观，就像是**一群人（订阅者）在监听同一个广播（Topic）**。

简单来说，它的工作原理如下：

1.  **订阅（Subscribe）—— “建立管道，派人值守”**
    *   当你调用 `Subscribe` 时，系统会创建一个 **“订阅者” (MemorySubscriber)(结构体，内部依赖channel)**。
    *   每个订阅者都有一个自己的 **“小邮箱” (Go Channel)**，用来接收消息。
    *   同时，系统会启动一个 **“专职快递员” (Goroutine)**，他一直盯着这个小邮箱。一旦有新信件（消息）进来，他就立马拿出来交给你的处理函数 (`handler`) 去执行。
    *   系统会把这个订阅者登记在一个 **“花名册” (registry)** 上，记下他订阅了哪些话题（Topic）。

2.  **发布（Publish）—— “查找名单，分发信件”**
    *   当你调用 `Publish` 发送一条消息时，系统会拿出 **“花名册” (registry)** 从头到尾扫一遍。
    *   **匹配**：看谁订阅了这个话题（Topic）。
    *   **广播**：对于每一个匹配到的订阅者，系统会启动一个临时的 **“投递员” (Goroutine)**，尝试把消息塞进这个订阅者的 **“小邮箱” (Channel)** 里。
    *   **并发**：因为是给每个订阅者都派了单独的投递员，所以大家是同时收到消息的（并发）。
    *   **超时丢弃**：如果某个订阅者的邮箱满了（处理不过来），投递员等了一会儿（SendTimeout）还没塞进去，就会把信扔掉（丢弃消息），避免堵塞其他人。

**总结一下核心技术点：**

*   **Registry (花名册)**: 一个切片，存了所有订阅者。
*   **Channel (管道/邮箱)**: 数据的传输通道，带有缓冲（ChannelSize），起到了削峰填谷的作用。
*   **Goroutine (协程/投递员)**:
    *   **消费者端**: 每个订阅者有一个长驻协程，不断从 Channel 取数据处理。
    *   **生产者端**: 每次发布，会为每个目标订阅者启动一个短协程去塞数据，互不影响。

这就是一个典型的 **“多对多、异步、解耦”** 的消息模型。

## Q&A

1、为什么初始化时有一个 Namespace，发布/订阅时又可以传 Namespace？
          
在 `pkg/pubsub` 包中，`Namespace` 主要用于**隔离不同业务场景下的 Topic，防止名称冲突**。具体来说，代码中有两处关于 Namespace 的设置：

1.  **全局默认 Namespace (`Config.Namespace`)**
    *   **定义位置**: `pkg/pubsub/config.go` 中的 `Config` 结构体。
    *   **设置方式**: 在创建 PubSub 实例时，通过 `NewMemory(WithNamespace("default"))` 进行设置。
    *   **作用**: 如果在发布或订阅时没有显式指定 Namespace，就会使用这个默认值。这相当于给整个 PubSub 实例设定了一个基准命名空间。

2.  **操作级 Namespace**
    *   **发布时 (`PublishConfig.namespace`)**:
        *   **定义位置**: `pkg/pubsub/optinons.go` 中的 `PublishConfig`。
        *   **设置方式**: 调用 `Publish` 方法时，通过 `WithPublishNamespace("custom")` 选项传入。
        *   **作用**: 允许某次发布操作临时覆盖默认 Namespace，将消息发送到特定的命名空间中。
    *   **订阅时 (`SubscribeConfig.namespace`)**:
        *   **定义位置**: `pkg/pubsub/optinons.go` 中的 `SubscribeConfig`。
        *   **设置方式**: 调用 `Subscribe` 方法时，通过 `WithChannelNamespace("custom")` 选项传入。
        *   **作用**: 允许某个消费者只监听特定命名空间下的 Topic。

**总结：**
最终的 Topic 名称是通过 `formatTopic` 函数生成的，格式为 `app:namespace:topic`。
*   **第一层 (全局)**: 是 PubSub 实例级别的默认配置，方便统一管理。
*   **第二层 (操作级)**: 是方法级别的覆盖配置，提供了更细粒度的控制，允许在同一个 PubSub 实例中处理多个隔离的业务流（例如 `system` 命名空间用于系统通知，`chat` 命名空间用于聊天消息）。