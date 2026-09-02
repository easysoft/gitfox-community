package repo

import (
	"context"
	"fmt"
	"strings"
	"testing"

	repoModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/repo"
	spaceStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/space"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&Repository{}, &spaceStore.Space{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &Store{db: db}
}

// TestStoreListJoinsOpsSpace 验证 List 联查 ops_space 后，
// Path 由 space.code 与代码库名称拼接生成（service 层基于 Path 生成 gitURL）。
func TestStoreListJoinsOpsSpace(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sp := spaceStore.Space{ID: 1, Name: "团队空间", Code: "team1"}
	if err := store.db.WithContext(ctx).Create(&sp).Error; err != nil {
		t.Fatalf("seed space failed: %v", err)
	}

	row := Repository{
		SpaceID:    sp.ID,
		Identifier: "repo1",
		SCMType:    "git",
		Status:     "active",
	}
	if err := store.db.WithContext(ctx).Create(&row).Error; err != nil {
		t.Fatalf("seed repo failed: %v", err)
	}

	pageResult, err := store.List(ctx, repoModel.ReqFilter{Page: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}

	if pageResult.Total != 1 {
		t.Fatalf("expected total 1, got %d", pageResult.Total)
	}
	got := pageResult.Data[0]
	if got.Path != "team1/repo1" {
		t.Fatalf("expected path %q, got %q", "team1/repo1", got.Path)
	}
}
