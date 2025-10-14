package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// SQLiteConfig SQLite 数据库配置
type SQLiteConfig struct {
	DBPath          string        // 数据库文件路径
	Debug           bool          // 是否开启调试模式
	MaxIdleConn     int           // 最大空闲连接数
	MaxOpenConn     int           // 最大打开连接数
	IsNotPanic      bool          // 是否不panic，而是返回nil
	ConnMaxLifeTime time.Duration // 连接最大生存时间
	// SQLite 特有配置
	WALMode         bool            // 是否启用 WAL 模式
	CacheMode       string          // 缓存模式: shared, private
	JournalMode     string          // 日志模式: DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF
	SynchronousMode string          // 同步模式: OFF, NORMAL, FULL, EXTRA
	LogLevel        logger.LogLevel // 日志级别
}

// NewSQLiteDB 创建 SQLite 数据库连接
func NewSQLiteDB(conf SQLiteConfig) *DB {
	// 确保数据库目录存在
	if conf.DBPath != ":memory:" {
		dir := filepath.Dir(conf.DBPath)

		// 检查目录是否可以创建
		if err := os.MkdirAll(dir, 0755); err != nil {
			if conf.IsNotPanic {
				return nil
			}
			panic(fmt.Errorf("failed to create database directory: %w", err))
		}

		// 额外检查：尝试在目录中创建一个临时文件来验证写权限
		tempFile := filepath.Join(dir, ".write_test")
		if f, err := os.Create(tempFile); err != nil {
			if conf.IsNotPanic {
				return nil
			}
			panic(fmt.Errorf("no write permission for database directory: %w", err))
		} else {
			f.Close()
			os.Remove(tempFile) // 清理临时文件
		}
	}

	// 构建 DSN
	dsn := buildSQLiteDSN(conf)

	// 配置 GORM
	logLevel := logger.Info
	if conf.LogLevel != 0 {
		logLevel = conf.LogLevel
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,  // 使用单数表名
			NoLowerCase:   false, // 使用 snake_case 命名
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  logLevel,
				Colorful:                  false,
				IgnoreRecordNotFoundError: false,
			},
		),
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dsn), gormConfig)
	if err != nil {
		if conf.IsNotPanic {
			return nil
		}
		panic(fmt.Errorf("failed to connect to SQLite database: %w", err))
	}

	// 开启调试模式
	if conf.Debug {
		db = db.Debug()
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		if conf.IsNotPanic {
			return nil
		}
		panic(fmt.Errorf("failed to get underlying sql.DB: %w", err))
	}

	// 测试连接
	if err = sqlDB.Ping(); err != nil {
		if conf.IsNotPanic {
			return nil
		}
		panic(fmt.Errorf("failed to ping SQLite database: %w", err))
	}

	// 设置连接池参数
	// 注意：SQLite 通常不需要太多并发连接，因为它是文件数据库
	maxIdleConn := conf.MaxIdleConn
	if maxIdleConn <= 0 {
		maxIdleConn = 1 // SQLite 默认建议使用较少的连接
	}

	maxOpenConn := conf.MaxOpenConn
	if maxOpenConn <= 0 {
		maxOpenConn = 1 // SQLite 在高并发下建议限制连接数
	}

	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetMaxOpenConns(maxOpenConn)

	if conf.ConnMaxLifeTime > 0 {
		sqlDB.SetConnMaxLifetime(conf.ConnMaxLifeTime)
	}

	// 执行 SQLite 特定的优化设置
	if err := configureSQLite(db, conf); err != nil {
		if conf.IsNotPanic {
			return nil
		}
		panic(fmt.Errorf("failed to configure SQLite: %w", err))
	}

	return &DB{db}
}

// buildSQLiteDSN 构建 SQLite DSN
func buildSQLiteDSN(conf SQLiteConfig) string {
	dsn := conf.DBPath

	// 添加查询参数
	params := make(map[string]string)

	if conf.CacheMode != "" {
		params["cache"] = conf.CacheMode
	}

	if conf.JournalMode != "" {
		params["_journal_mode"] = conf.JournalMode
	} else if conf.WALMode {
		params["_journal_mode"] = "WAL"
	}

	if conf.SynchronousMode != "" {
		params["_synchronous"] = conf.SynchronousMode
	}

	// 添加其他常用参数
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
func configureSQLite(db *gorm.DB, conf SQLiteConfig) error {
	// 设置 PRAGMA 参数来优化 SQLite 性能
	pragmas := []string{
		"PRAGMA foreign_keys = ON",    // 启用外键约束
		"PRAGMA temp_store = MEMORY",  // 临时表存储在内存中
		"PRAGMA cache_size = -64000",  // 设置缓存大小为 64MB
		"PRAGMA busy_timeout = 30000", // 设置忙等待超时为 30 秒
	}

	// 根据配置添加额外的 PRAGMA
	if conf.WALMode {
		pragmas = append(pragmas, "PRAGMA journal_mode = WAL")
		pragmas = append(pragmas, "PRAGMA wal_autocheckpoint = 1000")
	}

	if conf.SynchronousMode != "" {
		pragmas = append(pragmas, fmt.Sprintf("PRAGMA synchronous = %s", conf.SynchronousMode))
	}

	// 执行所有 PRAGMA 语句
	for _, pragma := range pragmas {
		if err := db.Exec(pragma).Error; err != nil {
			return fmt.Errorf("failed to execute pragma '%s': %w", pragma, err)
		}
	}

	return nil
}

// DefaultSQLiteConfig 返回默认的 SQLite 配置
func DefaultSQLiteConfig(dbPath string) SQLiteConfig {
	return SQLiteConfig{
		DBPath:          dbPath,
		Debug:           false,
		MaxIdleConn:     1,
		MaxOpenConn:     1,
		IsNotPanic:      false,
		ConnMaxLifeTime: time.Hour,
		WALMode:         true,
		CacheMode:       "shared",
		SynchronousMode: "NORMAL",
	}
}

// InMemorySQLiteConfig 返回内存数据库配置
func InMemorySQLiteConfig() SQLiteConfig {
	return SQLiteConfig{
		DBPath:          ":memory:",
		Debug:           true,
		MaxIdleConn:     1,
		MaxOpenConn:     1,
		IsNotPanic:      false,
		ConnMaxLifeTime: 0, // 内存数据库不需要设置连接生存时间
		WALMode:         false,
		CacheMode:       "shared",
		SynchronousMode: "OFF", // 内存数据库可以关闭同步以提高性能
	}
}
