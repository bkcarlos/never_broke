package main

import (
	"bufio"
	"neverbroke/common/logger"
	"neverbroke/internal/auth"
	"neverbroke/internal/citysocialrate"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/http/router"
	"neverbroke/internal/models"
	"neverbroke/internal/repository/postgres"
	"os"
	"strings"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func main() {
	// 加载 .env 文件
	loadEnvFile(".env")

	// 初始化日志 - 同时输出到控制台和文件
	if err := logger.Init(logger.Config{
		Level: "info", // 可配置：debug, info, warn, error
		OutputPaths: []string{
			"stdout",         // 输出到控制台
			"./logs/app.log", // 同时输出到文件
		},
		Format: "console", // 可配置：console 或 json
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("🚀 NeverBroke API Server starting...",
		zap.String("version", "v0.1.0"),
		zap.String("port", "8080"),
	)

	// 初始化城市社保比例数据
	citysocialrate.InitDefaultRates()
	logger.Info("City social rates initialized")

	// 加载数据库配置
	dbCfg := dbconfig.LoadFromEnv()
	logger.Info("Database config loaded", zap.String("dsn", maskPassword(dbCfg.DSN)))

	// 连接 PostgreSQL 数据库 (使用 GORM)
	db, err := postgres.OpenGorm(dbCfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// 获取底层 sql.DB 以便在 defer 中关闭
	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get underlying sql.DB", zap.Error(err))
	}
	defer sqlDB.Close()
	logger.Info("✅ Database connected successfully (GORM)")

	// 自动迁移数据库表
	if err := autoMigrate(db); err != nil {
		logger.Fatal("Failed to run auto migrations", zap.Error(err))
	}
	logger.Info("✅ Database tables migrated successfully")

	// 创建 PostgreSQL 存储
	store := postgres.NewPgStore(db)
	logger.Info("PostgreSQL store initialized (GORM)")

	// 设置路由（传入 db 用于认证系统）
	r := router.Setup(store, db)
	logger.Info("Router setup completed")

	// 启动服务
	logger.Info("Server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}

// loadEnvFile 加载 .env 文件中的环境变量
func loadEnvFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		// .env 文件不存在时不报错，使用默认值
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		// 认证模型（Token 方案）
		&auth.User{},
		&auth.RefreshToken{},
		&auth.OAuthIdentity{},
		&auth.OAuthState{},
		// 业务模型
		&models.Account{},
		&models.Transaction{},
		&models.InstallmentPlan{},
		&models.RecurringIncomePlan{},
	)
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
