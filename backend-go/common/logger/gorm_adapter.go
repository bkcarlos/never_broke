package logger

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// GormLogger 适配器，将自定义 Logger 适配到 GORM 的 logger 接口
type GormLogger struct {
	logger                    *Logger
	logLevel                  gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

// GormLoggerConfig GORM Logger 配置
type GormLoggerConfig struct {
	SlowThreshold             time.Duration // 慢查询阈值
	IgnoreRecordNotFoundError bool          // 是否忽略 RecordNotFound 错误
	LogLevel                  gormlogger.LogLevel
}

// NewGormLogger 创建 GORM logger 适配器
func NewGormLogger(logger *Logger, config GormLoggerConfig) *GormLogger {
	if config.SlowThreshold == 0 {
		config.SlowThreshold = 200 * time.Millisecond
	}

	return &GormLogger{
		logger:                    logger,
		logLevel:                  config.LogLevel,
		slowThreshold:             config.SlowThreshold,
		ignoreRecordNotFoundError: config.IgnoreRecordNotFoundError,
	}
}

// LogMode 设置日志级别
func (l *GormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := *l
	newLogger.logLevel = level
	return &newLogger
}

// Info 记录 info 级别日志
func (l *GormLogger) Info(ctx context.Context, msg string, args ...interface{}) {
	if l.logLevel >= gormlogger.Info {
		l.logger.Info(fmt.Sprintf(msg, args...), zap.String("source", "gorm"))
	}
}

// Warn 记录 warn 级别日志
func (l *GormLogger) Warn(ctx context.Context, msg string, args ...interface{}) {
	if l.logLevel >= gormlogger.Warn {
		l.logger.Warn(fmt.Sprintf(msg, args...), zap.String("source", "gorm"))
	}
}

// Error 记录 error 级别日志
func (l *GormLogger) Error(ctx context.Context, msg string, args ...interface{}) {
	if l.logLevel >= gormlogger.Error {
		l.logger.Error(fmt.Sprintf(msg, args...), zap.String("source", "gorm"))
	}
}

// Trace 记录 SQL 执行日志
func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("source", "gorm"),
		zap.Duration("elapsed", elapsed),
		zap.String("sql", sql),
		zap.Int64("rows", rows),
	}

	switch {
	case err != nil && l.logLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !l.ignoreRecordNotFoundError):
		// 记录错误
		fields = append(fields, zap.Error(err))
		l.logger.Error("SQL execution error", fields...)

	case elapsed > l.slowThreshold && l.slowThreshold != 0 && l.logLevel >= gormlogger.Warn:
		// 记录慢查询
		fields = append(fields, zap.Duration("threshold", l.slowThreshold))
		l.logger.Warn("Slow SQL query", fields...)

	case l.logLevel >= gormlogger.Info:
		// 记录正常查询
		l.logger.Debug("SQL query", fields...)
	}
}
