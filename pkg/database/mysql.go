package database

import (
	"fmt"

	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// TODO: 数据库的配置之后从 config 结构体读取
type Config struct {
	Username        string
	Password        string
	Address         string
	DBName          string
	DatabaseTZ      string
	Debug           bool
	MaxIdleConn     int
	MaxOpenConn     int
	IsNotPanic      bool
	ConnMaxLifeTime time.Duration
}

type DB struct {
	*gorm.DB
}

func NewMysqlDB(conf Config) *DB {
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=%s",
		conf.Username,
		conf.Password,
		conf.Address,
		conf.DBName, conf.DatabaseTZ) // todo
	// 参考 https://github.com/go-sql-driver/mysql#loc
	// 参考 https://github.com/go-sql-driver/mysql#system-variables
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "ops_", // 表名前缀，“repo” 的表应为 “ops_repo”
			SingularTable: true,   // 使用单数表名，启用此选项时，“repo” 的表将为 “repo”
			NoLowerCase:   false,  // 跳过名称的 snake_casing 格式
		},
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold:             time.Second, // 慢 SQL 阈值
				LogLevel:                  logger.Info, // Log level
				Colorful:                  false,       // 禁用彩色打印
				IgnoreRecordNotFoundError: false,       // 忽略ErrRecordNotFound（记录未找到）错误
			},
		),
	})
	if err != nil && conf.IsNotPanic {
		return nil
	}

	if err != nil {
		panic(err)
	}

	if conf.Debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil && conf.IsNotPanic {
		return nil
	}
	if err != nil {
		panic(err)
	}
	if err = sqlDB.Ping(); err != nil && conf.IsNotPanic {
		return nil
	}
	if err != nil {
		panic(err)
	}
	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(conf.MaxIdleConn)

	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(conf.MaxOpenConn)

	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	// SetConnMaxIdleTime 设置连接最长空闲时间。
	//sqlDB.SetConnMaxIdleTime(time.Duration(conf.ConnMaxLifeTime) * time.Second)
	return &DB{
		db,
	}

}

// CloseMysqlDB 关闭 MySQL 数据库连接
func CloseMysqlDB(db *DB) error {
	if db == nil {
		return fmt.Errorf("database instance is nil")
	}

	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
