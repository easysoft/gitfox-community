package main

import (
	"context"
	"fmt"
	"log"
	"time"

	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
)

// 这个示例演示了如何使用 sse 包来实现简单的发布和订阅流程
func main() {
	// 1. 初始化底层的 PubSub (使用内存实现)
	// 在实际应用中，这通常由 Wire 依赖注入完成
	ps := pubsub.NewMemory(
		pubsub.WithApp("gitfox-demo"),
		pubsub.WithNamespace("default"),
		pubsub.WithSendTimeout(time.Second),
		pubsub.WithSize(100),
	)

	// 2. 初始化 SSE 服务
	sseSvc := sse.NewSSE(ps, "gitfox-sse-namespace")

	// 模拟的一个空间 ID (例如项目 ID)
	var spaceID int64 = 1001

	// 创建一个上下文，用于控制生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// --- 模拟消费者 (前端/客户端) ---
	go func() {
		fmt.Printf("[Client] 开始订阅 SpaceID: %d 的事件...\n", spaceID)

		// 调用 Stream 方法订阅事件流
		// 它返回事件通道、错误通道和清理函数
		events, errs, cleanup := sseSvc.Stream(ctx, spaceID)

		// 确保退出时清理订阅资源
		defer func() {
			fmt.Println("[Client] 停止订阅，清理资源...")
			cleanup(context.Background())
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case err := <-errs:
				log.Printf("[Client] 接收错误: %v\n", err)
				return
			case evt := <-events:
				// 模拟前端收到 SSE 数据后的处理
				fmt.Printf("[Client] 收到事件 -> Type: %s, Data: %s\n", evt.Type, string(evt.Data))
			}
		}
	}()

	// 给消费者一点时间建立订阅 (在内存实现中其实很快，但在真实网络环境可能需要)
	time.Sleep(100 * time.Millisecond)

	// --- 模拟生产者 (后端业务逻辑) ---
	fmt.Println("[Server] 开始发布事件...")

	// 事件 1: PR 更新 (模拟初始状态)
	prData := map[string]interface{}{
		"id":    1,
		"title": "Fix login bug",
		"user":  "alice",
	}
	// 使用已存在的事件类型 eventModel.SSETypePullRequestUpdated
	err := sseSvc.Publish(ctx, spaceID, eventModel.SSETypePullRequestUpdated, prData)
	if err != nil {
		log.Printf("发布失败: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 事件 2: PR 更新
	updateData := map[string]interface{}{
		"id":     1,
		"status": "merged",
	}
	err = sseSvc.Publish(ctx, spaceID, eventModel.SSETypePullRequestUpdated, updateData)
	if err != nil {
		log.Printf("发布失败: %v\n", err)
	}
	time.Sleep(500 * time.Millisecond)

	// 事件 3: 发送给另一个 Space (消费者不应该收到)
	otherSpaceID := int64(9999)
	fmt.Printf("[Server] 发布事件到另一个 SpaceID: %d (预期客户端收不到)...\n", otherSpaceID)
	sseSvc.Publish(ctx, otherSpaceID, eventModel.SSETypePullRequestUpdated, map[string]string{"msg": "secret"})
	time.Sleep(500 * time.Millisecond)

	fmt.Println("[Server] 演示结束")
}
