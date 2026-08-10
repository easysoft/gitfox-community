package database

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	dm "gitfox.zcorp.cc/git/Pangu/dm-gorm"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// DaMengDB 是 gorm.DB 的命名类型，用于达梦数据库
// 使用命名类型而非类型别名，以便 wire 可以区分不同的数据库类型
type DaMengDB gorm.DB

// https://gitee.com/chunanyong/dm#https://gitee.com/link?target=https%3A%2F%2Fpackage.dameng.com%2Feco%2Fadapter
func NewDaMengDB(dbConf config.DatabaseConfig, logConf config.LoggingConfig) (*DaMengDB, func() error, error) {
	gormWriter := buildGormWriter(logConf, true)
	logLevel := logger.Warn
	if dbConf.PrintSQL {
		logLevel = logger.Info
	}
	slowThreshold := dbConf.SlowThreshold
	if slowThreshold <= 0 {
		slowThreshold = 500 * time.Millisecond
	}
	db, err := gorm.Open(dm.Open("dm://SYSDBA:0831will.DB@127.0.0.1:9906"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ops_", // 表名前缀，"repo" 的表应为 "ops_repo"
			SingularTable: true,   // 使用单数表名，启用此选项时，"repo" 的表将为 "repo"
		},
		Logger: newGormLogger(gormWriter, logger.Config{
			SlowThreshold:             slowThreshold, // 慢 SQL 阈值
			LogLevel:                  logLevel,
			Colorful:                  false, // 禁用彩色打印
			IgnoreRecordNotFoundError: false, // 忽略ErrRecordNotFound（记录未找到）错误
		}),
	})

	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, nil, err
	}
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(dbConf.MaxIdleConn)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(dbConf.MaxOpenConn)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Second * dbConf.ConnMaxLifeTime)

	// SetConnMaxIdleTime 设置连接最长空闲时间。
	//sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxLifeTime) * time.Second)

	// 返回清理函数
	cleanup := func() error {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}

		if err := sqlDB.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}

		return nil
	}

	return (*DaMengDB)(db), cleanup, nil
}

// DB 返回底层的 *gorm.DB 实例
func (d *DaMengDB) DB() *gorm.DB {
	return (*gorm.DB)(d)
}
