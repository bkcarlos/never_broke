package postgres

import (
	"database/sql"
	"neverbroke/internal/dbconfig"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// OpenPostgres 打开 PostgreSQL 数据库连接
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




