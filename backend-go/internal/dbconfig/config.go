package dbconfig

import (
	"fmt"
	"os"
)

// DBConfig 数据库配置
type DBConfig struct {
	DSN string
}

// LoadFromEnv 从环境变量加载数据库配置
func LoadFromEnv() DBConfig {
	// 优先读取完整的 DSN
	if dsn := os.Getenv("NEVERBROKE_PG_DSN"); dsn != "" {
		return DBConfig{DSN: dsn}
	}

	// 否则从各个环境变量组装
	host := getEnvOrDefault("NEVERBROKE_PG_HOST", "127.0.0.1")
	port := getEnvOrDefault("NEVERBROKE_PG_PORT", "5432")
	user := getEnvOrDefault("NEVERBROKE_PG_USER", "neverbroke")
	password := getEnvOrDefault("NEVERBROKE_PG_PASSWORD", "neverbroke")
	database := getEnvOrDefault("NEVERBROKE_PG_DATABASE", "neverbroke")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, password, host, port, database)

	return DBConfig{DSN: dsn}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}


