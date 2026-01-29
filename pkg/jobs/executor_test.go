package jobs

import (
	"context"
	"errors"
	"testing"
	"time"

	"gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/pubsub"
)

// mockPublisher is a mock implementation of pubsub.Publisher for testing
type mockPublisher struct {
	publishCalled bool
	publishError  error
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, payload []byte, options ...pubsub.PublishOption) error {
	m.publishCalled = true
	return m.publishError
}

// mockHandler is a mock implementation of Handler for testing
type mockHandler struct {
	handleFunc func(ctx context.Context, input string, reportFn ProgressReporter) (string, error)
}

func (m *mockHandler) Handle(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx, input, reportFn)
	}
	return "success", nil
}

func TestNewExecutor(t *testing.T) {
	publisher := &mockPublisher{}
	executor := NewExecutor(publisher)

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if executor.publisher == nil {
		t.Error("Publisher not set")
	}

	if executor.handlerMap == nil {
		t.Error("Handler map not initialized")
	}

	if executor.handlerComplete {
		t.Error("handlerComplete should be false initially")
	}
}

func TestExecutor_Register(t *testing.T) {
	tests := []struct {
		name        string
		jobType     string
		handler     Handler
		setup       func(*Executor)
		expectError bool
		errorMsg    string
	}{
		{
			name:        "successful registration",
			jobType:     "test-job",
			handler:     &mockHandler{},
			expectError: false,
		},
		{
			name:        "empty job type",
			jobType:     "",
			handler:     &mockHandler{},
			expectError: true,
			errorMsg:    "jobType must not be empty",
		},
		{
			name:        "nil handler",
			jobType:     "test-job",
			handler:     nil,
			expectError: true,
			errorMsg:    "provided Handler is nil",
		},
		{
			name:    "duplicate registration",
			jobType: "test-job",
			handler: &mockHandler{},
			setup: func(e *Executor) {
				_ = e.Register("test-job", &mockHandler{})
			},
			expectError: true,
			errorMsg:    "a Handler is already defined",
		},
		{
			name:    "registration after finish",
			jobType: "test-job",
			handler: &mockHandler{},
			setup: func(e *Executor) {
				e.FinishRegistration()
			},
			expectError: true,
			errorMsg:    "job handler registration is complete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(&mockPublisher{})

			if tt.setup != nil {
				tt.setup(executor)
			}

			err := executor.Register(tt.jobType, tt.handler)

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

func TestExecutor_FinishRegistration(t *testing.T) {
	executor := NewExecutor(&mockPublisher{})

	if executor.handlerComplete {
		t.Error("handlerComplete should be false initially")
	}

	executor.FinishRegistration()

	if !executor.handlerComplete {
		t.Error("handlerComplete should be true after FinishRegistration")
	}

	// Try to register after finish
	err := executor.Register("test", &mockHandler{})
	if err == nil {
		t.Error("Should not allow registration after finish")
	}
}

func TestExecutor_Exec(t *testing.T) {
	tests := []struct {
		name         string
		jobType      string
		input        string
		setup        func(*Executor)
		expectResult string
		expectError  bool
		errorMsg     string
	}{
		{
			name:    "successful execution",
			jobType: "test-job",
			input:   "test-input",
			setup: func(e *Executor) {
				handler := &mockHandler{
					handleFunc: func(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
						return "job completed: " + input, nil
					},
				}
				_ = e.Register("test-job", handler)
			},
			expectResult: "job completed: test-input",
			expectError:  false,
		},
		{
			name:    "handler not registered",
			jobType: "unknown-job",
			input:   "test-input",
			setup: func(e *Executor) {
				// Don't register any handler
			},
			expectError: true,
			errorMsg:    "no handler registered",
		},
		{
			name:    "handler returns error",
			jobType: "failing-job",
			input:   "test-input",
			setup: func(e *Executor) {
				handler := &mockHandler{
					handleFunc: func(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
						return "", errors.New("job execution failed")
					},
				}
				_ = e.Register("failing-job", handler)
			},
			expectError: true,
			errorMsg:    "job execution failed",
		},
		{
			name:    "handler panics",
			jobType: "panic-job",
			input:   "test-input",
			setup: func(e *Executor) {
				handler := &mockHandler{
					handleFunc: func(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
						panic("something went wrong")
					},
				}
				_ = e.Register("panic-job", handler)
			},
			expectError: true,
			errorMsg:    "panic while processing",
		},
		{
			name:    "context cancellation",
			jobType: "slow-job",
			input:   "test-input",
			setup: func(e *Executor) {
				handler := &mockHandler{
					handleFunc: func(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
						select {
						case <-ctx.Done():
							return "", ctx.Err()
						case <-time.After(1 * time.Second):
							return "should not reach here", nil
						}
					},
				}
				_ = e.Register("slow-job", handler)
			},
			expectError: true,
			errorMsg:    "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewExecutor(&mockPublisher{})

			if tt.setup != nil {
				tt.setup(executor)
			}

			ctx := context.Background()
			if tt.name == "context cancellation" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel() // Cancel immediately
			}

			noOpReporter := func(progress int, result string) ([]byte, error) {
				return nil, nil
			}

			result, err := executor.Exec(ctx, "test-job-name", tt.jobType, tt.input, noOpReporter)

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
				if result != tt.expectResult {
					t.Errorf("Expected result '%s', got '%s'", tt.expectResult, result)
				}
			}
		})
	}
}

// TestExecutor_Exec_WithProgressReporter tests that handlers can use progress reporter
// Note: The public Exec method uses a no-op progress reporter internally
func TestExecutor_Exec_WithProgressReporter(t *testing.T) {
	executor := NewExecutor(&mockPublisher{})

	progressCallCount := 0

	handler := &mockHandler{
		handleFunc: func(ctx context.Context, input string, reportFn ProgressReporter) (string, error) {
			// Call progress reporter multiple times
			_, err := reportFn(25, "quarter done")
			if err != nil {
				return "", err
			}
			progressCallCount++

			_, err = reportFn(50, "halfway done")
			if err != nil {
				return "", err
			}
			progressCallCount++

			_, err = reportFn(75, "three quarters done")
			if err != nil {
				return "", err
			}
			progressCallCount++

			return "completed", nil
		},
	}

	_ = executor.Register("progress-job", handler)

	// Use a no-op reporter for the test
	noOpReporter := func(progress int, result string) ([]byte, error) {
		return nil, nil
	}

	result, err := executor.Exec(context.Background(), "test", "progress-job", "input", noOpReporter)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if result != "completed" {
		t.Errorf("Expected result 'completed', got '%s'", result)
	}

	// The handler should have called reportFn 3 times (even though it's a no-op)
	if progressCallCount != 3 {
		t.Errorf("Expected progress reporter to be called 3 times, got %d", progressCallCount)
	}
}

// Helper function to check if a string contains a substring
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
