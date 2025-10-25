package provider

import (
	"context"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/migrate"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
)

var DBWireSet = wire.NewSet(
	NewMysqlDB,
	NewSqliteDB,
)

func NewMysqlDB(config *config.Config, systemLog *SystemLog) (*database.MysqlDB, func(), error) {
	systemLog.Info("[Provider] Initializing MySQL database...")

	db, cleanup, err := database.NewMysqlDB(config.Database)
	if err != nil {
		systemLog.Errorw("[Provider] Failed to create MySQL connection", "error", err)
		return nil, nil, err
	}
	systemLog.Info("[Provider] MySQL connection established successfully")

	// 执行数据库迁移
	sqlDB, err := db.DB.DB()
	if err != nil {
		systemLog.Errorw("[Provider] Failed to get underlying sql.DB", "error", err)
		_ = cleanup()
		return nil, nil, err
	}

	systemLog.Info("[Provider] Starting MySQL database migration...")
	ctx := context.Background()
	if err := migrate.Migrate(ctx, sqlDB, "mysql", systemLog.SugaredMateLogger); err != nil {
		systemLog.Errorw("[Provider] MySQL migration failed", "error", err)
		_ = cleanup()
		return nil, nil, err
	}
	systemLog.Info("[Provider] MySQL database migration completed")

	// 包装 cleanup 函数以忽略错误，匹配 Wire 的 func() 签名
	wrappedCleanup := func() {
		if err := cleanup(); err != nil {
			systemLog.Errorw("[Provider] Failed to close MySQL database", "error", err)
		} else {
			systemLog.Info("[Provider] MySQL database connections closed")
		}
	}
	return db, wrappedCleanup, nil
}

func NewSqliteDB(config *config.Config, systemLog *SystemLog) (*database.SqliteDB, func(), error) {
	systemLog.Infow("[Provider] Initializing SQLite database", "path", config.SQLite.Path)

	db, cleanup, err := database.NewSQLiteDB(config.SQLite)
	if err != nil {
		systemLog.Errorw("[Provider] Failed to create SQLite connection", "error", err)
		return nil, nil, err
	}
	systemLog.Info("[Provider] SQLite connection established successfully")

	// 执行数据库迁移
	sqlDB, err := db.DB.DB()
	if err != nil {
		systemLog.Errorw("[Provider] Failed to get underlying sql.DB", "error", err)
		_ = cleanup()
		return nil, nil, err
	}

	systemLog.Info("[Provider] Starting SQLite database migration...")
	ctx := context.Background()
	if err := migrate.Migrate(ctx, sqlDB, "sqlite3", systemLog.SugaredMateLogger); err != nil {
		systemLog.Errorw("[Provider] SQLite migration failed", "error", err)
		_ = cleanup()
		return nil, nil, err
	}
	systemLog.Info("[Provider] SQLite database migration completed")

	// 包装 cleanup 函数以忽略错误，匹配 Wire 的 func() 签名
	wrappedCleanup := func() {
		if err := cleanup(); err != nil {
			systemLog.Errorw("[Provider] Failed to close SQLite database", "error", err)
		} else {
			systemLog.Info("[Provider] SQLite database connections closed")
		}
	}
	return db, wrappedCleanup, nil
}
