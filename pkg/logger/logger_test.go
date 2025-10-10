package logger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 创建测试配置文件
func createTestConfig(t *testing.T) string {
	configContent := `
test:
  level: -1
  stdout: false
  caller: true
  encoder: console
  max-age: 7
  max-count: 5
  max-size: 10
  file-name: ./test_logs/test.log
  level-color: false
  time-format: 2006-01-02 15:04:05.999
  stack-trace-level: 2
`

	// 创建临时配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// 创建日志目录
	logDir := filepath.Join(tmpDir, "test_logs")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create log directory: %v", err)
	}

	// 更新配置文件中的日志路径
	configContent = strings.Replace(configContent, "./test_logs/test.log",
		filepath.Join(logDir, "test.log"), 1)
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to update test config: %v", err)
	}

	return configPath
}

// TestNewZapMateLogger 测试logger创建
func TestNewZapMateLogger(t *testing.T) {
	configPath := createTestConfig(t)

	logger := NewZapMateLogger(configPath, "test")

	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.isAsync {
		t.Error("Logger should not be async by default")
	}

	if logger.lock == nil {
		t.Error("Logger lock should not be nil")
	}

	if logger.Logger == nil {
		t.Error("Underlying zap logger should not be nil")
	}
}

// TestZapMateLoggerBasicLogging 测试基本日志记录
func TestZapMateLoggerBasicLogging(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")

	// 测试各个级别的日志
	logger.Debug("Debug message", zap.String("key", "debug"))
	logger.Info("Info message", zap.String("key", "info"))
	logger.Warn("Warn message", zap.String("key", "warn"))
	logger.Error("Error message", zap.String("key", "error"))

	// 刷新日志确保写入
	err := logger.Flush()
	if err != nil {
		t.Errorf("Failed to flush logger: %v", err)
	}
}

// TestAsyncLogger 测试异步日志功能
func TestAsyncLogger(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")

	// 设置异步模式
	asyncLogger := logger.SetAsyncer(100)

	if !asyncLogger.isAsync {
		t.Error("Logger should be in async mode")
	}

	if asyncLogger.chanLen != 100 {
		t.Errorf("Expected channel length 100, got %d", asyncLogger.chanLen)
	}

	if asyncLogger.entryChan == nil {
		t.Error("Entry channel should not be nil")
	}

	// 测试重复设置异步模式
	asyncLogger2 := asyncLogger.SetAsyncer(200)
	if asyncLogger2.chanLen != 100 {
		t.Error("Channel length should not change on repeated SetAsyncer call")
	}
}

// TestAsyncLoggerConcurrency 测试异步日志并发安全
func TestAsyncLoggerConcurrency(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test").SetAsyncer(1000)

	var wg sync.WaitGroup
	numGoroutines := 10
	numLogsPerGoroutine := 100

	// 启动多个goroutine并发写日志
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogsPerGoroutine; j++ {
				logger.Info("Concurrent log",
					zap.Int("goroutine", id),
					zap.Int("log_num", j))
			}
		}(i)
	}

	wg.Wait()

	// 刷新日志
	err := logger.Flush()
	if err != nil {
		t.Errorf("Failed to flush async logger: %v", err)
	}
}

// TestSugaredLogger 测试Sugar Logger功能
func TestSugaredLogger(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")
	sugar := logger.Sugar()

	if sugar == nil {
		t.Fatal("Sugar logger should not be nil")
	}

	if sugar.base == nil {
		t.Error("Sugar logger base should not be nil")
	}

	// 测试Sugar logger的各种方法
	sugar.Debug("Debug message")
	sugar.Info("Info message")
	sugar.Warn("Warn message")
	sugar.Error("Error message")

	// 测试格式化方法
	sugar.Debugf("Debug %s", "formatted")
	sugar.Infof("Info %d", 123)
	sugar.Warnf("Warn %v", true)
	sugar.Errorf("Error %f", 3.14)

	// 测试结构化方法
	sugar.Debugw("Debug with fields", "key1", "value1", "key2", 123)
	sugar.Infow("Info with fields", "user", "john", "age", 30)
	sugar.Warnw("Warn with fields", "status", "warning")
	sugar.Errorw("Error with fields", "error", "something went wrong")

	err := sugar.Flush()
	if err != nil {
		t.Errorf("Failed to flush sugar logger: %v", err)
	}
}

// TestSugaredLoggerDesugar 测试Sugar Logger的Desugar功能
func TestSugaredLoggerDesugar(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")
	sugar := logger.Sugar()

	// 测试Desugar
	desugared := sugar.Desugar()
	if desugared == nil {
		t.Fatal("Desugared logger should not be nil")
	}

	// 测试desugared logger功能
	desugared.Info("Desugared info", zap.String("type", "desugared"))

	err := desugared.Flush()
	if err != nil {
		t.Errorf("Failed to flush desugared logger: %v", err)
	}
}

// TestSweetenFields 测试字段转换功能
func TestSweetenFields(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")
	sugar := logger.Sugar()

	tests := []struct {
		name     string
		args     []interface{}
		expected int // 期望的有效字段数量
	}{
		{
			name:     "Empty args",
			args:     []interface{}{},
			expected: 0,
		},
		{
			name:     "Valid key-value pairs",
			args:     []interface{}{"key1", "value1", "key2", 123},
			expected: 2,
		},
		{
			name:     "Mixed with zap.Field",
			args:     []interface{}{zap.String("field1", "value1"), "key2", "value2"},
			expected: 2,
		},
		{
			name:     "Invalid non-string key",
			args:     []interface{}{123, "value1", "key2", "value2"},
			expected: 1, // 只有一个有效字段
		},
		{
			name:     "Odd number of args",
			args:     []interface{}{"key1", "value1", "key2"},
			expected: 1, // 最后一个参数被忽略
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := sugar.sweetenFields(tt.args)
			if len(fields) != tt.expected {
				t.Errorf("Expected %d fields, got %d", tt.expected, len(fields))
			}
		})
	}
}

// TestLoggerClone 测试logger克隆功能
func TestLoggerClone(t *testing.T) {
	configPath := createTestConfig(t)
	original := NewZapMateLogger(configPath, "test")
	original.SetAsyncer(100)

	cloned := original.clone()

	if cloned == original {
		t.Error("Cloned logger should be a different instance")
	}

	if cloned.isAsync != original.isAsync {
		t.Error("Cloned logger should have same async status")
	}

	if cloned.chanLen != original.chanLen {
		t.Error("Cloned logger should have same channel length")
	}
}

// TestLoggerWrite 测试写入功能
func TestLoggerWrite(t *testing.T) {
	configPath := createTestConfig(t)
	logger := NewZapMateLogger(configPath, "test")

	// 测试同步写入
	logger.write(zapcore.InfoLevel, "Sync write test", zap.String("mode", "sync"))

	// 设置异步模式并测试异步写入
	asyncLogger := logger.SetAsyncer(10)
	asyncLogger.write(zapcore.InfoLevel, "Async write test", zap.String("mode", "async"))

	// 等待异步写入完成
	time.Sleep(100 * time.Millisecond)

	err := asyncLogger.Flush()
	if err != nil {
		t.Errorf("Failed to flush logger: %v", err)
	}
}

// BenchmarkSyncLogging 同步日志性能基准测试
func BenchmarkSyncLogging(b *testing.B) {
	configPath := createTestConfig(&testing.T{})
	logger := NewZapMateLogger(configPath, "test")

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Benchmark sync log", zap.Int("iteration", b.N))
		}
	})
}

// BenchmarkAsyncLogging 异步日志性能基准测试
func BenchmarkAsyncLogging(b *testing.B) {
	configPath := createTestConfig(&testing.T{})
	logger := NewZapMateLogger(configPath, "test").SetAsyncer(10000)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("Benchmark async log", zap.Int("iteration", b.N))
		}
	})

	logger.Flush()
}

// BenchmarkSugaredLogging Sugar Logger性能基准测试
func BenchmarkSugaredLogging(b *testing.B) {
	configPath := createTestConfig(&testing.T{})
	logger := NewZapMateLogger(configPath, "test")
	sugar := logger.Sugar()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			sugar.Infow("Benchmark sugar log", "iteration", b.N, "type", "benchmark")
		}
	})
}
