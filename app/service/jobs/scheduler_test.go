// @copyright   Copyright 2009-2026 禅道软件（青岛）有限公司(ZenTao Software (Qingdao) Co., Ltd. www.zentao.net)
// @license     ZPL(http://zpl.pub/page/zplv12.html) or AGPL(https://www.gnu.org/licenses/agpl-3.0.en.html)
// @author      lidaogang <lidaogang@chandao.com>

package jobs

import (
	"context"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	jobModel "github.com/easysoft/gitfox-community/app/model/jobs"
	mockStore "github.com/easysoft/gitfox-community/app/store/jobs/mock"
	pkgErrors "github.com/easysoft/gitfox-community/pkg/errors"
	"github.com/easysoft/gitfox-community/pkg/lock"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/pubsub"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// mockPubSub is a minimal mock implementation of pubsub.PubSub for testing
type mockPubSub struct {
	publishCalled   bool
	subscribeCalled bool
}

func (m *mockPubSub) Publish(ctx context.Context, topic string, payload []byte, options ...pubsub.PublishOption) error {
	m.publishCalled = true
	return nil
}

func (m *mockPubSub) Subscribe(ctx context.Context, topic string, handler func(payload []byte) error, options ...pubsub.SubscribeOption) pubsub.Consumer {
	m.subscribeCalled = true
	return &mockConsumer{}
}

type mockConsumer struct{}

func (m *mockConsumer) Subscribe(ctx context.Context, topics ...string) error   { return nil }
func (m *mockConsumer) Unsubscribe(ctx context.Context, topics ...string) error { return nil }
func (m *mockConsumer) Close() error                                            { return nil }

// mockMutexManager is a minimal mock implementation of lock.MutexManager
type mockMutexManager struct{}

func (m *mockMutexManager) NewMutex(key string, options ...lock.Option) (lock.Mutex, error) {
	return &mockMutex{key: key}, nil
}

type mockMutex struct {
	key    string
	locked bool
}

func (m *mockMutex) Key() string {
	return m.key
}

func (m *mockMutex) Lock(ctx context.Context) error {
	m.locked = true
	return nil
}

func (m *mockMutex) Unlock(ctx context.Context) error {
	m.locked = false
	return nil
}

func TestNewScheduler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobStore := mockStore.NewMockInterface(ctrl)
	mockMxManager := &mockMutexManager{}
	mockPubsubService := &mockPubSub{}

	testLog := logger.NewNopLogger()
	scheduler, err := NewScheduler(
		mockJobStore,
		nil,
		mockMxManager,
		mockPubsubService,
		testLog,
		"test-instance",
		5,
		24*time.Hour,
	)

	if err != nil {
		t.Fatalf("NewScheduler failed: %v", err)
	}

	if scheduler == nil {
		t.Fatal("NewScheduler returned nil")
	}

	if scheduler.instanceID != "test-instance" {
		t.Errorf("Expected instanceID 'test-instance', got '%s'", scheduler.instanceID)
	}

	if scheduler.maxRunning != 5 {
		t.Errorf("Expected maxRunning 5, got %d", scheduler.maxRunning)
	}

	if scheduler.retentionTime != 24*time.Hour {
		t.Errorf("Expected retentionTime 24h, got %v", scheduler.retentionTime)
	}

	if scheduler.executor == nil {
		t.Error("Executor not initialized")
	}

	if scheduler.cancelJobMap == nil {
		t.Error("cancelJobMap not initialized")
	}
}

func TestNewScheduler_MinMaxRunning(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobStore := mockStore.NewMockInterface(ctrl)
	mockMxManager := &mockMutexManager{}
	mockPubsubService := &mockPubSub{}

	// Test that maxRunning is set to at least 1
	testLog := logger.NewNopLogger()
	scheduler, err := NewScheduler(
		mockJobStore,
		nil,
		mockMxManager,
		mockPubsubService,
		testLog,
		"test-instance",
		0, // Invalid: should be corrected to 1
		24*time.Hour,
	)

	if err != nil {
		t.Fatalf("NewScheduler failed: %v", err)
	}

	if scheduler.maxRunning != 1 {
		t.Errorf("Expected maxRunning to be corrected to 1, got %d", scheduler.maxRunning)
	}
}

func TestScheduler_addRecurringJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobStore := mockStore.NewMockInterface(ctrl)
	mockMxManager := &mockMutexManager{}
	mockPubsubService := &mockPubSub{}

	testLog := logger.NewNopLogger()
	scheduler, _ := NewScheduler(
		mockJobStore,
		nil,
		mockMxManager,
		mockPubsubService,
		testLog,
		"test-instance",
		5,
		24*time.Hour,
	)

	tests := []struct {
		name        string
		jobUID      string
		jobType     string
		cronDef     string
		maxDur      time.Duration
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid cron expression - every minute",
			jobUID:      "test-job-1",
			jobType:     "test-type",
			cronDef:     "* * * * *",
			maxDur:      5 * time.Second,
			expectError: false,
		},
		{
			name:        "valid cron expression - every 5 minutes",
			jobUID:      "test-job-2",
			jobType:     "test-type",
			cronDef:     "*/5 * * * *",
			maxDur:      10 * time.Second,
			expectError: false,
		},
		{
			name:        "invalid cron expression",
			jobUID:      "test-job-3",
			jobType:     "test-type",
			cronDef:     "invalid cron",
			maxDur:      5 * time.Second,
			expectError: true,
			errorMsg:    "invalid cron definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.expectError {
				// Expect Upsert to be called with correct job data
				mockJobStore.EXPECT().
					Upsert(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, job *jobModel.ScheduleJob) error {
						// Verify job fields
						if job.Name != tt.jobUID {
							t.Errorf("Expected job name '%s', got '%s'", tt.jobUID, job.Name)
						}
						if job.Type != tt.jobType {
							t.Errorf("Expected job type '%s', got '%s'", tt.jobType, job.Type)
						}
						if !job.IsRecurring {
							t.Error("Expected IsRecurring to be true")
						}
						if job.RecurringCron != tt.cronDef {
							t.Errorf("Expected cron '%s', got '%s'", tt.cronDef, job.RecurringCron)
						}
						if job.State != jobModel.StateScheduled {
							t.Errorf("Expected state 'scheduled', got '%s'", job.State)
						}
						if job.Scheduled == nil {
							t.Error("Expected Scheduled time to be set")
						}
						if job.Priority != jobModel.PriorityElevated {
							t.Error("Expected elevated priority for recurring jobs")
						}
						return nil
					}).
					Times(1)
			}

			err := scheduler.addRecurringJob(context.Background(), tt.jobUID, tt.jobType, tt.cronDef, "", tt.maxDur)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestScheduler_preExec(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobStore := mockStore.NewMockInterface(ctrl)
	mockMxManager := &mockMutexManager{}
	mockPubsubService := &mockPubSub{}

	testLog := logger.NewNopLogger()
	scheduler, _ := NewScheduler(
		mockJobStore,
		nil,
		mockMxManager,
		mockPubsubService,
		testLog,
		"test-instance-123",
		5,
		24*time.Hour,
	)

	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:               "test-job",
		Type:               "test-type",
		State:              jobModel.StateScheduled,
		MaxDurationSeconds: 30,
		TotalExecutions:    5,
		Scheduled:          &nowLocal,
	}

	scheduler.preExec(job)

	// Verify fields were updated correctly
	if job.State != jobModel.StateRunning {
		t.Errorf("Expected state 'running', got '%s'", job.State)
	}

	if job.RunBy != "test-instance-123" {
		t.Errorf("Expected RunBy 'test-instance-123', got '%s'", job.RunBy)
	}

	if job.RunProgress != ProgressMin {
		t.Errorf("Expected RunProgress %d, got %d", ProgressMin, job.RunProgress)
	}

	if job.TotalExecutions != 6 {
		t.Errorf("Expected TotalExecutions 6, got %d", job.TotalExecutions)
	}

	if job.LastExecuted == nil {
		t.Error("Expected LastExecuted to be set")
	}

	if job.RunDeadline == nil {
		t.Error("Expected RunDeadline to be set")
	}

	if job.Result != "" {
		t.Error("Expected Result to be cleared")
	}

	if job.LastFailureError != "" {
		t.Error("Expected LastFailureError to be cleared")
	}

	if job.EditedDate == nil {
		t.Error("Expected EditedDate to be set")
	}
}

func TestPostExec_Success(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:                "test-job",
		Type:                "test-type",
		State:               jobModel.StateRunning,
		RunBy:               "test-instance",
		IsRecurring:         false,
		ConsecutiveFailures: 2,
		Scheduled:           &nowLocal,
	}

	postExec(job, "success result", "")

	if job.State != jobModel.StateFinished {
		t.Errorf("Expected state 'finished', got '%s'", job.State)
	}

	if job.Result != "success result" {
		t.Errorf("Expected result 'success result', got '%s'", job.Result)
	}

	if job.RunBy != "" {
		t.Errorf("Expected RunBy to be cleared, got '%s'", job.RunBy)
	}

	if job.RunProgress != ProgressMax {
		t.Errorf("Expected RunProgress %d, got %d", ProgressMax, job.RunProgress)
	}

	if job.EditedDate == nil {
		t.Error("Expected EditedDate to be updated")
	}

	// ConsecutiveFailures should not be reset for non-recurring jobs on success
	if job.ConsecutiveFailures != 2 {
		t.Errorf("Expected ConsecutiveFailures to remain 2, got %d", job.ConsecutiveFailures)
	}
}

func TestPostExec_Failure(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:                "test-job",
		Type:                "test-type",
		State:               jobModel.StateRunning,
		RunBy:               "test-instance",
		IsRecurring:         false,
		ConsecutiveFailures: 1,
		Scheduled:           &nowLocal,
	}

	postExec(job, "", "execution failed")

	if job.State != jobModel.StateFailed {
		t.Errorf("Expected state 'failed', got '%s'", job.State)
	}

	if job.LastFailureError != "execution failed" {
		t.Errorf("Expected error 'execution failed', got '%s'", job.LastFailureError)
	}

	if job.ConsecutiveFailures != 2 {
		t.Errorf("Expected ConsecutiveFailures 2, got %d", job.ConsecutiveFailures)
	}

	if job.RunBy != "" {
		t.Errorf("Expected RunBy to be cleared, got '%s'", job.RunBy)
	}
}

func TestPostExec_RecurringJob_Success(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:                "recurring-job",
		Type:                "test-type",
		State:               jobModel.StateRunning,
		RunBy:               "test-instance",
		IsRecurring:         true,
		RecurringCron:       "*/5 * * * *", // Every 5 minutes
		ConsecutiveFailures: 2,
		Scheduled:           &nowLocal,
	}

	postExec(job, "success", "")

	if job.State != jobModel.StateScheduled {
		t.Errorf("Expected state 'scheduled', got '%s'", job.State)
	}

	if job.ConsecutiveFailures != 0 {
		t.Errorf("Expected ConsecutiveFailures to be reset to 0, got %d", job.ConsecutiveFailures)
	}

	if job.Scheduled == nil {
		t.Error("Expected Scheduled to be updated with next execution time")
	}

	nextTime := time.Time(*job.Scheduled)
	if !nextTime.After(now) {
		t.Error("Next scheduled time should be in the future")
	}
}

func TestPostExec_RecurringJob_Failure(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:                "recurring-job",
		Type:                "test-type",
		State:               jobModel.StateRunning,
		RunBy:               "test-instance",
		IsRecurring:         true,
		RecurringCron:       "*/10 * * * *", // Every 10 minutes
		ConsecutiveFailures: 1,
		Scheduled:           &nowLocal,
	}

	postExec(job, "", "execution failed")

	if job.State != jobModel.StateScheduled {
		t.Errorf("Expected state 'scheduled' (will retry), got '%s'", job.State)
	}

	if job.ConsecutiveFailures != 2 {
		t.Errorf("Expected ConsecutiveFailures 2, got %d", job.ConsecutiveFailures)
	}

	if job.Scheduled == nil {
		t.Error("Expected Scheduled to be updated with next execution time")
	}

	// Should still be rescheduled despite failure
	nextTime := time.Time(*job.Scheduled)
	if !nextTime.After(now) {
		t.Error("Next scheduled time should be in the future")
	}
}

func TestPostExec_RecurringJob_InvalidCron(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	job := &jobModel.ScheduleJob{
		Name:                "recurring-job",
		Type:                "test-type",
		State:               jobModel.StateRunning,
		RunBy:               "test-instance",
		IsRecurring:         true,
		RecurringCron:       "invalid cron expression",
		ConsecutiveFailures: 0,
		Scheduled:           &nowLocal,
	}

	postExec(job, "success", "")

	if job.State != jobModel.StateFailed {
		t.Errorf("Expected state 'failed' due to invalid cron, got '%s'", job.State)
	}

	if !contains(job.LastFailureError, "failed to parse cron string") {
		t.Errorf("Expected error message about cron parsing, got '%s'", job.LastFailureError)
	}
}

func TestPostExec_RetryLogic(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	tests := []struct {
		name                string
		consecutiveFailures int
		maxRetries          int
		expectedState       jobModel.State
		shouldReschedule    bool
	}{
		{
			name:                "first failure with retries available",
			consecutiveFailures: 0,
			maxRetries:          3,
			expectedState:       jobModel.StateScheduled,
			shouldReschedule:    true,
		},
		{
			name:                "within retry limit",
			consecutiveFailures: 2,
			maxRetries:          3,
			expectedState:       jobModel.StateScheduled,
			shouldReschedule:    true,
		},
		{
			name:                "at retry limit (final retry)",
			consecutiveFailures: 2,
			maxRetries:          3,
			expectedState:       jobModel.StateScheduled,
			shouldReschedule:    true,
		},
		{
			name:                "just exceeded retry limit",
			consecutiveFailures: 3,
			maxRetries:          3,
			expectedState:       jobModel.StateFailed,
			shouldReschedule:    false,
		},
		{
			name:                "no retries allowed",
			consecutiveFailures: 0,
			maxRetries:          0,
			expectedState:       jobModel.StateFailed,
			shouldReschedule:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &jobModel.ScheduleJob{
				Name:                "retry-job",
				Type:                "test-type",
				State:               jobModel.StateRunning,
				RunBy:               "test-instance",
				IsRecurring:         false,
				ConsecutiveFailures: tt.consecutiveFailures,
				MaxRetries:          tt.maxRetries,
				Scheduled:           &nowLocal,
			}

			postExec(job, "", "execution failed")

			if job.State != tt.expectedState {
				t.Errorf("Expected state '%s', got '%s'", tt.expectedState, job.State)
			}

			if tt.shouldReschedule {
				if job.Scheduled == nil {
					t.Error("Expected job to be rescheduled")
				} else {
					nextTime := time.Time(*job.Scheduled)
					if !nextTime.After(now) {
						t.Error("Retry should be scheduled in the future")
					}
					// Check that retry is scheduled ~15 seconds from now
					timeDiff := nextTime.Sub(now)
					if timeDiff < 10*time.Second || timeDiff > 20*time.Second {
						t.Errorf("Expected retry delay around 15s, got %v", timeDiff)
					}
				}

				if job.RunProgress != ProgressMin {
					t.Errorf("Expected RunProgress to be reset to %d, got %d", ProgressMin, job.RunProgress)
				}
			}
		})
	}
}

func TestScheduler_RunJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockJobStore := mockStore.NewMockInterface(ctrl)
	mockMxManager := &mockMutexManager{}
	mockPubsubService := &mockPubSub{}

	testLog := logger.NewNopLogger()
	scheduler, _ := NewScheduler(
		mockJobStore,
		nil,
		mockMxManager,
		mockPubsubService,
		testLog,
		"test-instance",
		5,
		24*time.Hour,
	)

	// Expect job lookup (not found, so it will be created)
	mockJobStore.EXPECT().
		GetByName(gomock.Any(), "manual-job").
		Return(nil, pkgErrors.ErrResourceNotFound).
		Times(1)

	// Expect job to be created
	mockJobStore.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, job *jobModel.ScheduleJob) error {
			if job.Name != "manual-job" {
				t.Errorf("Expected job name 'manual-job', got '%s'", job.Name)
			}
			if job.Type != "manual-type" {
				t.Errorf("Expected job type 'manual-type', got '%s'", job.Type)
			}
			if job.Data != "test data" {
				t.Errorf("Expected data 'test data', got '%s'", job.Data)
			}
			if job.State != jobModel.StateScheduled {
				t.Errorf("Expected state 'scheduled', got '%s'", job.State)
			}
			if job.MaxDurationSeconds != 60 {
				t.Errorf("Expected MaxDurationSeconds 60, got %d", job.MaxDurationSeconds)
			}
			if job.IsRecurring {
				t.Error("Expected IsRecurring to be false")
			}
			return nil
		}).
		Times(1)

	err := scheduler.RunJob(context.Background(), "manual-job", "manual-type", "test data", 60)

	if err != nil {
		t.Errorf("RunJob failed: %v", err)
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
