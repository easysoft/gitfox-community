# 队列系统演示程序

本目录包含了基于 GitFox 队列系统的三个演示程序，展示了不同层次的队列操作和使用模式。

## 📁 目录结构

```
queue/
├── main.go              # 基础队列操作演示（内存数据库）
├── schema_sqlite.sql    # SQLite 数据库表结构
├── server/             # 生产者程序目录
│   └── main.go         # 消息生产者（发送端）
└── worker/             # 消费者程序目录
    └── main.go         # 消息消费者（处理端）
```

## 🎯 演示程序说明

### 1. 基础队列操作 (`main.go`)

**功能**：演示队列系统的底层 API 使用方式
- 使用内存 SQLite 数据库
- 手动发送、接收、扩展超时、删除消息
- 展示消息锁定机制和超时处理

**运行方式**：
```bash
cd /pkg/demo/queue/queue
go run main.go
```

**特点**：
- 同步操作，适合理解队列基本概念
- 展示完整的消息生命周期
- 包含错误处理和日志记录

### 2. 生产者-消费者架构

#### 生产者程序 (`server/main.go`)

**功能**：模拟消息生产者，持续发送各种类型的任务消息
- 连接到共享 SQLite 数据库文件
- 自动初始化数据库表结构
- 发送 5 种不同类型的任务消息

**支持的任务类型**：
- `email` - 邮件发送任务
- `image_resize` - 图片缩放任务
- `data_sync` - 数据同步任务
- `report_gen` - 报表生成任务
- `cleanup` - 清理任务

**运行方式**：
```bash
cd /pkg/demo/queue/queue/server
go run main.go [数据库路径]
```

**示例**：
```bash
# 使用默认数据库路径
go run main.go

# 指定数据库路径
go run main.go /tmp/my_queue.db
```

#### 消费者程序 (`worker/main.go`)

**功能**：模拟消息消费者，从队列中接收并处理任务消息
- 连接到共享数据库，与生产者协同工作
- 支持多种任务类型的处理逻辑
- 实现优雅关闭和错误恢复

**运行方式**：
```bash
cd /pkg/demo/queue/queue/worker
go run main.go [数据库路径] [工作器ID]
```

**示例**：
```bash
# 使用默认配置
go run main.go

# 指定数据库和工作器ID
go run main.go /tmp/my_queue.db worker-2
```

## 🚀 快速开始

### 场景1：基础队列操作体验

```bash
cd /pkg/demo/queue/queue
go run main.go
```

### 场景2：生产者-消费者模式

**终端1 - 启动生产者**：
```bash
cd /pkg/demo/queue/queue/server
go run main.go
```

**终端2 - 启动消费者**：
```bash
cd /pkg/demo/queue/queue/worker
go run main.go
```

### 场景3：多消费者负载均衡

**终端1 - 启动生产者**：
```bash
cd /pkg/demo/queue/queue/server
go run main.go shared_queue.db
```

**终端2 - 启动消费者1**：
```bash
cd /pkg/demo/queue/queue/worker
go run main.go shared_queue.db worker-1
```

**终端3 - 启动消费者2**：
```bash
cd /pkg/demo/queue/queue/worker
go run main.go shared_queue.db worker-2
```

## 🔧 技术特性

### 数据库支持
- **SQLite**：适合单机部署和开发测试
- **共享文件数据库**：支持多进程访问
- **WAL 模式**：提高并发性能
- **连接池**：优化数据库连接管理

### 队列特性
- **消息持久化**：消息存储在数据库中，重启不丢失
- **消息锁定**：防止消息被多个消费者重复处理
- **超时机制**：处理超时的消息可被重新接收
- **优先级支持**：高优先级消息优先处理
- **重试机制**：失败的消息可以重新处理

### 可靠性保证
- **事务支持**：确保消息操作的原子性
- **优雅关闭**：支持 SIGINT/SIGTERM 信号处理
- **错误恢复**：自动重连和错误重试
- **日志记录**：详细的操作日志和错误信息

## 📊 消息格式

### 邮件任务
```json
{
  "type": "email",
  "to": "user@example.com",
  "subject": "通知标题",
  "body": "邮件内容"
}
```

### 图片缩放任务
```json
{
  "type": "image_resize",
  "file": "image.jpg",
  "width": 800,
  "height": 600
}
```

### 数据同步任务
```json
{
  "type": "data_sync",
  "source": "api",
  "target": "database",
  "batch_id": 123
}
```

### 报表生成任务
```json
{
  "type": "report_gen",
  "report_type": "daily",
  "date": "2025-01-01",
  "user_id": 456
}
```

### 清理任务
```json
{
  "type": "cleanup",
  "target": "temp_files",
  "older_than": "24h",
  "batch": 789
}
```

## 🎓 学习路径

1.  **入门**：运行 `main.go`，理解队列基本概念
2.  **进阶**：运行生产者-消费者程序，体验分布式处理
3.  **实践**：启动多个消费者，观察负载均衡效果
4.  **扩展**：修改任务类型，添加自定义处理逻辑

## 🔍 监控和调试

### 日志输出
程序使用结构化日志，包含以下信息：
- 消息ID和类型
- 工作器ID
- 处理时间
- 错误信息

### 数据库查询
可以直接查询数据库了解队列状态：
```sql
-- 查看队列中的消息
SELECT * FROM queue WHERE queue = 'tasks';

-- 查看消息统计
SELECT queue, COUNT(*) as count FROM queue GROUP BY queue;
```

## 🚨 注意事项

1.  **数据库路径**：确保生产者和消费者使用相同的数据库路径
2.  **并发限制**：SQLite 在高并发场景下性能有限
3.  **错误处理**：生产环境需要更完善的错误处理和监控
4.  **资源清理**：长时间运行需要定期清理已处理的消息

## 🔗 相关文档

- [队列系统核心代码](../../queue/)
- [作业系统演示](../jobs/)
- [GitFox 项目文档](../../../../README.md)