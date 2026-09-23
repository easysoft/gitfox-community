// @copyright   Copyright 2009-2026 渠成极客软件（青岛）有限公司(Qucheng Geek Software (Qingdao) Co., Ltd. www.gitfox.net)
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
	pkgJobs "github.com/easysoft/gitfox-community/pkg/jobs"
	"github.com/easysoft/gitfox-community/pkg/logger"
	"github.com/easysoft/gitfox-community/pkg/utils/datetime"
)

// TestIntegration_SimpleJobExecution tests a complete job execution flow
func TestIntegration_SimpleJobExecution(t *testing.T) {
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
		t.Fatalf("Failed to create scheduler: %v", err)
	}

	// Register a simple test job handler
	executionCount := 0
	testHandler := &simpleHandler{
		handleFunc: func(ctx context.Context, input string, reportFn pkgJobs.ProgressReporter) (string, error) {
			executionCount++
			return "Job executed successfully with input: " + input, nil
		},
	}

	err = scheduler.executor.Register("test-job-type", testHandler)
	if err != nil {
		t.Fatalf("Failed to register handler: %v", err)
	}

	// Expect job lookup (not found, so it will be created)
	mockJobStore.EXPECT().
		GetByName(gomock.Any(), "integration-test-job").
		Return(nil, pkgErrors.ErrResourceNotFound).
		Times(1)

	// Expect job creation
	mockJobStore.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, job *jobModel.ScheduleJob) error {
			// Verify the job was created correctly
			if job.Name != "integration-test-job" {
				t.Errorf("Expected job name 'integration-test-job', got '%s'", job.Name)
			}
			if job.Type != "test-job-type" {
				t.Errorf("Expected job type 'test-job-type', got '%s'", job.Type)
			}
			if job.Data != "test-data" {
				t.Errorf("Expected data 'test-data', got '%s'", job.Data)
			}
			return nil
		}).
		Times(1)

	// Run the job
	err = scheduler.RunJob(context.Background(), "integration-test-job", "test-job-type", "test-data", 10)
	if err != nil {
		t.Fatalf("RunJob failed: %v", err)
	}

	// Verify the job was triggered (note: in real scenario, scheduler.Run() would pick it up)
	// For now we just verify the job was created in the database
}

// TestIntegration_RecurringJobCronCalculation tests that cron expressions calculate correctly
func TestIntegration_RecurringJobCronCalculation(t *testing.T) {
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
		name            string
		cronDef         string
		validateNextRun func(t *testing.T, nextRun time.Time)
	}{
		{
			name:    "every 5 minutes",
			cronDef: "*/5 * * * *",
			validateNextRun: func(t *testing.T, nextRun time.Time) {
				now := time.Now()
				// Next run should be within the next 5 minutes
				if nextRun.Before(now) {
					t.Error("Next run should be in the future")
				}
				diff := nextRun.Sub(now)
				if diff > 6*time.Minute {
					t.Errorf("Next run should be within 6 minutes, got %v", diff)
				}
			},
		},
		{
			name:    "every hour at minute 15",
			cronDef: "15 * * * *",
			validateNextRun: func(t *testing.T, nextRun time.Time) {
				if nextRun.Minute() != 15 {
					t.Errorf("Next run should be at minute 15, got minute %d", nextRun.Minute())
				}
			},
		},
		{
			name:    "every day at midnight",
			cronDef: "0 0 * * *",
			validateNextRun: func(t *testing.T, nextRun time.Time) {
				if nextRun.Hour() != 0 || nextRun.Minute() != 0 {
					t.Errorf("Next run should be at midnight, got %02d:%02d", nextRun.Hour(), nextRun.Minute())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedJob *jobModel.ScheduleJob

			mockJobStore.EXPECT().
				Upsert(gomock.Any(), gomock.Any()).
				DoAndReturn(func(ctx context.Context, job *jobModel.ScheduleJob) error {
					capturedJob = job
					return nil
				}).
				Times(1)

			err := scheduler.addRecurringJob(
				context.Background(),
				"test-cron-job",
				"test-type",
				tt.cronDef,
				"",
				5*time.Second,
			)

			if err != nil {
				t.Fatalf("addRecurringJob failed: %v", err)
			}

			if capturedJob == nil {
				t.Fatal("Job was not created")
			}

			if capturedJob.Scheduled == nil {
				t.Fatal("Scheduled time was not set")
			}

			nextRun := time.Time(*capturedJob.Scheduled)
			tt.validateNextRun(t, nextRun)
		})
	}
}

// TestIntegration_JobLifecycle tests the complete lifecycle of a job
func TestIntegration_JobLifecycle(t *testing.T) {
	now := time.Now()
	nowLocal := datetime.Local(now)

	// Create a job in scheduled state
	job := &jobModel.ScheduleJob{
		Name:               "lifecycle-job",
		Type:               "test-type",
		State:              jobModel.StateScheduled,
		MaxDurationSeconds: 30,
		TotalExecutions:    0,
		MaxRetries:         2,
		Scheduled:          &nowLocal,
		IsRecurring:        false,
	}

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

	// Step 1: PreExec - prepare for execution
	t.Log("Step 1: PreExec")
	scheduler.preExec(job)

	if job.State != jobModel.StateRunning {
		t.Errorf("After preExec, expected state 'running', got '%s'", job.State)
	}
	if job.TotalExecutions != 1 {
		t.Errorf("After preExec, expected TotalExecutions 1, got %d", job.TotalExecutions)
	}
	if job.RunBy != "test-instance" {
		t.Errorf("After preExec, expected RunBy 'test-instance', got '%s'", job.RunBy)
	}

	// Step 2: Simulate execution failure
	t.Log("Step 2: PostExec with failure")
	postExec(job, "", "execution failed")

	if job.State != jobModel.StateScheduled {
		t.Errorf("After first failure, expected state 'scheduled' (retry), got '%s'", job.State)
	}
	if job.ConsecutiveFailures != 1 {
		t.Errorf("After first failure, expected ConsecutiveFailures 1, got %d", job.ConsecutiveFailures)
	}
	if job.Scheduled == nil {
		t.Error("After failure, job should be rescheduled")
	}

	// Step 3: Second execution attempt
	t.Log("Step 3: PreExec for retry")
	scheduler.preExec(job)

	if job.State != jobModel.StateRunning {
		t.Errorf("After second preExec, expected state 'running', got '%s'", job.State)
	}
	if job.TotalExecutions != 2 {
		t.Errorf("After second preExec, expected TotalExecutions 2, got %d", job.TotalExecutions)
	}

	// Step 4: Second failure
	t.Log("Step 4: PostExec with second failure")
	postExec(job, "", "execution failed again")

	if job.State != jobModel.StateScheduled {
		t.Errorf("After second failure, expected state 'scheduled' (retry), got '%s'", job.State)
	}
	if job.ConsecutiveFailures != 2 {
		t.Errorf("After second failure, expected ConsecutiveFailures 2, got %d", job.ConsecutiveFailures)
	}

	// Step 5: Third execution attempt
	t.Log("Step 5: PreExec for final retry")
	scheduler.preExec(job)

	// Step 6: Third failure - should exceed retry limit
	t.Log("Step 6: PostExec with third failure (exceeds retry limit)")
	postExec(job, "", "final failure")

	if job.State != jobModel.StateFailed {
		t.Errorf("After exceeding retries, expected state 'failed', got '%s'", job.State)
	}
	if job.ConsecutiveFailures != 3 {
		t.Errorf("After third failure, expected ConsecutiveFailures 3, got %d", job.ConsecutiveFailures)
	}

	t.Log("Job lifecycle test completed successfully")
}

// simpleHandler is a simple handler for testing
type simpleHandler struct {
	handleFunc func(ctx context.Context, input string, reportFn pkgJobs.ProgressReporter) (string, error)
}

func (h *simpleHandler) Handle(ctx context.Context, input string, reportFn pkgJobs.ProgressReporter) (string, error) {
	if h.handleFunc != nil {
		return h.handleFunc(ctx, input, reportFn)
	}
	return "default result", nil
}
