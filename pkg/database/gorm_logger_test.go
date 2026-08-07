package database

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm/logger"

	applogger "gitfox.zcorp.cc/git/Pangu/gitfox-community/pkg/logger"
)

func TestGormLoggerTraceWithTraceID(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: time.Second,
		LogLevel:      logger.Info,
		Colorful:      false,
	})

	ctx := applogger.WithTraceID(context.Background(), "trace-abc-123")
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT * FROM users WHERE id = 1", 1
	}, nil)

	out := buf.String()
	if !strings.Contains(out, "[trace_id=trace-abc-123]") {
		t.Fatalf("expected trace_id in output, got: %q", out)
	}
	if !strings.Contains(out, "SELECT * FROM users WHERE id = 1") {
		t.Fatalf("expected sql in output, got: %q", out)
	}
}

func TestGormLoggerTraceWithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: time.Second,
		LogLevel:      logger.Info,
		Colorful:      false,
	})

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	out := buf.String()
	if strings.Contains(out, "trace_id") {
		t.Fatalf("expected no trace_id in output, got: %q", out)
	}
	if !strings.Contains(out, "SELECT 1") {
		t.Fatalf("expected sql in output, got: %q", out)
	}
}

func TestGormLoggerTraceError(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: time.Second,
		LogLevel:      logger.Error,
		Colorful:      false,
	})

	ctx := applogger.WithTraceID(context.Background(), "trace-err")
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT * FROM missing", -1
	}, errors.New("table not found"))

	out := buf.String()
	if !strings.Contains(out, "[trace_id=trace-err]") {
		t.Fatalf("expected trace_id in error output, got: %q", out)
	}
	if !strings.Contains(out, "table not found") {
		t.Fatalf("expected error in output, got: %q", out)
	}
}

func TestGormLoggerTraceSilent(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: time.Second,
		LogLevel:      logger.Silent,
		Colorful:      false,
	})

	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		return "SELECT 1", 1
	}, nil)

	if buf.Len() != 0 {
		t.Fatalf("expected no output for silent level, got: %q", buf.String())
	}
}

func TestGormLoggerTraceSlowThreshold(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: 1 * time.Nanosecond,
		LogLevel:      logger.Warn,
		Colorful:      false,
	})

	time.Sleep(2 * time.Millisecond)
	ctx := applogger.WithTraceID(context.Background(), "trace-slow")
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT * FROM big_table", 100
	}, nil)

	out := buf.String()
	if !strings.Contains(out, "SLOW SQL") {
		t.Fatalf("expected SLOW SQL in output, got: %q", out)
	}
	if !strings.Contains(out, "[trace_id=trace-slow]") {
		t.Fatalf("expected trace_id in slow output, got: %q", out)
	}
}

func TestGormLoggerTraceWarnFastQueryNoOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: 500 * time.Millisecond,
		LogLevel:      logger.Warn,
		Colorful:      false,
	})

	fcCalled := false
	l.Trace(context.Background(), time.Now(), func() (string, int64) {
		fcCalled = true
		return "SELECT * FROM users", 1
	}, nil)

	if buf.Len() != 0 {
		t.Fatalf("expected no output for fast query at Warn level, got: %q", buf.String())
	}
	if fcCalled {
		t.Fatal("expected fc() not to be called for fast query at Warn level")
	}
}

func TestGormLoggerTraceWarnSlowQueryOutput(t *testing.T) {
	var buf bytes.Buffer
	l := newGormLogger(&buf, logger.Config{
		SlowThreshold: 1 * time.Nanosecond,
		LogLevel:      logger.Warn,
		Colorful:      false,
	})

	time.Sleep(2 * time.Millisecond)
	ctx := applogger.WithTraceID(context.Background(), "trace-warn-slow")
	l.Trace(ctx, time.Now(), func() (string, int64) {
		return "SELECT * FROM big_table WHERE status = 1", 5
	}, nil)

	out := buf.String()
	if !strings.Contains(out, "SLOW SQL") {
		t.Fatalf("expected SLOW SQL in output, got: %q", out)
	}
	if !strings.Contains(out, "[trace_id=trace-warn-slow]") {
		t.Fatalf("expected trace_id in slow output, got: %q", out)
	}
}
