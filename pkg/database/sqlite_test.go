// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package database

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/easysoft/gitfox-community/pkg/config"
)

// TestUser 测试用的用户模型
type TestUser struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"size:100;not null"`
	Email     string `gorm:"size:100;uniqueIndex"`
	Age       int    `gorm:"default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func TestInitSQLiteDB(t *testing.T) {
	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		t.Fatal("Failed to initialize database")
	}

	// 验证数据库文件是否创建
	if _, err := os.Stat(cfg.Path); os.IsNotExist(err) {
		t.Fatalf("Database file was not created")
	}

	// 测试连接
	db, err := dbWrapper.DB().DB()
	if err != nil {
		t.Fatalf("Failed to unwrapper sql db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	t.Logf("File-based SQLite database connection successful: %s", cfg.Path)
}

func TestSQLiteDB_CRUD_Operations(t *testing.T) {
	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		t.Fatal("Failed to initialize database")
	}
	db := dbWrapper.DB()

	// 自动迁移表结构
	err := db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	// 测试创建
	user := TestUser{
		Name:  "张三",
		Email: "zhangsan@example.com",
		Age:   25,
	}

	result := db.Create(&user)
	if result.Error != nil {
		t.Fatalf("Failed to create user: %v", result.Error)
	}

	if user.ID == 0 {
		t.Fatal("Expected user ID to be set after creation")
	}

	t.Logf("Created user with ID: %d", user.ID)

	// 测试查询
	var foundUser TestUser
	result = db.First(&foundUser, user.ID)
	if result.Error != nil {
		t.Fatalf("Failed to find user: %v", result.Error)
	}

	if foundUser.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, foundUser.Name)
	}

	// 测试更新
	foundUser.Age = 30
	result = db.Save(&foundUser)
	if result.Error != nil {
		t.Fatalf("Failed to update user: %v", result.Error)
	}

	// 验证更新
	var updatedUser TestUser
	db.First(&updatedUser, user.ID)
	if updatedUser.Age != 30 {
		t.Errorf("Expected age 30, got %d", updatedUser.Age)
	}

	// 测试删除
	result = db.Delete(&foundUser)
	if result.Error != nil {
		t.Fatalf("Failed to delete user: %v", result.Error)
	}

	// 验证删除
	var deletedUser TestUser
	result = db.First(&deletedUser, user.ID)
	if result.Error != gorm.ErrRecordNotFound {
		t.Errorf("Expected record not found error, got: %v", result.Error)
	}

	t.Log("CRUD operations completed successfully")
}

func TestSQLiteDB_Transaction(t *testing.T) {

	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		t.Fatal("Failed to initialize database")
	}
	db := dbWrapper.DB()

	// 迁移表结构
	err := db.AutoMigrate(&TestUser{})
	if err != nil {
		t.Fatalf("Failed to migrate table: %v", err)
	}

	// 测试事务
	err = db.Transaction(func(tx *gorm.DB) error {
		// 在事务中创建多个用户
		users := []TestUser{
			{Name: "用户1", Email: "user1@example.com", Age: 20},
			{Name: "用户2", Email: "user2@example.com", Age: 25},
			{Name: "用户3", Email: "user3@example.com", Age: 30},
		}

		for _, user := range users {
			if err := tx.Create(&user).Error; err != nil {
				return err // 这会触发回滚
			}
		}

		return nil // 提交事务
	})

	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// 验证事务结果
	var count int64
	db.Model(&TestUser{}).Count(&count)
	if count != 3 {
		t.Errorf("Expected 3 users, got %d", count)
	}

	t.Log("Transaction test completed successfully")
}

func TestSQLiteDB_ConnectionPool(t *testing.T) {
	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		t.Fatal("Failed to initialize database")
	}

	sqlDB, err := dbWrapper.DB().DB()
	if err != nil {
		t.Fatalf("Failed to get underlying sql.DB: %v", err)
	}

	// 验证连接池设置
	stats := sqlDB.Stats()
	t.Logf("Connection pool stats: MaxOpenConnections=%d",
		stats.MaxOpenConnections)

	// 注意：SQLite 的连接池行为可能与 MySQL 不同
	// 这里主要验证设置不会导致错误
}

func TestSQLiteDB_WALMode(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "wal_test.db")
	cfg := config.SQLiteConfig{Path: dbPath}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		t.Fatal("Failed to initialize database")
	}
	db := dbWrapper.DB()

	// 验证 WAL 模式是否启用
	var journalMode string
	err := db.Raw("PRAGMA journal_mode").Scan(&journalMode).Error
	if err != nil {
		t.Fatalf("Failed to check journal mode: %v", err)
	}

	if journalMode != "wal" {
		t.Errorf("Expected WAL mode, got: %s", journalMode)
	}

	t.Logf("WAL mode enabled successfully: %s", journalMode)
}

func TestSQLiteDB_ErrorHandling(t *testing.T) {
	// 测试无效路径
	cfg := config.SQLiteConfig{Path: "/invalid/path/test.db"}
	db, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if db == nil {
		t.Fatal("Failed to initialize database")
	}

	// 测试 panic 模式
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid path")
		}
	}()

	NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true}) // 这应该会 panic
}

func BenchmarkSQLiteDB_Insert(b *testing.B) {

	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		b.Fatal("Failed to initialize database")
	}
	db := dbWrapper.DB()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		user := TestUser{
			Name:  "BenchUser",
			Email: "bench@example.com",
			Age:   25,
		}
		db.Create(&user)
	}
}

func BenchmarkSQLiteDB_Select(b *testing.B) {

	cfg := config.SQLiteConfig{Path: "./gitfox_test.db"}
	dbWrapper, _, _ := NewSQLiteDB(cfg, config.LoggingConfig{Stdout: true})
	if dbWrapper == nil {
		b.Fatal("Failed to initialize database")
	}
	db := dbWrapper.DB()

	err := db.AutoMigrate(&TestUser{})
	if err != nil {
		b.Fatalf("Failed to migrate table: %v", err)
	}

	// 预先插入一些数据
	for i := 0; i < 100; i++ {
		user := TestUser{
			Name:  "TestUser",
			Email: "test@example.com",
			Age:   25,
		}
		db.Create(&user)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var users []TestUser
		db.Find(&users)
	}
}
