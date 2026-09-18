package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	eventModel "github.com/easysoft/gitfox-community/app/model/event"
	"github.com/easysoft/gitfox-community/app/service/sse"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
)

// 这个示例演示了 Go 服务端如何通过 HTTP 暴露 SSE 接口
// 以及一个模拟的 PHP 客户端如何连接并接收这些事件
func main() {
	// --- 1. 服务端部分：启动 HTTP Server 并暴露 SSE 接口 ---

	// 初始化 PubSub 和 SSE Service
	ps := pubsub.NewMemory(
		pubsub.WithApp("gitfox-php-demo"),
		pubsub.WithNamespace("default"),
		pubsub.WithSendTimeout(time.Second),
		pubsub.WithSize(100),
	)
	sseSvc := sse.NewSSE(ps, "gitfox-sse-namespace")

	// 创建一个 HTTP Handler 处理 SSE 请求
	http.HandleFunc("/sse", func(w http.ResponseWriter, r *http.Request) {
		// 从 Query 参数获取 space_id
		spaceIDStr := r.URL.Query().Get("space_id")
		spaceID, err := strconv.ParseInt(spaceIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid space_id", http.StatusBadRequest)
			return
		}

		log.Printf("[Server] 收到新的 SSE 连接请求，SpaceID: %d\n", spaceID)

		// 设置 SSE 必须的响应头
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		// 跨域设置（如果 PHP 客户端在不同域）
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 刷新头部，让客户端建立连接
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// 订阅事件流
		events, errs, cleanup := sseSvc.Stream(r.Context(), spaceID)
		defer cleanup(context.Background())

		// 连接建立后立即发送一条 ping 消息，确认连接成功
		fmt.Fprintf(w, "event: ping\n")
		fmt.Fprintf(w, "data: {\"msg\": \"connected\"}\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// 循环推送事件
		for {
			select {
			case <-r.Context().Done():
				log.Printf("[Server] 客户端断开连接 SpaceID: %d\n", spaceID)
				return
			case err := <-errs:
				log.Printf("[Server] 内部错误: %v\n", err)
				return
			case event := <-events:
				// 按照 SSE 协议格式化数据
				// event: <type>\n
				// data: <json>\n\n
				fmt.Fprintf(w, "event: %s\n", event.Type)
				fmt.Fprintf(w, "data: %s\n\n", event.Data)

				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
				log.Printf("[Server] 推送事件 -> Type: %s\n", event.Type)
			}
		}
	})

	// 在后台启动 HTTP 服务
	go func() {
		log.Println("[Server] HTTP 服务启动在 :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatal(err)
		}
	}()

	// 等待服务启动
	time.Sleep(500 * time.Millisecond)

	// --- 2. 模拟业务触发事件 ---
	// 启动一个协程，定期发送事件，等待真实的 PHP 客户端连接
	go func() {
		targetSpaceID := int64(2024)
		log.Printf("[Business] 准备向 SpaceID: %d 发送事件 (请运行 PHP 脚本连接)...", targetSpaceID)

		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		count := 0
		for range ticker.C {
			count++
			log.Printf("[Business] 触发 'PR 更新' 事件 #%d...", count)
			err := sseSvc.Publish(context.Background(), targetSpaceID, eventModel.SSETypePullRequestUpdated, map[string]interface{}{
				"status": "merged",
				"msg":    fmt.Sprintf("Message #%d for PHP client", count),
				"ts":     time.Now().Unix(),
			})
			if err != nil {
				log.Printf("[Business] 发布失败: %v", err)
			}
		}
	}()

	// 阻塞主进程
	select {}
}
