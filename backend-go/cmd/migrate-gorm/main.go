package main

import (
	"fmt"
	"log"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/models"
	"neverbroke/internal/repository/postgres"
	"os"
)

func main() {
	// 加载 .env 文件
	loadEnvFile("../../.env")

	// 加载数据库配置
	dbCfg := dbconfig.LoadFromEnv()
	fmt.Printf("Connecting to database: %s\n", maskPassword(dbCfg.DSN))

	// 连接数据库 (使用 GORM)
	db, err := postgres.OpenGorm(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 获取底层 sql.DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	fmt.Println("✅ Database connected successfully (GORM)")

	// 自动迁移所有模型
	fmt.Println("📝 Running GORM auto migrations...")
	err = db.AutoMigrate(
		&models.Account{},
		&models.Transaction{},
		&models.InstallmentPlan{},
		&models.RecurringIncomePlan{},
	)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("✅ All GORM migrations completed successfully!")
	fmt.Println("\n📋 Migrated tables:")
	fmt.Println("  - accounts")
	fmt.Println("  - transactions")
	fmt.Println("  - installment_plans")
	fmt.Println("  - recurring_income_plans")
}

// loadEnvFile 加载 .env 文件中的环境变量
func loadEnvFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// .env 文件不存在时不报错
		return
	}

	lines := splitLines(string(data))
	for _, line := range lines {
		line = trimSpace(line)
		// 跳过空行和注释
		if line == "" || hasPrefix(line, "#") {
			continue
		}
		// 解析 KEY=VALUE
		parts := splitN(line, "=", 2)
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}

// maskPassword 隐藏 DSN 中的密码用于日志输出
func maskPassword(dsn string) string {
	// 查找 @ 符号
	atIdx := -1
	for i := 0; i < len(dsn); i++ {
		if dsn[i] == '@' {
			atIdx = i
			break
		}
	}

	if atIdx > 0 {
		// 查找最后一个冒号
		colonIdx := -1
		for i := atIdx - 1; i >= 0; i-- {
			if dsn[i] == ':' {
				colonIdx = i
				break
			}
		}

		if colonIdx > 0 {
			return dsn[:colonIdx+1] + "****" + dsn[atIdx:]
		}
	}
	return dsn
}

// 简单的字符串工具函数
func splitLines(s string) []string {
	var lines []string
	var line string
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, line)
			line = ""
		} else if s[i] != '\r' {
			line += string(s[i])
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func trimSpace(s string) string {
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}

	return s[start:end]
}

func hasPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}

func splitN(s, sep string, n int) []string {
	if n <= 0 {
		return nil
	}

	var parts []string
	start := 0

	for i := 0; i <= len(s)-len(sep) && len(parts) < n-1; i++ {
		match := true
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				match = false
				break
			}
		}
		if match {
			parts = append(parts, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}

	if start < len(s) {
		parts = append(parts, s[start:])
	} else if start == len(s) && len(parts) < n {
		parts = append(parts, "")
	}

	return parts
}
