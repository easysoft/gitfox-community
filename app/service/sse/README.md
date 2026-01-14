# SSE (Server-Sent Events) 服务模块

`sse` 模块基于 `pkg/pubsub` 实现了服务器发送事件（Server-Sent Events）的流式服务。它主要用于在后端发生特定业务事件时，实时向前端或其他客户端推送数据变更。

该模块将业务层面的“事件”概念与底层的“发布/订阅”机制结合，提供了一套面向业务的流式接口。

## 功能特性

*   **基于 PubSub**：利用 `pubsub` 模块作为底层消息总线，支持解耦的消息分发。
*   **结构化事件**：定义了标准的 `Event` 结构（包含 Type 和 Data），自动处理 JSON 序列化。
*   **空间（Space）隔离**：默认支持基于 `spaceID` 的频道隔离，适合多租户或多项目场景（底层对应 PubSub Topic `spaces:<id>`）。
*   **流式消费**：提供 Go Channel 风格的 API (`<-chan *Event`)，方便在 HTTP Handler 中实现长连接推送。

## 接口说明

核心接口 `Streamer` 定义了两个主要方法：

*   `Publish(ctx, spaceID, eventType, data)`: 向指定空间发布一个业务事件。
*   `Stream(ctx, spaceID)`: 订阅指定空间的事件流，返回事件通道、错误通道和清理函数。

## 使用示例

### 1. 初始化

通常通过依赖注入框架（如 Wire）创建 `Streamer` 实例。

```go
import (
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
    "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/service/sse"
)

// 假设已有一个 pubsub 实例
var ps pubsub.PubSub = ... 

// 创建 SSE 服务，指定 namespace
sseSvc := sse.NewSSE(ps, "gitfox-sse")
```

### 2. 发布事件 (Publish)

在业务逻辑中（例如创建 Pull Request 后），调用 `Publish` 推送事件。

```go
import "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/event"

func (s *PullRequestService) CreatePR(ctx context.Context, req CreatePRReq) error {
    // ... 业务逻辑 ...
    
    // 推送事件：通知该仓库所在空间的用户有新 PR 创建
    eventData := map[string]interface{}{
        "pr_id": 101,
        "title": "Fix bug",
    }
    
    // spaceID 通常对应仓库所属的项目或组 ID
    err := s.sseSvc.Publish(ctx, req.SpaceID, event.SSETypePullRequestCreated, eventData)
    if err != nil {
        logger.Error("failed to publish sse event", err)
    }
    
    return nil
}
```

### 3. 消费事件流 (Stream)

在 HTTP Handler 中建立 SSE 连接，将事件流实时推送给前端。

```go
func (c *Controller) HandleSSE(ctx *gin.Context) {
    spaceID := parseSpaceID(ctx)

    // 开始订阅流
    events, errs, cleanup := c.sseSvc.Stream(ctx.Request.Context(), spaceID)
    defer cleanup(context.Background())

    // 设置 SSE 响应头
    ctx.Writer.Header().Set("Content-Type", "text/event-stream")
    ctx.Writer.Header().Set("Cache-Control", "no-cache")
    ctx.Writer.Header().Set("Connection", "keep-alive")

    for {
        select {
        case <-ctx.Request.Context().Done():
            return
        case err := <-errs:
            // 处理错误（可选）
            return
        case event := <-events:
            // 发送 SSE 格式数据
            // 格式: data: <json_payload>\n\n
            ctx.Writer.WriteString("event: " + string(event.Type) + "\n")
            ctx.Writer.WriteString("data: " + string(event.Data) + "\n\n")
            ctx.Writer.Flush()
        }
    }
}
```