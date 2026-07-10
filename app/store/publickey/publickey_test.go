package publickey

import (
	"context"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/database"
	pkgErrors "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&PublicKey{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &Store{db: db}
}

func TestStoreGetByID(t *testing.T) {
	t.Parallel()

	store := newTestStore(t)
	ctx := context.Background()

	record := PublicKey{
		Account:    "tester",
		Identifier: "key-1",
		Content:    database.NoNull("ssh-ed25519 AAAA"),
		Type:       "ssh-ed25519",
	}
	if err := store.db.WithContext(ctx).Create(&record).Error; err != nil {
		t.Fatalf("seed public key failed: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		got, err := store.GetByID(ctx, "tester", record.ID)
		if err != nil {
			t.Fatalf("GetByID returned error: %v", err)
		}

		if got.ID != record.ID {
			t.Fatalf("expected id %d, got %d", record.ID, got.ID)
		}
		if got.Identifier != "key-1" {
			t.Fatalf("expected identifier key-1, got %s", got.Identifier)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := store.GetByID(ctx, "tester", record.ID+1000)
		if !pkgErrors.IsNotFound(err) {
			t.Fatalf("expected not found error, got: %v", err)
		}
	})

	t.Run("account mismatch", func(t *testing.T) {
		_, err := store.GetByID(ctx, "other", record.ID)
		if !pkgErrors.IsNotFound(err) {
			t.Fatalf("expected not found for account mismatch, got: %v", err)
		}
	})
}
