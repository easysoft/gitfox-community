# PHP Client SSE 示例

本目录包含一个完整的示例，演示了如何使用 Go 搭建 SSE (Server-Sent Events) 服务端，并使用 PHP 客户端（基于 cURL）进行实时事件订阅。

## 文件说明

*   `main.go`: Go 服务端。启动 HTTP 服务 (`:8080`)，提供 `/sse` 接口，并持续模拟产生业务事件。
*   `client.php`: PHP 客户端。使用 `curl` 连接 SSE 接口，并通过回调函数实时流式处理接收到的事件。

## 运行步骤

### 1. 启动 Go 服务端

在项目根目录下运行：

```bash
go run app/service/sse/example/php/main.go
```

**服务端输出示例：**

```text
2026/01/15 09:17:16 [Server] HTTP 服务启动在 :8080
2026/01/15 09:17:17 [Business] 准备向 SpaceID: 2024 发送事件 (请运行 PHP 脚本连接)...
2026/01/15 09:17:19 [Business] 触发 'PR 更新' 事件 #1...
...
2026/01/15 09:17:19 [Server] 收到新的 SSE 连接请求，SpaceID: 2024
2026/01/15 09:17:21 [Server] 推送事件 -> Type: pullreq_updated
```

> 注意：前几条日志可能会提示 `no subscribers registered`，这是正常的，因为此时 PHP 客户端尚未连接。

### 2. 启动 PHP 客户端

保持 Go 服务端运行，打开一个新的终端窗口，运行 PHP 脚本：

```bash
php app/service/sse/example/php/client.php
```

**客户端输出示例：**

```text
正在连接 SSE 服务: http://localhost:8080/sse?space_id=2024 ...

[收到事件] 类型: ping
[收到数据] 内容: {"msg": "connected"}

[收到事件] 类型: pullreq_updated
[收到数据] 内容: {"msg":"Message #2 for PHP client","status":"merged","ts":1768439841}

[收到事件] 类型: pullreq_updated
[收到数据] 内容: {"msg":"Message #3 for PHP client","status":"merged","ts":1768439843}
...
```

## 技术要点

1.  **服务端 (Go)**:
    *   使用 `http.Flusher` 确保每条消息写入后立即刷新到网络，防止缓冲。
    *   连接建立时立即发送 `ping` 事件以确认链路通畅。
    *   使用 `for select` 循环监听 PubSub 管道。

2.  **客户端 (PHP)**:
    *   使用 `curl` 扩展连接 SSE 接口。
    *   配置 `CURLOPT_WRITEFUNCTION` 回调函数，实现**流式处理**（边下载边处理），完全绕过 PHP 自身的缓冲区。
    *   设置 `CURLOPT_TIMEOUT` 为 0 以支持长连接。