package logger

import (
	"sync"
	"testing"

	"go.uber.org/zap"
)

// TestConcurrentInit 测试并发初始化的安全性
func TestConcurrentInit(t *testing.T) {
	ResetLogger()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := Init(Config{
				Level:       "info",
				OutputPaths: []string{"stdout"},
				Format:      "console",
			})
			if err != nil {
				t.Errorf("Init failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// 验证 logger 已初始化
	if GetLogger() == nil {
		t.Error("Logger should be initialized")
	}
}

// TestConcurrentLogging 测试并发日志记录的安全性
func TestConcurrentLogging(t *testing.T) {
	ResetLogger()

	err := Init(Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	})
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	logsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				Info("Concurrent log test",
					zap.Int("goroutine", id),
					zap.Int("iteration", j),
				)
			}
		}(i)
	}

	wg.Wait()
}

// TestGetLoggerLazyInit 测试懒加载初始化
func TestGetLoggerLazyInit(t *testing.T) {
	ResetLogger()

	// 不调用 Init，直接使用 GetLogger
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger should return a default logger")
	}

	// 应该能正常记录日志
	Info("Test lazy init")
}

// TestConcurrentGetLogger 测试并发获取 logger
func TestConcurrentGetLogger(t *testing.T) {
	ResetLogger()

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger := GetLogger()
			if logger == nil {
				t.Error("GetLogger should never return nil")
			}
		}()
	}

	wg.Wait()
}

// TestSetLogger 测试设置自定义 logger
func TestSetLogger(t *testing.T) {
	ResetLogger()

	customLogger, err := NewLogger(Config{
		Level:       "debug",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	})
	if err != nil {
		t.Fatalf("Failed to create custom logger: %v", err)
	}

	SetLogger(customLogger)

	logger := GetLogger()
	if logger != customLogger {
		t.Error("GetLogger should return the custom logger")
	}
}

// TestResetLogger 测试重置 logger
func TestResetLogger(t *testing.T) {
	// 初始化 logger
	err := Init(Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	})
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// 验证已初始化
	if GetLogger() == nil {
		t.Error("Logger should be initialized")
	}

	// 重置
	ResetLogger()

	// 重置后 GetLogger 应该触发懒加载
	logger := GetLogger()
	if logger == nil {
		t.Error("GetLogger should return a default logger after reset")
	}
}

// BenchmarkConcurrentLogging 并发日志记录性能测试
func BenchmarkConcurrentLogging(b *testing.B) {
	ResetLogger()

	err := Init(Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "json",
	})
	if err != nil {
		b.Fatalf("Failed to init logger: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Info("Benchmark log",
				zap.Int("iteration", b.N),
				zap.String("data", "test"),
			)
		}
	})
}

// BenchmarkGetLogger 获取 logger 性能测试
func BenchmarkGetLogger(b *testing.B) {
	ResetLogger()

	err := Init(Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	})
	if err != nil {
		b.Fatalf("Failed to init logger: %v", err)
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = GetLogger()
		}
	})
}
