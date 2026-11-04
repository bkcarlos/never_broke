package integration

import (
	"fmt"
	"neverbroke/common/logger"
	"neverbroke/internal/auth"
	"neverbroke/internal/citysocialrate"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/http/router"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"neverbroke/internal/repository/postgres"
	"neverbroke/internal/workspace"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	testDB     *gorm.DB
	testRouter *gin.Engine
	testStore  *postgres.PgStore
)

// TestMain 测试主入口，在所有测试前后进行设置和清理
func TestMain(m *testing.M) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 初始化日志（测试模式）
	if err := logger.Init(logger.Config{
		Level:       "info",
		OutputPaths: []string{"stdout"},
		Format:      "console",
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	// 初始化城市社保比例数据
	citysocialrate.InitDefaultRates()

	// 设置测试数据库（使用独立的测试数据库）
	setupTestDB()

	// 运行所有测试
	code := m.Run()

	// 清理
	teardownTestDB()

	os.Exit(code)
}

// setupTestDB 设置测试数据库
func setupTestDB() {
	// 使用测试数据库
	testDBName := os.Getenv("NEVERBROKE_TEST_DB")
	if testDBName == "" {
		testDBName = "neverbroke_test"
	}

	// 构建测试数据库配置
	cfg := dbconfig.DBConfig{
		DSN: fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
			getEnvOrDefault("NEVERBROKE_PG_USER", "neverbroke"),
			getEnvOrDefault("NEVERBROKE_PG_PASSWORD", "neverbroke"),
			getEnvOrDefault("NEVERBROKE_PG_HOST", "127.0.0.1"),
			getEnvOrDefault("NEVERBROKE_PG_PORT", "5432"),
			testDBName,
		),
	}

	var err error
	testDB, err = postgres.OpenGorm(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to test database: %v", err))
	}

	// 自动迁移表结构
	if err := autoMigrate(testDB); err != nil {
		panic(fmt.Sprintf("Failed to migrate test database: %v", err))
	}

	if err := payroll.SeedPolicyHistory(testDB); err != nil {
		panic(err)
	}
	// 创建存储和路由
	testStore = postgres.NewPgStore(testDB)
	testRouter = router.Setup(testStore, testDB)

	logger.Info("Test database setup completed")
}

// teardownTestDB 清理测试数据库
func teardownTestDB() {
	if testDB != nil {
		// 清理所有测试数据
		cleanupTestData(testDB)

		// 关闭连接
		sqlDB, err := testDB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// cleanupTestData 清理测试数据
func cleanupTestData(db *gorm.DB) {
	// 按照外键依赖顺序删除数据
	db.Exec("TRUNCATE TABLE transactions CASCADE")
	db.Exec("TRUNCATE TABLE installment_plans CASCADE")
	db.Exec("TRUNCATE TABLE recurring_income_plans CASCADE")
	db.Exec("TRUNCATE TABLE accounts CASCADE")
	db.Exec("TRUNCATE TABLE refresh_tokens CASCADE")
	db.Exec("TRUNCATE TABLE oauth_identities CASCADE")
	db.Exec("TRUNCATE TABLE oauth_states CASCADE")
	db.Exec("TRUNCATE TABLE users CASCADE")
}

// autoMigrate 自动迁移数据库表
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&auth.User{},
		&auth.RefreshToken{},
		&auth.OAuthIdentity{},
		&auth.OAuthState{},
		&models.Account{},
		&models.Transaction{},
		&models.InstallmentPlan{},
		&models.RecurringIncomePlan{},
		&models.Budget{},
		&models.CuratedImportRecord{}, &workspace.State{}, &workspace.Change{}, &workspace.Setting{}, &workspace.Batch{}, &workspace.SourceAlias{}, &workspace.Attachment{}, &workspace.SyncRun{},
	)
}

// getEnvOrDefault 获取环境变量或默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// setupTestTransaction 开启测试事务（用于每个测试用例）
func setupTestTransaction(t *testing.T) *gorm.DB {
	tx := testDB.Begin()
	t.Cleanup(func() {
		tx.Rollback()
	})
	return tx
}

// waitForDB 等待数据库操作完成（模拟异步情况）
func waitForDB() {
	time.Sleep(100 * time.Millisecond)
}
