package provider

import (
	"context"
	"log"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/migrate"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

var DBWireSet = wire.NewSet(
	NewMysqlDB,
	NewSqliteDB,
)

func NewMysqlDB(config *config.Config) (*database.MysqlDB, func(), error) {
	log.Println("[Provider] Initializing MySQL database...")

	db, cleanup, err := database.NewMysqlDB(config.Database)
	if err != nil {
		log.Printf("[Provider] Failed to create MySQL connection: %v", err)
		return nil, nil, err
	}
	log.Println("[Provider] MySQL connection established successfully")

	// 执行数据库迁移
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("[Provider] Failed to get underlying sql.DB: %v", err)
		_ = cleanup()
		return nil, nil, err
	}

	log.Println("[Provider] Starting MySQL database migration...")
	ctx := context.Background()
	if err := migrate.Migrate(ctx, sqlDB, "mysql"); err != nil {
		log.Printf("[Provider] MySQL migration failed: %v", err)
		_ = cleanup()
		return nil, nil, err
	}
	log.Println("[Provider] MySQL database migration completed")

	// 包装 cleanup 函数以忽略错误，匹配 Wire 的 func() 签名
	wrappedCleanup := func() {
		_ = cleanup()
	}
	return db, wrappedCleanup, nil
}

func NewSqliteDB(config *config.Config) (*database.SqliteDB, func(), error) {
	log.Printf("[Provider] Initializing SQLite database at: %s", config.SQLite.Path)

	db, cleanup, err := database.NewSQLiteDB(config.SQLite)
	if err != nil {
		log.Printf("[Provider] Failed to create SQLite connection: %v", err)
		return nil, nil, err
	}
	log.Println("[Provider] SQLite connection established successfully")

	// 执行数据库迁移
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Printf("[Provider] Failed to get underlying sql.DB: %v", err)
		_ = cleanup()
		return nil, nil, err
	}

	log.Println("[Provider] Starting SQLite database migration...")
	ctx := context.Background()
	if err := migrate.Migrate(ctx, sqlDB, "sqlite3"); err != nil {
		log.Printf("[Provider] SQLite migration failed: %v", err)
		_ = cleanup()
		return nil, nil, err
	}
	log.Println("[Provider] SQLite database migration completed")

	// 包装 cleanup 函数以忽略错误，匹配 Wire 的 func() 签名
	wrappedCleanup := func() {
		_ = cleanup()
	}
	return db, wrappedCleanup, nil
}
