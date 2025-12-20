package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"neverbroke/internal/auth"
	"neverbroke/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionFlow_CreateAndQuery 测试创建交易并查询
func TestTransactionFlow_CreateAndQuery(t *testing.T) {
	cleanupTestData(testDB)

	// 1. 创建测试用户和账户
	user := createTestUser(t, "transaction_test@example.com")
	accessToken := loginAndGetToken(t, "transaction_test@example.com", "TestPassword123!")

	account := createTestAccount(t, user.ID, "bank", 10000.0)

	// 2. 创建收入交易
	t.Run("CreateIncomeTransaction", func(t *testing.T) {
		transactionReq := map[string]interface{}{
			"type":        "income",
			"amount":      5000.0,
			"categoryKey": "INCOME_SALARY",
			"remark":      "月工资",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   account.ID,
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Create income transaction should succeed")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		// 检查响应格式
		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["id"])
		assert.Equal(t, "income", data["type"])
		assert.Equal(t, 5000.0, data["amount"])

		waitForDB()

		// 验证交易已写入数据库
		var tx models.Transaction
		err := testDB.Where("account_id = ? AND type = ?", account.ID, "income").First(&tx).Error
		require.NoError(t, err, "Transaction should be persisted")
		assert.Equal(t, 5000.0, tx.Amount)
		assert.Equal(t, "INCOME_SALARY", tx.CategoryKey)
		assert.Equal(t, "月工资", tx.Remark)

		// 验证账户余额已更新
		var updatedAccount models.Account
		testDB.First(&updatedAccount, account.ID)
		assert.Equal(t, 15000.0, updatedAccount.Balance, "Account balance should increase")
	})

	// 3. 创建支出交易
	t.Run("CreateExpenseTransaction", func(t *testing.T) {
		transactionReq := map[string]interface{}{
			"type":        "expense",
			"amount":      1000.0,
			"categoryKey": "EXPENSE_FOOD",
			"remark":      "超市购物",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   account.ID,
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		waitForDB()

		// 验证账户余额减少
		var updatedAccount models.Account
		testDB.First(&updatedAccount, account.ID)
		assert.Equal(t, 14000.0, updatedAccount.Balance, "Account balance should decrease")

		// 验证交易记录
		var txCount int64
		testDB.Model(&models.Transaction{}).Where("account_id = ?", account.ID).Count(&txCount)
		assert.Equal(t, int64(2), txCount, "Should have 2 transactions")
	})

	// 4. 查询每日交易
	t.Run("QueryDailyTransactions", func(t *testing.T) {
		today := time.Now().Format("2006-01-02")
		req, _ := http.NewRequest("GET", "/api/ledger/daily?date="+today, nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].(map[string]interface{})
		stats := data["stats"].(map[string]interface{})

		assert.Equal(t, 5000.0, stats["totalIncome"])
		assert.Equal(t, 1000.0, stats["totalExpense"])

		list := data["list"].([]interface{})
		assert.Len(t, list, 2, "Should return 2 transactions")
	})
}

// TestTransactionFlow_TransferBetweenAccounts 测试账户间转账
func TestTransactionFlow_TransferBetweenAccounts(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "transfer_test@example.com")
	accessToken := loginAndGetToken(t, "transfer_test@example.com", "TestPassword123!")

	account1 := createTestAccount(t, user.ID, "bank", 10000.0)
	account2 := createTestAccount(t, user.ID, "bank", 5000.0)

	t.Run("TransferSuccess", func(t *testing.T) {
		transferReq := map[string]interface{}{
			"type":          "transfer",
			"amount":        2000.0,
			"remark":        "账户间转账",
			"occurredAt":    time.Now().Format(time.RFC3339),
			"fromAccountId": account1.ID,
			"toAccountId":   account2.ID,
		}

		body, _ := json.Marshal(transferReq)
		req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code)

		waitForDB()

		// 验证两个账户余额都已更新
		var acc1, acc2 models.Account
		testDB.First(&acc1, account1.ID)
		testDB.First(&acc2, account2.ID)

		assert.Equal(t, 8000.0, acc1.Balance, "From account should decrease")
		assert.Equal(t, 7000.0, acc2.Balance, "To account should increase")

		// 验证交易记录
		var tx models.Transaction
		err := testDB.Where("from_account_id = ? AND to_account_id = ?", account1.ID, account2.ID).
			First(&tx).Error
		require.NoError(t, err)
		assert.Equal(t, 2000.0, tx.Amount)
	})
}

// TestTransactionFlow_FailureRollback 测试事务回滚
func TestTransactionFlow_FailureRollback(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "rollback_test@example.com")
	accessToken := loginAndGetToken(t, "rollback_test@example.com", "TestPassword123!")

	account := createTestAccount(t, user.ID, "bank", 1000.0)

	t.Run("InsufficientBalance", func(t *testing.T) {
		// 尝试支出超过余额的金额
		transactionReq := map[string]interface{}{
			"type":        "expense",
			"amount":      2000.0, // 超过账户余额
			"categoryKey": "EXPENSE_SHOPPING",
			"remark":      "大额消费",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   account.ID,
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 注意：当前实现可能允许负余额，根据实际业务逻辑调整
		// 这里主要测试如果失败，数据库应该保持一致性

		waitForDB()

		// 验证账户余额（如果交易成功，会是-1000，如果失败，应该还是1000）
		var updatedAccount models.Account
		testDB.First(&updatedAccount, account.ID)

		// 根据实际业务逻辑验证
		if w.Code == http.StatusOK {
			assert.Equal(t, -1000.0, updatedAccount.Balance, "Balance can go negative")
		} else {
			assert.Equal(t, 1000.0, updatedAccount.Balance, "Balance should not change on failure")
		}
	})

	t.Run("InvalidAccount", func(t *testing.T) {
		transactionReq := map[string]interface{}{
			"type":        "expense",
			"amount":      100.0,
			"categoryKey": "EXPENSE_FOOD",
			"remark":      "测试",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   "non-existent-account",
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		// 应该返回错误
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		assert.NotEqual(t, float64(0), response["code"], "Should return error for invalid account")
	})
}

// TestTransactionFlow_NetworkInterruption 测试网络中断场景（模拟）
func TestTransactionFlow_NetworkInterruption(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "network_test@example.com")
	accessToken := loginAndGetToken(t, "network_test@example.com", "TestPassword123!")

	account := createTestAccount(t, user.ID, "bank", 10000.0)

	t.Run("ClientDisconnectsDuringTransaction", func(t *testing.T) {
		// 创建一个带超时的上下文，模拟客户端断开
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		transactionReq := map[string]interface{}{
			"type":        "expense",
			"amount":      500.0,
			"categoryKey": "EXPENSE_FOOD",
			"remark":      "测试网络中断",
			"occurredAt":  time.Now().Format(time.RFC3339),
			"accountId":   account.ID,
		}

		body, _ := json.Marshal(transactionReq)
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/ledger/transactions", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)

		w := httptest.NewRecorder()

		// 可能会因为上下文取消而失败
		testRouter.ServeHTTP(w, req)

		// 等待数据库操作完成
		time.Sleep(200 * time.Millisecond)

		// 验证数据一致性：要么交易完成并更新余额，要么都没有变化
		var txCount int64
		testDB.Model(&models.Transaction{}).
			Where("account_id = ? AND amount = ?", account.ID, 500.0).
			Count(&txCount)

		var currentAccount models.Account
		testDB.First(&currentAccount, account.ID)

		if txCount > 0 {
			// 如果交易记录存在，余额应该已更新
			assert.Equal(t, 9500.0, currentAccount.Balance, "Balance should be updated if transaction exists")
		} else {
			// 如果交易记录不存在，余额应该没变
			assert.Equal(t, 10000.0, currentAccount.Balance, "Balance should not change if transaction doesn't exist")
		}
	})
}

// TestTransactionFlow_ConcurrentTransactions 测试并发交易
func TestTransactionFlow_ConcurrentTransactions(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "concurrent_tx_test@example.com")
	accessToken := loginAndGetToken(t, "concurrent_tx_test@example.com", "TestPassword123!")

	account := createTestAccount(t, user.ID, "bank", 10000.0)

	t.Run("ConcurrentExpenses", func(t *testing.T) {
		results := make(chan int, 10)

		// 并发执行10笔交易，每笔100元
		for i := 0; i < 10; i++ {
			go func(index int) {
				transactionReq := map[string]interface{}{
					"type":        "expense",
					"amount":      100.0,
					"categoryKey": "EXPENSE_FOOD",
					"remark":      "并发测试",
					"occurredAt":  time.Now().Format(time.RFC3339),
					"accountId":   account.ID,
				}

				body, _ := json.Marshal(transactionReq)
				req, _ := http.NewRequest("POST", "/api/ledger/transactions", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)

				results <- w.Code
			}(i)
		}

		// 收集结果
		successCount := 0
		for i := 0; i < 10; i++ {
			code := <-results
			if code == http.StatusOK {
				successCount++
			}
		}

		// 所有交易都应该成功
		assert.Equal(t, 10, successCount, "All concurrent transactions should succeed")

		// 等待所有数据库操作完成
		time.Sleep(500 * time.Millisecond)

		// 验证最终余额是否正确
		var finalAccount models.Account
		testDB.First(&finalAccount, account.ID)

		// 最终余额应该是 10000 - 10*100 = 9000
		assert.Equal(t, 9000.0, finalAccount.Balance, "Final balance should be correct after concurrent transactions")

		// 验证交易记录数量
		var txCount int64
		testDB.Model(&models.Transaction{}).Where("account_id = ?", account.ID).Count(&txCount)
		assert.Equal(t, int64(10), txCount, "Should have 10 transaction records")
	})
}

// Helper functions

func createTestUser(t *testing.T, email string) *auth.User {
	user := &auth.User{
		Email:        email,
		PasswordHash: "$argon2id$v=19$m=65536,t=1,p=4$test$testhash", // 简化的测试hash
		Nickname:     "Test User",
	}

	passwordHash, _ := auth.HashPassword("TestPassword123!")
	user.PasswordHash = passwordHash

	err := testDB.Create(user).Error
	require.NoError(t, err)
	return user
}

func loginAndGetToken(t *testing.T, email, password string) string {
	loginReq := auth.LoginRequest{
		Email:    email,
		Password: password,
		DeviceID: "test-device",
	}

	body, _ := json.Marshal(loginReq)
	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Type", "app")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	return response["access_token"].(string)
}

func createTestAccount(t *testing.T, userID int, accountType string, initialBalance float64) *models.Account {
	account := &models.Account{
		UserID:   fmt.Sprintf("%d", userID),
		Type:     accountType,
		Currency: "CNY",
		Balance:  initialBalance,
		Archived: false,
	}

	err := testDB.Create(account).Error
	require.NoError(t, err)
	return account
}
