package main

import (
	"context"
	"fmt"
	"log"
	"neverbroke/internal/accounts"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/repository/postgres"
	"os"

	"github.com/google/uuid"
)

func main() {
	// 加载 .env 文件
	loadEnvFile("../../.env")

	// 加载数据库配置
	dbCfg := dbconfig.LoadFromEnv()
	fmt.Printf("Connecting to database...\n")

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

	// 创建仓储
	store := postgres.NewPgStore(db)
	ctx := context.Background()
	userID := "test-user-" + uuid.NewString()[:8]

	// 测试创建账户
	fmt.Println("\n📝 Testing account creation...")
	testAccount := &accounts.Account{
		ID:       uuid.NewString(),
		Type:     accounts.AccountTypeCash,
		Currency: "CNY",
		Balance:  1000.00,
		Archived: false,
	}

	err = store.Accounts().Create(ctx, userID, testAccount)
	if err != nil {
		log.Fatalf("Failed to create account: %v", err)
	}
	fmt.Printf("✅ Account created: %s\n", testAccount.ID)

	// 测试查询账户
	fmt.Println("\n📝 Testing account retrieval...")
	retrievedAccount, err := store.Accounts().GetByID(ctx, userID, testAccount.ID)
	if err != nil {
		log.Fatalf("Failed to get account: %v", err)
	}
	if retrievedAccount == nil {
		log.Fatal("Account not found")
	}
	fmt.Printf("✅ Account retrieved: ID=%s, Type=%s, Balance=%.2f\n",
		retrievedAccount.ID, retrievedAccount.Type, retrievedAccount.Balance)

	// 测试列出账户
	fmt.Println("\n📝 Testing account listing...")
	accounts, err := store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		log.Fatalf("Failed to list accounts: %v", err)
	}
	fmt.Printf("✅ Found %d account(s) for user %s\n", len(accounts), userID)

	// 测试更新账户
	fmt.Println("\n📝 Testing account update...")
	testAccount.Balance = 2000.00
	err = store.Accounts().Update(ctx, userID, testAccount)
	if err != nil {
		log.Fatalf("Failed to update account: %v", err)
	}
	fmt.Println("✅ Account updated")

	// 验证更新
	updatedAccount, err := store.Accounts().GetByID(ctx, userID, testAccount.ID)
	if err != nil {
		log.Fatalf("Failed to get updated account: %v", err)
	}
	fmt.Printf("✅ Updated balance: %.2f\n", updatedAccount.Balance)

	fmt.Println("\n🎉 All GORM tests passed!")
	fmt.Println("\n⚠️  Note: This test created a test account in the database.")
	fmt.Printf("   User ID: %s\n", userID)
	fmt.Printf("   Account ID: %s\n", testAccount.ID)
	fmt.Println("   You may want to clean up test data manually.")
}

// loadEnvFile 加载 .env 文件中的环境变量
func loadEnvFile(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	lines := splitLines(string(data))
	for _, line := range lines {
		line = trimSpace(line)
		if line == "" || hasPrefix(line, "#") {
			continue
		}
		parts := splitN(line, "=", 2)
		if len(parts) == 2 {
			key := trimSpace(parts[0])
			value := trimSpace(parts[1])
			os.Setenv(key, value)
		}
	}
}

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
