package publickey

import (
	"testing"

	"gorm.io/gorm"
)

// 这个测试验证 PublicKeyStore 的结构和初始化
func TestNewPublicKeyStore(t *testing.T) {
	// 创建一个 nil DB (仅用于结构验证)
	var db *gorm.DB
	store := NewPublicKeyStore(db)

	if store == nil {
		t.Fatal("NewPublicKeyStore 返回 nil")
	}

	// 验证返回的是正确的类型
	if _, ok := store.(*PublicKeyStore); !ok {
		t.Error("NewPublicKeyStore 返回的类型不正确")
	}

	t.Log("PublicKeyStore 初始化测试通过")
}

// 注意: ListByFingerprint 和 MarkAsVerified 的完整测试需要数据库连接
// 可以使用集成测试或 Docker 容器进行测试
// 或者使用 sqlmock: go get github.com/DATA-DOG/go-sqlmock
