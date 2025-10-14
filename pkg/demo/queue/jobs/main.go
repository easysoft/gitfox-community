package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/queue"
)

func main() {
	log := slog.Default()

	// Setup the db
	db, err := sql.Open("sqlite3", ":memory:?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		log.Info("Error opening db", "error", err)
		return
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// Setup the schema
	schema, err := os.ReadFile("schema_sqlite.sql")
	if err != nil {
		log.Info("Error reading schema:", "error", err)
		return
	}

	if _, err := db.Exec(string(schema)); err != nil {
		log.Info("Error executing schema:", "error", err)
		return
	}

	// 为作业创建一个新队列。你可以想要多少就有多少，只是取不同的名字。
	q := queue.New(queue.NewOpts{
		DB:   db,
		Name: "jobs",
	})

	// 创建一个作业运行器，作业限制为1，消息轮询间隔为10毫秒。
	r := queue.NewRunner(queue.NewRunnerOpts{
		Limit:        1,
		Log:          log,
		PollInterval: 10 * time.Millisecond,
		Queue:        q,
	})

	// 注册我们的“print”作业。
	r.Register("print", func(ctx context.Context, m []byte) error {
		fmt.Println(string(m))
		return nil
	})

	// 创建一个“print”作业，消息为“Yo”。
	if _, err := queue.Create(context.Background(), q, "print", queue.Message{Body: []byte("Yo")}); err != nil {
		log.Info("Error creating job", "error", err)
	}

	// 停止作业运行器，等待作业完成。
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	// 启动作业运行器并查看作业运行。
	r.Start(ctx)
}
