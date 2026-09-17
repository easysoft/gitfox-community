package statistics

import (
	"context"
	"fmt"
	"strings"
	"testing"

	repoStore "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/store/repo"
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

	if err := db.AutoMigrate(&repoStore.Repository{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &Store{db: db}
}

// TestListGitRepos 验证只返回未删除的git代码库。
func TestListGitRepos(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	repos := []repoStore.Repository{
		{Identifier: "repo1", SCMType: "git", GitUID: "uid1", Status: "active", Deleted: 0},
		{Identifier: "repo2", SCMType: "svn", GitUID: "uid2", Status: "active", Deleted: 0},
		{Identifier: "repo3", SCMType: "git", GitUID: "uid3", Status: "active", Deleted: 1},
		{Identifier: "repo4", SCMType: "git", GitUID: "uid4", Status: "active", Deleted: 0},
	}
	if err := store.db.Create(&repos).Error; err != nil {
		t.Fatalf("seed repos failed: %v", err)
	}

	got, err := store.ListGitRepos(ctx)
	if err != nil {
		t.Fatalf("ListGitRepos returned error: %v", err)
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
}
