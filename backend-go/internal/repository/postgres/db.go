package postgres

import (
	"database/sql"
	"neverbroke/common/logger"
	"neverbroke/internal/dbconfig"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// OpenPostgres 打开 PostgreSQL 数据库连接（原生 sql.DB）
func OpenPostgres(cfg dbconfig.DBConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.DSN)
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// OpenGorm 打开 GORM 数据库连接
func OpenGorm(cfg dbconfig.DBConfig) (*gorm.DB, error) {
	// 获取全局 logger 并创建 GORM 适配器
	customLogger := logger.GetLogger()
	gormLogger := logger.NewGormLogger(customLogger, logger.GormLoggerConfig{
		SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
		IgnoreRecordNotFoundError: true,                   // 忽略 RecordNotFound 错误
		LogLevel:                  gormlogger.Info,        // 日志级别
	})

	// 配置 GORM
	gormConfig := &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now()
		},
		// 禁用外键约束（如果需要的话）
		DisableForeignKeyConstraintWhenMigrating: true,
	}

	// 打开连接
	db, err := gorm.Open(postgres.Open(cfg.DSN), gormConfig)
	if err != nil {
		return nil, err
	}

	// 获取底层的 sql.DB 以设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
