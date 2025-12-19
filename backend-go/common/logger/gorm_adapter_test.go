package logger

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestNewGormLogger(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		SlowThreshold:             100 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  gormlogger.Info,
	})

	if gormLogger == nil {
		t.Fatal("Expected non-nil gorm logger")
	}

	if gormLogger.slowThreshold != 100*time.Millisecond {
		t.Errorf("Expected slow threshold 100ms, got %v", gormLogger.slowThreshold)
	}
}

func TestGormLoggerInfo(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		LogLevel: gormlogger.Info,
	})

	// 测试 Info 日志
	ctx := context.Background()
	gormLogger.Info(ctx, "test message: %s", "hello")

	output := buf.String()
	if !contains(output, "test message: hello") {
		t.Errorf("Expected log to contain 'test message: hello', got: %s", output)
	}
}

func TestGormLoggerWarn(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		LogLevel: gormlogger.Warn,
	})

	// 测试 Warn 日志
	ctx := context.Background()
	gormLogger.Warn(ctx, "warning: %s", "slow query")

	output := buf.String()
	if !contains(output, "warning: slow query") {
		t.Errorf("Expected log to contain 'warning: slow query', got: %s", output)
	}
}

func TestGormLoggerError(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		LogLevel: gormlogger.Error,
	})

	// 测试 Error 日志
	ctx := context.Background()
	gormLogger.Error(ctx, "error: %s", "connection failed")

	output := buf.String()
	if !contains(output, "error: connection failed") {
		t.Errorf("Expected log to contain 'error: connection failed', got: %s", output)
	}
}

func TestGormLoggerTrace(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
		LevelKey:   "level",
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		SlowThreshold:             50 * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  gormlogger.Info,
	})

	ctx := context.Background()

	// 测试慢查询
	t.Run("slow query", func(t *testing.T) {
		buf.Reset()
		begin := time.Now().Add(-100 * time.Millisecond)
		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users", 10
		}, nil)

		output := buf.String()
		if !contains(output, "Slow SQL query") {
			t.Errorf("Expected log to contain 'Slow SQL query', got: %s", output)
		}
	})

	// 测试错误
	t.Run("error", func(t *testing.T) {
		buf.Reset()
		begin := time.Now()
		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users", 0
		}, errors.New("connection error"))

		output := buf.String()
		if !contains(output, "SQL execution error") {
			t.Errorf("Expected log to contain 'SQL execution error', got: %s", output)
		}
	})

	// 测试忽略 RecordNotFound 错误
	t.Run("ignore record not found", func(t *testing.T) {
		buf.Reset()
		begin := time.Now()
		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users WHERE id = ?", 0
		}, gorm.ErrRecordNotFound)

		output := buf.String()
		if contains(output, "SQL execution error") {
			t.Errorf("Should not log RecordNotFound error when ignoreRecordNotFoundError is true, got: %s", output)
		}
	})

	// 测试正常查询
	t.Run("normal query", func(t *testing.T) {
		buf.Reset()
		begin := time.Now()
		gormLogger.Trace(ctx, begin, func() (string, int64) {
			return "SELECT * FROM users", 5
		}, nil)

		output := buf.String()
		if !contains(output, "SQL query") {
			t.Errorf("Expected log to contain 'SQL query', got: %s", output)
		}
	})
}

func TestGormLoggerLogMode(t *testing.T) {
	// 创建测试用 logger
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&bytes.Buffer{}), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		LogLevel: gormlogger.Info,
	})

	// 测试 LogMode
	newLogger := gormLogger.LogMode(gormlogger.Error)
	if newLogger == nil {
		t.Fatal("Expected non-nil logger from LogMode")
	}

	gormLoggerTyped, ok := newLogger.(*GormLogger)
	if !ok {
		t.Fatal("Expected *GormLogger type")
	}

	if gormLoggerTyped.logLevel != gormlogger.Error {
		t.Errorf("Expected log level Error, got %v", gormLoggerTyped.logLevel)
	}
}

func TestGormLoggerSilentMode(t *testing.T) {
	// 创建测试用 logger
	var buf bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		MessageKey: "msg",
	})
	core := zapcore.NewCore(encoder, zapcore.AddSync(&buf), zapcore.DebugLevel)
	zapLogger := zap.New(core)
	logger := &Logger{zapLogger}

	// 创建 GORM 适配器，设置为 Silent 模式
	gormLogger := NewGormLogger(logger, GormLoggerConfig{
		LogLevel: gormlogger.Silent,
	})

	ctx := context.Background()

	// 即使有慢查询也不应该记录
	begin := time.Now().Add(-1 * time.Second)
	gormLogger.Trace(ctx, begin, func() (string, int64) {
		return "SELECT * FROM users", 10
	}, nil)

	output := buf.String()
	if len(output) > 0 {
		t.Errorf("Expected no log output in Silent mode, got: %s", output)
	}
}

// 辅助函数
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
