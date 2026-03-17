package provider

import (
	"context"
	"fmt"

	"github.com/google/wire"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/migrate"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	"gorm.io/gorm"
)

var DBWireSet = wire.NewSet(
	NewDB,
	// NewSqliteDB,
)

// NewDB 根据配置中的 Driver 类型初始化相应的数据库
func NewDB(cfg *config.Config, systemLog *SystemLog) (*gorm.DB, func(), error) {
	driver := cfg.Database.Driver
	systemLog.Infow("[Provider] Initializing database", "driver", driver)

	var db *gorm.DB
	var cleanup func() error
	var err error
	var dbType string

	switch driver {
	case "mysql":
		dbType = "mysql"
		var mysqlDB *database.MysqlDB
		mysqlDB, cleanup, err = database.NewMysqlDB(cfg.Database)
		if mysqlDB != nil {
			db = mysqlDB.DB()
		}
	case "dm":
		dbType = "dm"
		var dmDB *database.DaMengDB
		dmDB, cleanup, err = database.NewDaMengDB(cfg.Database)
		if dmDB != nil {
			db = dmDB.DB()
		}
	default:
		systemLog.Errorw("[Provider] Unsupported database driver", "driver", driver)
		return nil, nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	if err != nil {
		systemLog.Errorw("[Provider] Failed to create database connection", "driver", driver, "error", err)
		return nil, nil, err
	}
	systemLog.Infow("[Provider] Database connection established successfully", "driver", driver)

	// 执行数据库迁移
	sqlDB, err := db.DB()
	if err != nil {
		systemLog.Errorw("[Provider] Failed to get underlying sql.DB", "driver", driver, "error", err)
		_ = cleanup()
		return nil, nil, err
	}

	systemLog.Infow("[Provider] Starting database migration", "driver", driver)
	ctx := context.Background()
	if err := migrate.Migrate(ctx, sqlDB, dbType, systemLog.SugaredMateLogger); err != nil {
		systemLog.Errorw("[Provider] Database migration failed", "driver", driver, "error", err)
		_ = cleanup()
		return nil, nil, err
	}
	systemLog.Infow("[Provider] Database migration completed", "driver", driver)

	// 包装 cleanup 函数以忽略错误，匹配 Wire 的 func() 签名
	wrappedCleanup := func() {
		if err := cleanup(); err != nil {
			systemLog.Errorw("[Provider] Failed to close database", "driver", driver, "error", err)
		} else {
			systemLog.Infow("[Provider] Database connections closed", "driver", driver)
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
	sqlDB, err := db.DB().DB()
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
