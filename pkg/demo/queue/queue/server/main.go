package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/queue"
	"gorm.io/gorm/logger"
)

func main() {
	log := slog.Default()

	// 从命令行参数获取数据库路径，默认使用文件数据库
	dbPath := "../schema_sqlite.db"
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	log.Info("生产者启动", "database", dbPath)

	config := database.SQLiteConfig{
		DBPath:          dbPath,
		WALMode:         true,
		MaxOpenConn:     10,
		MaxIdleConn:     5,
		IsNotPanic:      true,
		ConnMaxLifeTime: time.Hour,
		LogLevel:        logger.Silent,
	}

	dbWrapper := database.NewSQLiteDB(config)
	if dbWrapper == nil {
		log.Error("数据库连接失败")
		return
	}

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

	log.Info("开始生产消息...")

	// 发送10条不同类型的消息
	for i := 1; i <= 10; i++ {
		taskType := getTaskType(i)
		message := createTaskMessage(taskType, i)

		// 发送消息到队列
		id, err := q.SendAndGetID(context.Background(), queue.Message{
			Body:     []byte(message),
			Priority: i % 3, // 设置不同优先级
		})

		if err != nil {
			log.Error("消息发送失败", "error", err, "count", i)
			continue
		}

		log.Info("消息已发送",
			"id", id,
			"type", taskType,
			"count", i,
			"message", message)

		// 间隔2秒发送下一条消息
		time.Sleep(2 * time.Second)
	}

	log.Info("生产者完成，共发送10条消息")
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

// 根据计数获取任务类型
func getTaskType(count int) string {
	types := []string{"email", "image_resize", "data_sync", "report_gen", "cleanup"}
	return types[(count-1)%len(types)]
}

// 创建任务消息
func createTaskMessage(taskType string, count int) string {
	switch taskType {
	case "email":
		return fmt.Sprintf(`{"type":"email","to":"user%d@example.com","subject":"通知 #%d","body":"这是第%d条邮件通知"}`, count, count, count)
	case "image_resize":
		return fmt.Sprintf(`{"type":"image_resize","file":"image_%d.jpg","width":800,"height":600}`, count)
	case "data_sync":
		return fmt.Sprintf(`{"type":"data_sync","source":"api","target":"database","batch_id":%d}`, count)
	case "report_gen":
		return fmt.Sprintf(`{"type":"report_gen","report_type":"daily","date":"%s","user_id":%d}`, time.Now().Format("2006-01-02"), count)
	case "cleanup":
		return fmt.Sprintf(`{"type":"cleanup","target":"temp_files","older_than":"24h","batch":%d}`, count)
	default:
		return fmt.Sprintf(`{"type":"unknown","data":"message_%d"}`, count)
	}
}
