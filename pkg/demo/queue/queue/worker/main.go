package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/queue"
	"gorm.io/gorm/logger"
)

// 任务消息结构
type TaskMessage struct {
	Type       string `json:"type"`
	To         string `json:"to,omitempty"`
	Subject    string `json:"subject,omitempty"`
	Body       string `json:"body,omitempty"`
	File       string `json:"file,omitempty"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Source     string `json:"source,omitempty"`
	Target     string `json:"target,omitempty"`
	BatchID    int    `json:"batch_id,omitempty"`
	ReportType string `json:"report_type,omitempty"`
	Date       string `json:"date,omitempty"`
	UserID     int    `json:"user_id,omitempty"`
	OlderThan  string `json:"older_than,omitempty"`
	Batch      int    `json:"batch,omitempty"`
}

func main() {
	log := slog.Default()

	// 从命令行参数获取数据库路径和工作器ID
	dbPath := "../schema_sqlite.db"
	workerID := "worker-1"

	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	if len(os.Args) > 2 {
		workerID = os.Args[2]
	}

	log.Info("消费者启动", "database", dbPath, "worker_id", workerID)

	config := database.SQLiteConfig{
		DBPath:          dbPath,
		WALMode:         true,
		MaxIdleConn:     5,
		MaxOpenConn:     10,
		IsNotPanic:      true,
		ConnMaxLifeTime: time.Hour,
		LogLevel:        logger.Silent,
	}

	dbWrapper := database.NewSQLiteDB(config)
	if dbWrapper == nil {
		log.Error("数据库连接失败")
		return
	}

	// 获取原生 sql.DB 用于队列系统
	db, err := dbWrapper.DB.DB()
	if err != nil {
		log.Error("获取原生数据库连接失败", "error", err)
		return
	}
	defer db.Close()

	// 初始化数据库表结构（如果不存在）
	if err := initDatabase(db); err != nil {
		log.Error("数据库初始化失败", "error", err)
		return
	}

	// 创建队列
	q := queue.New(queue.NewOpts{
		DB:   db,
		Name: "tasks",
	})

	// 设置优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("收到关闭信号，正在优雅关闭...", "worker_id", workerID)
		cancel()
	}()

	log.Info("开始消费消息...", "worker_id", workerID)

	processedCount := 0

	// 消费者主循环
	for {
		select {
		case <-ctx.Done():
			log.Info("消费者关闭", "worker_id", workerID, "processed_total", processedCount)
			return

		default:
			// 接收消息（等待最多2秒）
			message, err := q.ReceiveAndWait(ctx, 2*time.Second)
			if err != nil {
				if err == context.Canceled || err == context.DeadlineExceeded {
					continue
				}
				log.Error("接收消息失败", "error", err, "worker_id", workerID)
				time.Sleep(1 * time.Second)
				continue
			}

			if message == nil {
				continue // 没有消息，继续轮询
			}

			// 处理消息
			if err := processMessage(message, workerID, log); err != nil {
				log.Error("处理消息失败", "error", err, "message_id", message.ID, "worker_id", workerID)
				continue
			}

			// 扩展消息超时（模拟长时间处理）
			if err := q.Extend(ctx, message.ID, 30*time.Second); err != nil {
				log.Warn("扩展消息超时失败", "error", err, "message_id", message.ID, "worker_id", workerID)
			}

			// 删除已处理的消息
			if err := q.Delete(ctx, message.ID); err != nil {
				log.Error("删除消息失败", "error", err, "message_id", message.ID, "worker_id", workerID)
			} else {
				processedCount++
				log.Info("消息处理完成", "message_id", message.ID, "worker_id", workerID, "processed_count", processedCount)
			}
		}
	}
}

// 初始化数据库表结构
func initDatabase(db *sql.DB) error {
	// 读取 schema 文件
	schemaSQL, err := os.ReadFile("../schema_sqlite.sql")
	if err != nil {
		return fmt.Errorf("读取 schema 文件失败: %w", err)
	}

	// 执行 schema
	if _, err := db.Exec(string(schemaSQL)); err != nil {
		return fmt.Errorf("执行 schema 失败: %w", err)
	}

	return nil
}

// 处理消息
func processMessage(message *queue.Message, workerID string, log *slog.Logger) error {
	log.Info("开始处理消息",
		"message_id", message.ID,
		"worker_id", workerID)

	// 解析消息
	var task TaskMessage
	if err := json.Unmarshal(message.Body, &task); err != nil {
		return fmt.Errorf("解析消息失败: %w", err)
	}

	// 根据任务类型处理
	switch task.Type {
	case "email":
		return processEmailTask(&task, workerID, log)
	case "image_resize":
		return processImageResizeTask(&task, workerID, log)
	case "data_sync":
		return processDataSyncTask(&task, workerID, log)
	case "report_gen":
		return processReportGenTask(&task, workerID, log)
	case "cleanup":
		return processCleanupTask(&task, workerID, log)
	default:
		log.Warn("未知任务类型", "type", task.Type, "worker_id", workerID)
		return nil
	}
}

func processEmailTask(task *TaskMessage, workerID string, log *slog.Logger) error {
	log.Info("处理邮件任务",
		"to", task.To,
		"subject", task.Subject,
		"worker_id", workerID)

	// 模拟邮件发送耗时
	time.Sleep(1 * time.Second)

	log.Info("邮件发送成功", "to", task.To, "worker_id", workerID)
	return nil
}

// 处理图片缩放任务
func processImageResizeTask(task *TaskMessage, workerID string, log *slog.Logger) error {
	log.Info("处理图片缩放任务",
		"file", task.File,
		"size", fmt.Sprintf("%dx%d", task.Width, task.Height),
		"worker_id", workerID)

	// 模拟图片处理耗时
	time.Sleep(2 * time.Second)

	log.Info("图片缩放完成", "file", task.File, "worker_id", workerID)
	return nil
}

// 处理数据同步任务
func processDataSyncTask(task *TaskMessage, workerID string, log *slog.Logger) error {
	log.Info("处理数据同步任务",
		"source", task.Source,
		"target", task.Target,
		"batch_id", task.BatchID,
		"worker_id", workerID)

	// 模拟数据同步耗时
	time.Sleep(1500 * time.Millisecond)

	log.Info("数据同步完成", "batch_id", task.BatchID, "worker_id", workerID)
	return nil
}

// 处理报表生成任务
func processReportGenTask(task *TaskMessage, workerID string, log *slog.Logger) error {
	log.Info("处理报表生成任务",
		"type", task.ReportType,
		"date", task.Date,
		"user_id", task.UserID,
		"worker_id", workerID)

	// 模拟报表生成耗时
	time.Sleep(3 * time.Second)

	log.Info("报表生成完成", "type", task.ReportType, "worker_id", workerID)
	return nil
}

// 处理清理任务
func processCleanupTask(task *TaskMessage, workerID string, log *slog.Logger) error {
	log.Info("处理清理任务",
		"target", task.Target,
		"older_than", task.OlderThan,
		"batch", task.Batch,
		"worker_id", workerID)

	// 模拟清理耗时
	time.Sleep(1 * time.Second)

	log.Info("清理任务完成", "target", task.Target, "worker_id", workerID)
	return nil
}
