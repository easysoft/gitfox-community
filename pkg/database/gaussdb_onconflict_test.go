//go:build gaussdbverify

package database

import (
	"os"
	"testing"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/config"
	"gorm.io/gorm/clause"
)

type verifyScheduleJob struct {
	ID         int64   `gorm:"column:id;primaryKey;autoIncrement"`
	Name       string  `gorm:"column:name"`
	Type       string  `gorm:"column:type"`
	State      string  `gorm:"column:state"`
	EditedDate *string `gorm:"column:editedDate"`
}

func (verifyScheduleJob) TableName() string { return "ops_schedule_jobs" }

// TestNewGaussDB_OnConflictUPSERT 验证修复后的 ON CONFLICT 在 GaussDB 上正常 upsert。
// 运行：GITFOX_GAUSSDB_DSN='host=... dbname=zentaopg2 ...' go test -tags gaussdbverify -run TestNewGaussDB_OnConflictUPSERT ./pkg/database/
func TestNewGaussDB_OnConflictUPSERT(t *testing.T) {
	dsn := os.Getenv("GITFOX_GAUSSDB_DSN")
	if dsn == "" {
		t.Fatal("GITFOX_GAUSSDB_DSN not set")
	}
	conf := config.DatabaseConfig{
		Driver:          "gaussdb",
		Datasource:      dsn,
		PrintSQL:        true,
		SlowThreshold:   0,
		MaxIdleConn:     5,
		MaxOpenConn:     10,
		ConnMaxLifeTime: 3600,
	}
	logConf := config.LoggingConfig{Stdout: true, LogDir: t.TempDir()}

	db, cleanup, err := NewGaussDB(conf, logConf)
	if err != nil {
		t.Fatalf("NewGaussDB failed: %v", err)
	}
	defer cleanup()

	// 清理
	db.DB().Exec("DELETE FROM ops_schedule_jobs WHERE name = 'test:upsert'")

	// 首次插入
	job1 := verifyScheduleJob{Name: "test:upsert", Type: "gitfox:test", State: "pending"}
	err = db.DB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "state", "editedDate"}),
	}).Create(&job1).Error
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}
	t.Logf("first insert OK, id=%d", job1.ID)

	// 二次 upsert（应更新而非重复插入）
	job2 := verifyScheduleJob{Name: "test:upsert", Type: "gitfox:test", State: "running"}
	err = db.DB().Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "state", "editedDate"}),
	}).Create(&job2).Error
	if err != nil {
		t.Fatalf("second upsert failed: %v", err)
	}
	if job2.ID != job1.ID {
		t.Errorf("upsert should keep same id, got %d want %d", job2.ID, job1.ID)
	}
	t.Logf("second upsert OK, id=%d (unchanged)", job2.ID)

	// 验证状态已更新
	var s verifyScheduleJob
	if err := db.DB().Where("name = ?", "test:upsert").First(&s).Error; err != nil {
		t.Fatalf("read back failed: %v", err)
	}
	if s.State != "running" {
		t.Errorf("expected state=running after upsert, got %q", s.State)
	}

	db.DB().Exec("DELETE FROM ops_schedule_jobs WHERE name = 'test:upsert'")
	t.Log("ON CONFLICT upsert verified OK")
}
