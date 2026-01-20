# SSE (Server-Sent Events) 服务模块

`sse` 模块基于 `pkg/pubsub` 实现了服务器发送事件（Server-Sent Events）的流式服务。它主要用于在后端发生特定业务事件时，实时向前端推送数据变更。

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

## 技术实现

SSE 模块充当了 **业务事件** 与 **HTTP 长连接** 之间的桥梁，其核心流程如下：

1.  **事件总线 (PubSub)**: 底层依赖 `pkg/pubsub`。当业务发生变更（如 `Publish`）时，本质上是向 PubSub 的特定 Topic（如 `spaces:1001`）发送了一条消息。
2.  **管道转换 (Channel)**: `Stream` 方法调用时，会创建一个临时的订阅者。这个订阅者会把从 PubSub 收到的原始 `[]byte` 消息，反序列化为 `Event` 对象，并塞入一个 Go Channel (`<-chan *Event`)。
3.  **HTTP 长连接**: 在 Controller 层（Web 接口），通过 `for select` 循环不断从这个 Channel 读取数据。每读到一个事件，就按照 [SSE 协议规范](https://html.spec.whatwg.org/multipage/server-sent-events.html)（`data: ...\n\n`）直接写入 HTTP Response Body，并刷新缓冲区，从而实现实时推送。


## Q&A

1、SSE 是怎么实现单向长连接的？

本质上SSE 是“一直没发完”的 HTTP 请求。

- 原理 :
  1. 浏览器发一个普通的 HTTP GET 请求给服务器。
  2. 服务器响应头里说： Content-Type: text/event-stream 。
  3. 关键点 : 服务器 不发送 Content-Length ，也不关闭连接。
  4. 服务器每隔一会就往这个连接里写一行数据（ data: hello\n\n ），然后 Flush 一下。
  5. 浏览器收到数据就处理，但依然保持连接打开，等待下一行数据。
- 形象比喻 :
    - Webhook : 寄信。我写好信（事件），投递到你的邮箱（URL），然后我就走了。
    - WebSocket : 打电话。我们俩同时拿着听筒，我能说话，你也能说话，谁都可以随时打断对方。
    - SSE : 收音机/大喇叭 。我（服务器）一直拿着麦克风说话，你（浏览器）只能戴着耳机听。你不能通过这个耳机跟我说话（单向），只要我不关麦，你就能一直听到新消息（长连接）

2、 在gitfox中都有哪些应用场景？

| 事件类型 (`EventType`) | 触发场景 | 推送数据内容 (`Data`) | 作用 |
| :--- | :--- | :--- | :--- |
| **Pull Request 相关** | | | |
| `pullreq_updated` | PR 创建、编辑、合并、关闭、源分支更新 | `PullReq` 对象 (JSON) | 刷新 PR 列表状态、合并按钮状态 |
| `pullreq_reviewer_added` | 添加评审人 | - | 刷新评审人列表 |
| `pullreq_reviewer_removed` | 移除评审人 | - | 刷新评审人列表 |
| `pullreq_ai_review_updated`| AI 评审状态更新 | - | 显示 AI 评审进度 |
| **CI/CD 流水线相关** | | | |
| `execution_updated` | 流水线状态变更 (如 Pending -> Running) | `Execution` 对象 | 实时显示 Pipeline 进度条 |
| `execution_running` | 流水线开始运行 | `Execution` 对象 | 标记开始 |
| `execution_completed` | 流水线执行结束 | `Execution` 对象 | 显示最终结果 (成功/失败) |
| `execution_canceled` | 流水线被取消 | `Execution` 对象 | 显示取消状态 |
| `log_line_appended` | 产生新的日志行 | 日志片段 | 实时滚动展示构建日志 |
| **仓库操作相关** | | | |
| `repository_import_completed`| 仓库导入完成 | `Repository` 对象 | 通知用户导入已完成 |
| `repository_export_completed`| 仓库导出完成 | `Repository` 对象 | 通知用户下载链接可用 |

---