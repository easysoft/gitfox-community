package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/queue"
)

func main() {
	log := slog.Default()

	// 使用 pkg/database 创建内存 SQLite 连接
	cfg := config.SQLiteConfig{
		Path: ":memory:", // 使用内存数据库
	}

	dbWrapper, cleanup, err := database.NewSQLiteDB(cfg)
	if err != nil {
		log.Error("数据库连接失败", "error", err)
		return
	}
	defer func() {
		if err := cleanup(); err != nil {
			log.Error("关闭数据库连接失败", "error", err)
		}
	}()

	// 获取原生 sql.DB 用于队列系统
	db, err := dbWrapper.DB.DB()
	if err != nil {
		log.Error("获取原生数据库连接失败", "error", err)
		return
	}

	// 设置数据库架构
	schema, err := os.ReadFile("schema_sqlite.sql")
	if err != nil {
		log.Error("读取 schema 文件失败", "error", err)
		return
	}

	if _, err := db.Exec(string(schema)); err != nil {
		log.Error("执行 schema 失败", "error", err)
		return
	}

	// 创建队列
	q := queue.New(queue.NewOpts{
		DB:   db,
		Name: "example",
	})

	ctx := context.Background()

	// 发送消息
	log.Info("发送消息到队列")
	id, err := q.SendAndGetID(ctx, queue.Message{Body: []byte("hello world")})
	if err != nil {
		log.Error("发送消息失败", "error", err)
		return
	}
	log.Info("消息已发送", "id", id)

	// 接收消息
	log.Info("从队列接收消息")
	m, err := q.Receive(ctx)
	if err != nil {
		log.Error("接收消息失败", "error", err)
		return
	}

	if m == nil {
		log.Info("队列中没有消息")
		return
	}

	log.Info("收到消息", "id", m.ID, "body", string(m.Body))

	// 扩展消息超时
	log.Info("扩展消息超时")
	if err := q.Extend(ctx, m.ID, 10*time.Second); err != nil {
		log.Error("扩展超时失败", "error", err)
		return
	}
	log.Info("消息超时已扩展")

	// 删除消息
	log.Info("删除消息")
	if err := q.Delete(ctx, m.ID); err != nil {
		log.Error("删除消息失败", "error", err)
		return
	}
	log.Info("消息已删除")

	log.Info("演示完成")
}
