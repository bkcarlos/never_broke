package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"neverbroke/internal/auth"
	"neverbroke/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultiUser_SameUserDifferentDevices 测试同一用户从不同设备登录和操作
func TestMultiUser_SameUserDifferentDevices(t *testing.T) {
	cleanupTestData(testDB)

	email := "multi_device_user@example.com"
	password := "TestPassword123!"

	// 创建用户和账户
	user := createTestUser(t, email)
	account := createTestAccount(t, user.ID, "bank", 10000.0)

	t.Run("SameUserLoginFromMultipleDevices", func(t *testing.T) {
		devices := []string{"iPhone", "iPad", "MacBook", "Web", "Android"}
		tokens := make(map[string]string)
		var wg sync.WaitGroup

		// 同一用户从5个不同设备并发登录
		for _, device := range devices {
			wg.Add(1)
			go func(deviceID string) {
				defer wg.Done()

				loginReq := auth.LoginRequest{
					Email:    email,
					Password: password,
					DeviceID: deviceID,
				}

				body, _ := json.Marshal(loginReq)
				req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Client-Type", "app")
				req.Header.Set("User-Agent", deviceID)

				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)

				if w.Code == http.StatusOK {
					var response map[string]interface{}
					json.Unmarshal(w.Body.Bytes(), &response)
					tokens[deviceID] = response["access_token"].(string)
				}
			}(device)
		}

		wg.Wait()
		waitForDB()

		// 验证所有设备都成功登录
		assert.Len(t, tokens, 5, "All 5 devices should login successfully")

		// 验证数据库中有5个活跃的 refresh token
		var tokenCount int64
		testDB.Model(&auth.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", user.ID).
			Count(&tokenCount)
		assert.Equal(t, int64(5), tokenCount, "Should have 5 active refresh tokens")

		// 验证每个设备的 token 都是不同的
		uniqueTokens := make(map[string]bool)
		for _, token := range tokens {
			uniqueTokens[token] = true
		}
		assert.Len(t, uniqueTokens, 5, "All tokens should be unique")
	})

	t.Run("MultiDevicesConcurrentOperations", func(t *testing.T) {
		// 重新登录获取最新的 tokens
		device1Token := loginAndGetTokenWithDevice(t, email, password, "device1")
		device2Token := loginAndGetTokenWithDevice(t, email, password, "device2")
		device3Token := loginAndGetTokenWithDevice(t, email, password, "device3")

		waitForDB()

		// 获取初始余额
		var initialAccount models.Account
		testDB.Where("id = ?", account.ID).First(&initialAccount)
		initialBalance := initialAccount.Balance

		var wg sync.WaitGroup
		results := make(chan bool, 9)

		// Device 1: 3笔支出交易
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				success := createTransaction(t, device1Token, account.ID, "expense", 100.0,
					fmt.Sprintf("Device1 支出 #%d", index))
				results <- success
			}(i)
		}

		// Device 2: 2笔收入交易
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				success := createTransaction(t, device2Token, account.ID, "income", 500.0,
					fmt.Sprintf("Device2 收入 #%d", index))
				results <- success
			}(i)
		}

		// Device 3: 4笔支出交易
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				success := createTransaction(t, device3Token, account.ID, "expense", 50.0,
					fmt.Sprintf("Device3 支出 #%d", index))
				results <- success
			}(i)
		}

		wg.Wait()
		close(results)

		// 统计成功的交易数
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		assert.Equal(t, 9, successCount, "All 9 transactions should succeed")

		// 等待所有数据库操作完成
		time.Sleep(500 * time.Millisecond)

		// 验证最终余额
		var finalAccount models.Account
		testDB.Where("id = ?", account.ID).First(&finalAccount)

		// 计算期望余额：初始余额 - (3*100 + 4*50) + (2*500) = 初始余额 - 500 + 1000 = 初始余额 + 500
		expectedBalance := initialBalance + 500.0
		assert.Equal(t, expectedBalance, finalAccount.Balance,
			"Final balance should be correct: initial %.2f + 500 = %.2f, got %.2f",
			initialBalance, expectedBalance, finalAccount.Balance)

		// 验证交易记录数
		var txCount int64
		testDB.Model(&models.Transaction{}).Where("account_id = ?", account.ID).Count(&txCount)
		assert.GreaterOrEqual(t, txCount, int64(9), "Should have at least 9 transaction records")
	})

	t.Run("OneDeviceLogoutOthersStillWork", func(t *testing.T) {
		// 3个设备登录
		device1Token := loginAndGetTokenWithDevice(t, email, password, "logout-device-1")
		device2Token := loginAndGetTokenWithDevice(t, email, password, "logout-device-2")
		device3Token := loginAndGetTokenWithDevice(t, email, password, "logout-device-3")

		waitForDB()

		// Device 1 登出
		logoutReq := auth.LogoutRequest{DeviceID: "logout-device-1"}
		body, _ := json.Marshal(logoutReq)
		req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+device1Token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Device 1 logout should succeed")

		waitForDB()

		// Device 1 的 token 应该无法使用
		tx1Success := createTransaction(t, device1Token, account.ID, "expense", 10.0, "Device1 after logout")
		assert.False(t, tx1Success, "Device 1 should not work after logout")

		// Device 2 和 3 应该仍然可以使用
		tx2Success := createTransaction(t, device2Token, account.ID, "expense", 10.0, "Device2 still working")
		tx3Success := createTransaction(t, device3Token, account.ID, "expense", 10.0, "Device3 still working")

		assert.True(t, tx2Success, "Device 2 should still work")
		assert.True(t, tx3Success, "Device 3 should still work")
	})
}

// TestMultiUser_DifferentUsersConcurrentOperations 测试多个不同用户并发操作
func TestMultiUser_DifferentUsersConcurrentOperations(t *testing.T) {
	cleanupTestData(testDB)

	// 创建3个用户，每个用户有自己的账户
	users := []struct {
		email   string
		token   string
		userID  int
		account *models.Account
	}{
		{email: "user1@example.com"},
		{email: "user2@example.com"},
		{email: "user3@example.com"},
	}

	for i := range users {
		user := createTestUser(t, users[i].email)
		users[i].userID = user.ID
		users[i].token = loginAndGetToken(t, users[i].email, "TestPassword123!")
		users[i].account = createTestAccount(t, user.ID, "bank", 5000.0)
	}

	t.Run("MultipleUsersCreateTransactionsConcurrently", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan string, 30)

		// 每个用户并发创建10笔交易
		for _, user := range users {
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(userEmail, token, accountID string, index int) {
					defer wg.Done()

					success := createTransaction(t, token, accountID, "expense", 100.0,
						fmt.Sprintf("%s 交易 #%d", userEmail, index))

					if success {
						results <- fmt.Sprintf("%s-success", userEmail)
					} else {
						results <- fmt.Sprintf("%s-fail", userEmail)
					}
				}(user.email, user.token, user.account.ID, i)
			}
		}

		wg.Wait()
		close(results)

		// 统计每个用户的成功数
		successCounts := make(map[string]int)
		for result := range results {
			if len(result) > 8 && result[len(result)-7:] == "success" {
				email := result[:len(result)-8]
				successCounts[email]++
			}
		}

		// 验证每个用户都成功创建了10笔交易
		for _, user := range users {
			assert.Equal(t, 10, successCounts[user.email],
				"User %s should have 10 successful transactions", user.email)
		}

		time.Sleep(500 * time.Millisecond)

		// 验证每个用户的余额独立更新
		for _, user := range users {
			var account models.Account
			testDB.Where("id = ?", user.account.ID).First(&account)
			expectedBalance := 5000.0 - 1000.0 // 初始5000 - 10*100
			assert.Equal(t, expectedBalance, account.Balance,
				"User %s balance should be correct", user.email)
		}

		// 验证交易记录隔离（每个用户只能看到自己的交易）
		for _, user := range users {
			var txCount int64
			testDB.Model(&models.Transaction{}).
				Where("account_id = ?", user.account.ID).
				Count(&txCount)
			assert.Equal(t, int64(10), txCount,
				"User %s should have exactly 10 transaction records", user.email)
		}
	})

	t.Run("UsersCannotAccessOthersData", func(t *testing.T) {
		// User1 尝试操作 User2 的账户
		user1Token := users[0].token
		user2AccountID := users[1].account.ID

		// 尝试创建交易到别人的账户（应该失败或被拒绝）
		transactionReq := map[string]interface{}{
			"type":        "expense",
			"amount":      100.0,
			"categoryKey": "EXPENSE_FOOD",
			"remark":      "尝试访问别人账户",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   user2AccountID,
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+user1Token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 注意：根据你的实际权限验证逻辑，这可能返回 403 或 400
		// 这里主要验证 User2 的账户余额没有被改变
		time.Sleep(100 * time.Millisecond)

		var user2Account models.Account
		testDB.Where("id = ?", user2AccountID).First(&user2Account)
		assert.Equal(t, 4000.0, user2Account.Balance,
			"User2's balance should not change when User1 tries to access it")
	})
}

// TestMultiUser_SharedResourceConflict 测试共享资源冲突场景
func TestMultiUser_SharedResourceConflict(t *testing.T) {
	cleanupTestData(testDB)

	// 创建一个用户和账户
	user := createTestUser(t, "shared_resource@example.com")
	account := createTestAccount(t, user.ID, "bank", 1000.0)

	// 从两个不同设备登录
	token1 := loginAndGetTokenWithDevice(t, "shared_resource@example.com", "TestPassword123!", "device-1")
	token2 := loginAndGetTokenWithDevice(t, "shared_resource@example.com", "TestPassword123!", "device-2")

	t.Run("SimultaneousWithdrawalsFromSameAccount", func(t *testing.T) {
		// 两个设备同时尝试取款600元（账户只有1000元，两笔都成功的话会导致负余额）
		var wg sync.WaitGroup
		results := make(chan bool, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			success := createTransaction(t, token1, account.ID, "expense", 600.0, "Device1 大额支出")
			results <- success
		}()

		go func() {
			defer wg.Done()
			success := createTransaction(t, token2, account.ID, "expense", 600.0, "Device2 大额支出")
			results <- success
		}()

		wg.Wait()
		close(results)

		// 收集结果
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		time.Sleep(300 * time.Millisecond)

		// 验证最终状态一致性
		var finalAccount models.Account
		testDB.Where("id = ?", account.ID).First(&finalAccount)

		// 根据业务逻辑验证：
		// 选项1: 两笔都成功，允许负余额（余额 = -200）
		// 选项2: 只有一笔成功（余额 = 400）
		// 选项3: 都失败（余额 = 1000）

		if successCount == 2 {
			assert.Equal(t, -200.0, finalAccount.Balance, "Both withdrawals succeeded, balance should be -200")
		} else if successCount == 1 {
			assert.Equal(t, 400.0, finalAccount.Balance, "Only one withdrawal succeeded, balance should be 400")
		} else {
			assert.Equal(t, 1000.0, finalAccount.Balance, "Both failed, balance should remain 1000")
		}

		// 验证交易记录数与成功数一致
		var txCount int64
		testDB.Model(&models.Transaction{}).
			Where("account_id = ? AND amount = 600.0", account.ID).
			Count(&txCount)
		assert.Equal(t, int64(successCount), txCount, "Transaction count should match success count")
	})
}

// TestMultiUser_TokenRefreshRace 测试多设备同时刷新 Token 的竞态条件
func TestMultiUser_TokenRefreshRace(t *testing.T) {
	cleanupTestData(testDB)

	email := "token_refresh_user@example.com"
	user := createTestUser(t, email)

	// 登录获取初始 token
	loginReq := auth.LoginRequest{
		Email:    email,
		Password: "TestPassword123!",
		DeviceID: "refresh-device",
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Type", "app")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	refreshToken := loginResponse["refresh_token"].(string)

	waitForDB()

	t.Run("ConcurrentTokenRefresh", func(t *testing.T) {
		var wg sync.WaitGroup
		results := make(chan int, 5)

		// 5个线程同时使用同一个 refresh token 刷新
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()

				refreshReq := auth.RefreshRequest{
					RefreshToken: refreshToken,
				}

				body, _ := json.Marshal(refreshReq)
				req, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Client-Type", "app")

				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)

				results <- w.Code
			}()
		}

		wg.Wait()
		close(results)

		// 统计结果
		successCount := 0
		failCount := 0
		for code := range results {
			if code == http.StatusOK {
				successCount++
			} else {
				failCount++
			}
		}

		// 由于 Token Rotation，只有一个应该成功，其他应该失败
		// （或者根据实际实现，可能有不同的行为）
		t.Logf("Success: %d, Failed: %d", successCount, failCount)

		// 至少验证不会产生多个有效的新 token
		var activeTokens int64
		testDB.Model(&auth.RefreshToken{}).
			Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", user.ID, "refresh-device").
			Count(&activeTokens)

		assert.LessOrEqual(t, activeTokens, int64(2),
			"Should not have too many active tokens after concurrent refresh")
	})
}

// Helper: 创建交易（返回是否成功）
func createTransaction(t *testing.T, token string, accountID string, txType string, amount float64, remark string) bool {
	transactionReq := map[string]interface{}{
		"type":        txType,
		"amount":      amount,
		"categoryKey": "EXPENSE_FOOD",
		"remark":      remark,
		"occurredAt":  time.Now().Format(time.RFC3339),
		"accountId":   accountID,
	}

	body, _ := json.Marshal(transactionReq)
	req, _ := http.NewRequest("POST", "/api/transactions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	return w.Code == http.StatusOK
}

// Helper: 登录并获取指定设备的 token
func loginAndGetTokenWithDevice(t *testing.T, email, password, deviceID string) string {
	loginReq := auth.LoginRequest{
		Email:    email,
		Password: password,
		DeviceID: deviceID,
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Type", "app")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Login failed for %s on device %s: %d", email, deviceID, w.Code)
		return ""
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["access_token"].(string)
}
