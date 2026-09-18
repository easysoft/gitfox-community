package repo

import (
	"context"
	"fmt"
	"strings"
	"testing"

	repoModel "github.com/easysoft/gitfox-community/app/model/repo"
	spaceStore "github.com/easysoft/gitfox-community/app/store/space"
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

// TestStoreListAllGitRepos 验证一次性查询只返回未删除的git代码库，且按 ID 升序排列。
func TestStoreListAllGitRepos(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	repos := []Repository{
		{Identifier: "repo1", SCMType: "git", GitUID: "uid1", Status: "active", Deleted: 0},
		{Identifier: "repo2", SCMType: "svn", GitUID: "uid2", Status: "active", Deleted: 0},
		{Identifier: "repo3", SCMType: "git", GitUID: "uid3", Status: "active", Deleted: 1},
		{Identifier: "repo4", SCMType: "git", GitUID: "uid4", Status: "active", Deleted: 0},
	}
	if err := store.db.WithContext(ctx).Create(&repos).Error; err != nil {
		t.Fatalf("seed repos failed: %v", err)
	}

	got, err := store.ListAllGitRepos(ctx)
	if err != nil {
		t.Fatalf("ListAllGitRepos returned error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 repos, got %d(%+v)", len(got), got)
	}
	if got[0].Identifier != "repo1" || got[1].Identifier != "repo4" {
		t.Errorf("unexpected repos: %s, %s", got[0].Identifier, got[1].Identifier)
	}
	if got[0].GitUID != "uid1" || got[1].GitUID != "uid4" {
		t.Errorf("unexpected git uids: %s, %s", got[0].GitUID, got[1].GitUID)
	}
	if got[0].ID >= got[1].ID {
		t.Errorf("expected repos ordered by id asc, got %d, %d", got[0].ID, got[1].ID)
	}
}
