package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *Logger
	mu           sync.RWMutex // 保护 globalLogger 的读写锁
)

// Logger 封装 zap.Logger
type Logger struct {
	*zap.Logger
}

// Config 日志配置
type Config struct {
	Level       string   // debug, info, warn, error
	OutputPaths []string // 输出路径列表，可包含 stdout, stderr, 或文件路径，支持多个同时输出
	Format      string   // json 或 console
}

// Init 初始化全局 logger（线程安全）
func Init(cfg Config) error {
	logger, err := NewLogger(cfg)
	if err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger

	return nil
}

// NewLogger 创建新的 logger 实例
func NewLogger(cfg Config) (*Logger, error) {
	// 解析日志级别
	level := parseLevel(cfg.Level)

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 如果没有配置输出路径，默认输出到 stdout
	if len(cfg.OutputPaths) == 0 {
		cfg.OutputPaths = []string{"stdout"}
	}

	// 创建多个 core（支持同时输出到多个目标）
	var cores []zapcore.Core

	for _, outputPath := range cfg.OutputPaths {
		// 为每个输出目标创建编码器
		var encoder zapcore.Encoder
		var encoderCfg zapcore.EncoderConfig

		// 复制编码器配置
		encoderCfg = encoderConfig

		// 根据输出目标选择编码器
		if outputPath == "stdout" || outputPath == "stderr" {
			// 控制台输出使用彩色编码器
			if cfg.Format == "json" {
				encoder = zapcore.NewJSONEncoder(encoderCfg)
			} else {
				encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder // 彩色输出
				encoder = zapcore.NewConsoleEncoder(encoderCfg)
			}
		} else {
			// 文件输出使用普通编码器（JSON 格式更适合文件）
			if cfg.Format == "json" {
				encoder = zapcore.NewJSONEncoder(encoderCfg)
			} else {
				encoder = zapcore.NewConsoleEncoder(encoderCfg)
			}
		}

		// 配置输出目标
		var writeSyncer zapcore.WriteSyncer
		if outputPath == "stdout" {
			writeSyncer = zapcore.AddSync(os.Stdout)
		} else if outputPath == "stderr" {
			writeSyncer = zapcore.AddSync(os.Stderr)
		} else {
			// 文件输出
			file, err := os.OpenFile(outputPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				return nil, err
			}
			writeSyncer = zapcore.AddSync(file)
		}

		// 创建 core 并添加到列表
		core := zapcore.NewCore(encoder, writeSyncer, level)
		cores = append(cores, core)
	}

	// 合并多个 core
	var core zapcore.Core
	if len(cores) == 1 {
		core = cores[0]
	} else {
		core = zapcore.NewTee(cores...)
	}

	// 创建 logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{zapLogger}, nil
}

// parseLevel 解析日志级别
func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// GetLogger 获取全局 logger（线程安全）
func GetLogger() *Logger {
	mu.RLock()
	logger := globalLogger
	mu.RUnlock()

	if logger == nil {
		// 使用写锁进行初始化
		mu.Lock()
		// 双重检查，避免重复初始化
		if globalLogger == nil {
			_ = initDefault()
		}
		logger = globalLogger
		mu.Unlock()
	}

	return logger
}

// initDefault 使用默认配置初始化（内部使用，调用前需持有写锁）
func initDefault() error {
	logger, err := NewLogger(Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	})
	if err != nil {
		return err
	}
	globalLogger = logger
	return nil
}

// 便捷方法 - 使用全局 logger

// Debug 记录 debug 级别日志
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Info 记录 info 级别日志
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Warn 记录 warn 级别日志
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error 记录 error 级别日志
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal 记录 fatal 级别日志并退出程序
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With 创建带有额外字段的 logger
func With(fields ...zap.Field) *Logger {
	return &Logger{GetLogger().With(fields...)}
}

// Sync 刷新日志缓冲区（线程安全）
func Sync() error {
	mu.RLock()
	logger := globalLogger
	mu.RUnlock()

	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// SetLogger 设置全局 logger（高级用法，线程安全）
func SetLogger(logger *Logger) {
	mu.Lock()
	defer mu.Unlock()
	globalLogger = logger
}

// ResetLogger 重置全局 logger（主要用于测试，线程安全）
func ResetLogger() {
	mu.Lock()
	defer mu.Unlock()
	if globalLogger != nil {
		_ = globalLogger.Sync()
	}
	globalLogger = nil
}
