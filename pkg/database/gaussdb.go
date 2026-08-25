package database

import (
	"fmt"
	"time"

	"gorm.io/driver/gaussdb"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
)

// GaussDB 是 gorm.DB 的命名类型，用于 GaussDB 数据库
// 使用命名类型而非类型别名，以便 wire 可以区分不同的数据库类型
type GaussDB gorm.DB

// NewGaussDB 创建 GaussDB 数据库连接
// DSN 格式参考 https://github.com/HuaweiCloudDeveloper/gaussdb-go
// 例如: host=localhost user=gorm password=gorm dbname=gorm port=8000 sslmode=disable TimeZone=Asia/Shanghai
func NewGaussDB(dbConf config.DatabaseConfig, logConf config.LoggingConfig) (*GaussDB, func() error, error) {
	gormWriter := buildGormWriter(logConf, true)
	logLevel := logger.Warn
	if dbConf.PrintSQL {
		logLevel = logger.Info
	}
	slowThreshold := dbConf.SlowThreshold
	if slowThreshold <= 0 {
		slowThreshold = 500 * time.Millisecond
	}
	db, err := gorm.Open(gaussdb.Open(dbConf.Datasource), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ops_", // 表名前缀，"repo" 的表应为 "ops_repo"
			SingularTable: true,   // 使用单数表名，启用此选项时，"repo" 的表将为 "repo"
			NoLowerCase:   true,   // 跳过名称的 snake_casing 格式
		},
		Logger: newGormLogger(gormWriter, logger.Config{
			SlowThreshold:             slowThreshold, // 慢 SQL 阈值
			LogLevel:                  logLevel,
			Colorful:                  false, // 禁用彩色打印
			IgnoreRecordNotFoundError: false, // 忽略ErrRecordNotFound（记录未找到）错误
		}),
		// 覆盖 gorm.io/driver/gaussdb 的 ON CONFLICT 实现 bug：
		// 该驱动将 clause.OnConflict 错误地生成为 MySQL 的
		// "ON DUPLICATE KEY UPDATE ... VALUES(...)" 语法，GaussDB 不支持，
		// 报 "RETURNING clause is not yet supported within INSERT ON DUPLICATE KEY UPDATE"。
		// 这里走 GORM 内置实现，生成标准 GaussDB 语法
		// "ON CONFLICT (...) DO UPDATE SET ... EXCLUDED ..."。
		// 驱动 Initialize 仅在 ClauseBuilders 无对应 key 时注册，故此处预设可覆盖之。
		ClauseBuilders: map[string]clause.ClauseBuilder{
			"ON CONFLICT": func(c clause.Clause, builder clause.Builder) {
				c.Build(builder)
			},
		},
	})

	if err != nil {
		return nil, nil, err
	}

	// 获取底层 *sql.DB（包装前，此时 db.ConnPool 即 *sql.DB）
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	if err = sqlDB.Ping(); err != nil {
		return nil, nil, err
	}

	// 包装连接池：GaussDB 预建表使用驼峰列名（如 isCustom），
	// 未加引号的标识符会被 GaussDB 折叠为小写，导致 Column not found。
	// 通过 SQLRewriter 在 SQL 执行前给驼峰标识符加双引号。
	pool := &RewritingConnPool{DB: sqlDB, Rewriter: GaussDBSQLRewriter{}}
	db.ConnPool = pool
	if db.Statement != nil {
		db.Statement.ConnPool = pool
	}
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(dbConf.MaxIdleConn)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(dbConf.MaxOpenConn)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Second * dbConf.ConnMaxLifeTime)

	// SetConnMaxIdleTime 设置连接最长空闲时间。
	//sqlDB.SetConnMaxIdleTime(time.Duration(dbConf.ConnMaxLifeTime) * time.Second)

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

	return (*GaussDB)(db), cleanup, nil
}

// DB 返回底层的 *gorm.DB 实例
func (g *GaussDB) DB() *gorm.DB {
	return (*gorm.DB)(g)
}
