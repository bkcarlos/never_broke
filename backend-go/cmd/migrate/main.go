package main

import (
	"fmt"
	"log"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/repository/postgres"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	// 加载 .env 文件
	loadEnvFile("../../.env")

	// 加载数据库配置
	dbCfg := dbconfig.LoadFromEnv()
	fmt.Printf("Connecting to database: %s\n", maskPassword(dbCfg.DSN))

	// 连接数据库
	db, err := postgres.OpenPostgres(dbCfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	fmt.Println("✅ Database connected successfully")

	// 创建迁移表
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// 读取已应用的迁移
	appliedMigrations := make(map[string]bool)
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		log.Fatalf("Failed to query migrations: %v", err)
	}
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			log.Fatalf("Failed to scan migration version: %v", err)
		}
		appliedMigrations[version] = true
	}
	rows.Close()

	// 读取迁移文件
	migrationsDir := "../../migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}
	sort.Strings(migrations)

	// 应用未执行的迁移
	for _, migration := range migrations {
		version := strings.TrimSuffix(migration, ".sql")
		if appliedMigrations[version] {
			fmt.Printf("⏭️  Skipping %s (already applied)\n", migration)
			continue
		}

		fmt.Printf("📝 Applying %s...\n", migration)

		// 读取 SQL 文件
		sqlFile := filepath.Join(migrationsDir, migration)
		sqlBytes, err := os.ReadFile(sqlFile)
		if err != nil {
			log.Fatalf("Failed to read migration file %s: %v", migration, err)
		}

		// 执行 SQL
		_, err = db.Exec(string(sqlBytes))
		if err != nil {
			log.Fatalf("Failed to apply migration %s: %v", migration, err)
		}

		// 记录迁移
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version)
		if err != nil {
			log.Fatalf("Failed to record migration %s: %v", migration, err)
		}

		fmt.Printf("✅ Applied %s\n", migration)
	}

	fmt.Println("\n🎉 All migrations applied successfully!")
}

// loadEnvFile 加载 .env 文件中的环境变量
func loadEnvFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		// .env 文件不存在时不报错
		return
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}

// maskPassword 隐藏 DSN 中的密码用于日志输出
func maskPassword(dsn string) string {
	if idx := strings.Index(dsn, "@"); idx > 0 {
		if colonIdx := strings.LastIndex(dsn[:idx], ":"); colonIdx > 0 {
			return dsn[:colonIdx+1] + "****" + dsn[idx:]
		}
	}
	return dsn
}



