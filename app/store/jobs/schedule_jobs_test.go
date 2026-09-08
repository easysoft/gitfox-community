package jobs

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	jobModel "gitfox.zcorp.cc/git/Pangu/gitfox-community/app/model/jobs"
	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/utils/datetime"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newJobsTestStore(t *testing.T) *Store {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "jobs.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&ScheduleJob{}); err != nil {
		t.Fatalf("auto migrate failed: %v", err)
	}

	return &Store{db: db}
}

func sampleJob(name string) *jobModel.ScheduleJob {
	now := datetime.NowLocalTime()
	scheduled := now.Add(24 * time.Hour)
	return &jobModel.ScheduleJob{
		Name:               name,
		Type:               "gitfox:cleanup:garbage-collector",
		Priority:           jobModel.PriorityElevated,
		Data:               `{"retention_time":0}`,
		Result:             "",
		MaxDurationSeconds: 3600,
		MaxRetries:         0,
		State:              jobModel.StateScheduled,
		Scheduled:          scheduled,
		IsRecurring:        true,
		RecurringCron:      "0 2 * * *",
		CreatedDate:        now,
		EditedDate:         now,
	}
}

func TestStoreUpsertCreatesNewJob(t *testing.T) {
	t.Parallel()

	store := newJobsTestStore(t)
	ctx := context.Background()

	job := sampleJob("gitfox:recurring:test-gc")
	if err := store.Upsert(ctx, job); err != nil {
		t.Fatalf("Upsert returned error: %v", err)
	}

	if job.ID == 0 {
		t.Fatal("expected Upsert to backfill job.ID for a new job")
	}

	var count int64
	if err := store.db.WithContext(ctx).Model(&ScheduleJob{}).Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}

func TestStoreUpsertUpdatesExistingJob(t *testing.T) {
	t.Parallel()

	store := newJobsTestStore(t)
	ctx := context.Background()

	name := "gitfox:recurring:test-gc"
	first := sampleJob(name)
	if err := store.Upsert(ctx, first); err != nil {
		t.Fatalf("first Upsert returned error: %v", err)
	}

	// 模拟该行在运行期被写入非 upsert 列（应被保留）
	if err := store.db.WithContext(ctx).
		Model(&ScheduleJob{}).
		Where("name = ?", name).
		Update("totalExecutions", 5).Error; err != nil {
		t.Fatalf("seed totalExecutions failed: %v", err)
	}

	second := sampleJob(name)
	second.Data = `{"retention_time":720}`
	second.State = jobModel.StateFinished
	second.RecurringCron = "0 3 * * *"
	if err := store.Upsert(ctx, second); err != nil {
		t.Fatalf("second Upsert returned error: %v", err)
	}

	if second.ID != first.ID {
		t.Fatalf("expected same row id %d, got %d", first.ID, second.ID)
	}

	var count int64
	if err := store.db.WithContext(ctx).Model(&ScheduleJob{}).Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row after upsert, got %d", count)
	}

	got, err := store.GetByName(ctx, name)
	if err != nil {
		t.Fatalf("GetByName returned error: %v", err)
	}
	if got.Data != `{"retention_time":720}` {
		t.Fatalf("expected updated data, got %q", got.Data)
	}
	if got.State != jobModel.StateFinished {
		t.Fatalf("expected updated state %q, got %q", jobModel.StateFinished, got.State)
	}
	if got.RecurringCron != "0 3 * * *" {
		t.Fatalf("expected updated recurringCron, got %q", got.RecurringCron)
	}
	if got.TotalExecutions != 5 {
		t.Fatalf("expected totalExecutions preserved as 5, got %d", got.TotalExecutions)
	}
}

func TestStoreUpsertKeepsDistinctNamesSeparate(t *testing.T) {
	t.Parallel()

	store := newJobsTestStore(t)
	ctx := context.Background()

	a := sampleJob("gitfox:recurring:test-a")
	b := sampleJob("gitfox:recurring:test-b")
	if err := store.Upsert(ctx, a); err != nil {
		t.Fatalf("Upsert(a) returned error: %v", err)
	}
	if err := store.Upsert(ctx, b); err != nil {
		t.Fatalf("Upsert(b) returned error: %v", err)
	}

	var count int64
	if err := store.db.WithContext(ctx).Model(&ScheduleJob{}).Count(&count).Error; err != nil {
		t.Fatalf("count failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 rows, got %d", count)
	}
}
