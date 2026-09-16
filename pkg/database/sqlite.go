package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/easysoft/gitfox-community/pkg/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// SqliteDB 是 gorm.DB 的命名类型，用于 SQLite 数据库
// 使用命名类型而非类型别名，以便 wire 可以区分不同的数据库类型
type SqliteDB gorm.DB

// NewSQLiteDB 创建 SQLite 数据库连接
func NewSQLiteDB(conf config.SQLiteConfig, logConf config.LoggingConfig) (*SqliteDB, func() error, error) {
	// 确保数据库目录存在
	dir := filepath.Dir(conf.Path)
	// 检查目录是否可以创建
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, err
	}
	// 额外检查：尝试在目录中创建一个临时文件来验证写权限
	tempFile := filepath.Join(dir, ".write_test")
	if f, err := os.Create(tempFile); err != nil {
		return nil, nil, err
	} else {
		f.Close()
		os.Remove(tempFile) // 清理临时文件
	}
	// 构建 DSN
	dsn := buildSQLiteDSN(conf)

	gormWriter := buildGormWriter(logConf, false)
	// 配置 GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,  // 使用单数表名
			NoLowerCase:   false, // 使用 snake_case 命名
		},
		Logger: newGormLogger(gormWriter, logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			Colorful:                  false,
			IgnoreRecordNotFoundError: false,
		}),
	}
	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		return nil, nil, err
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	// 测试连接
	if err = sqlDB.Ping(); err != nil {
		return nil, nil, err
	}

	// 设置连接池参数
	// 注意：SQLite 通常不需要太多并发连接，因为它是文件数据库
	maxIdleConn := 1 // SQLite 默认建议使用较少的连接
	maxOpenConn := 1 // SQLite 在高并发下建议限制连接数
	connMaxLifeTime := time.Hour

	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetMaxOpenConns(maxOpenConn)
	sqlDB.SetConnMaxLifetime(connMaxLifeTime)

	// 执行 SQLite 特定的优化设置
	if err := configureSQLite(db, conf); err != nil {
		panic(fmt.Errorf("failed to configure SQLite: %w", err))
	}

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

	return (*SqliteDB)(db), cleanup, nil
}

// DB 返回底层的 *gorm.DB 实例
func (s *SqliteDB) DB() *gorm.DB {
	return (*gorm.DB)(s)
}

// buildSQLiteDSN 构建 SQLite DSN
func buildSQLiteDSN(conf config.SQLiteConfig) string {
	dsn := conf.Path

	// 添加查询参数
	params := make(map[string]string)
	params["cache"] = "shared"
	params["_journal_mode"] = "WAL"
	params["_synchronous"] = "NORMAL"
	params["_foreign_keys"] = "on" // 启用外键约束

	// 构建完整的 DSN
	if len(params) > 0 {
		dsn += "?"
		first := true
		for key, value := range params {
			if !first {
				dsn += "&"
			}
			dsn += fmt.Sprintf("%s=%s", key, value)
			first = false
		}
	}

	return dsn
}

// configureSQLite 配置 SQLite 特定设置
func configureSQLite(db *gorm.DB, conf config.SQLiteConfig) error {
	// 设置 PRAGMA 参数来优化 SQLite 性能
	pragmas := []string{
		"PRAGMA foreign_keys = ON",         // 启用外键约束
		"PRAGMA temp_store = MEMORY",       // 临时表存储在内存中
		"PRAGMA cache_size = -64000",       // 设置缓存大小为 64MB
		"PRAGMA busy_timeout = 30000",      // 设置忙等待超时为 30 秒
		"PRAGMA journal_mode = WAL",        // 启用 Write-Ahead Logging (WAL) 模式
		"PRAGMA wal_autocheckpoint = 1000", // 当 WAL 文件达到 1000 页时,自动执行检查点操作(将 WAL 中的数据合并回主数据库文件)
		"PRAGMA synchronous = NORMAL",      // 在 WAL 模式下足够安全, 推荐用于 WAL 模式
	}

	// 执行所有 PRAGMA 语句（使用 Silent 模式避免打印日志）
	silentDB := db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
	for _, pragma := range pragmas {
		if err := silentDB.Exec(pragma).Error; err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}
