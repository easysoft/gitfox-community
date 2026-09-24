// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      mazhiyuan <mazhiyuan@gitfox.com>

package provider

import (
	"fmt"

	"github.com/google/wire"

	"github.com/easysoft/gitfox-community/pkg/config"
	"github.com/easysoft/gitfox-community/pkg/database"
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

	switch driver {
	case "mysql":
		var mysqlDB *database.MysqlDB
		mysqlDB, cleanup, err = database.NewMysqlDB(cfg.Database, cfg.Logging)
		if mysqlDB != nil {
			db = mysqlDB.DB()
		}
	case "dm":
		var dmDB *database.DaMengDB
		dmDB, cleanup, err = database.NewDaMengDB(cfg.Database, cfg.Logging)
		if dmDB != nil {
			db = dmDB.DB()
		}
	case "gaussdb":
		var gaussDB *database.GaussDB
		gaussDB, cleanup, err = database.NewGaussDB(cfg.Database, cfg.Logging)
		if gaussDB != nil {
			db = gaussDB.DB()
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

	db, cleanup, err := database.NewSQLiteDB(config.SQLite, config.Logging)
	if err != nil {
		systemLog.Errorw("[Provider] Failed to create SQLite connection", "error", err)
		return nil, nil, err
	}
	systemLog.Info("[Provider] SQLite connection established successfully")

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
