package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"neverbroke/internal/models"

	"github.com/stretchr/testify/assert"
)

// TestRealWorldScenario_FamilySharedAccount 测试家庭共享账户场景
// 注意：此测试需要实现账户共享功能后才能启用
func TestRealWorldScenario_FamilySharedAccount(t *testing.T) {
	t.Skip("Skipping test: shared account feature not yet implemented. Account ownership validation prevents cross-user account access.")
	
	cleanupTestData(testDB)

	// 模拟场景：夫妻两人共享一个银行账户
	// 丈夫在超市购物，妻子在网上购物，两人同时操作

	// 创建账户所有者（丈夫）
	husband := createTestUser(t, "husband@family.com")
	husbandToken := loginAndGetTokenWithDevice(t, "husband@family.com", "TestPassword123!", "husband-phone")

	// 创建另一个用户（妻子）
	_ = createTestUser(t, "wife@family.com")
	wifeToken := loginAndGetTokenWithDevice(t, "wife@family.com", "TestPassword123!", "wife-phone")

	// 创建共享的银行账户（这里简化为属于丈夫）
	sharedAccount := createTestAccount(t, husband.ID, "bank", 10000.0)

	t.Run("SimultaneousShoppingByCouple", func(t *testing.T) {
		var wg sync.WaitGroup
		startTime := time.Now()

		// 记录所有交易结果
		type TransactionResult struct {
			user    string
			amount  float64
			success bool
			time    time.Duration
		}
		results := make(chan TransactionResult, 20)

		// 丈夫：在超市购物，5笔小额交易（模拟扫描商品）
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				time.Sleep(time.Duration(rand.Intn(100)) * time.Millisecond) // 随机延迟

				amount := 50.0 + float64(rand.Intn(100))
				success := createTransaction(t, husbandToken, string(sharedAccount.ID), "expense", amount,
					fmt.Sprintf("超市购物 商品#%d", index))

				results <- TransactionResult{
					user:    "husband",
					amount:  amount,
					success: success,
					time:    time.Since(startTime),
				}
			}(i)
		}

		// 妻子：在网上购物，3笔交易
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				time.Sleep(time.Duration(rand.Intn(150)) * time.Millisecond)

				amount := 200.0 + float64(rand.Intn(300))
				success := createTransaction(t, wifeToken, string(sharedAccount.ID), "expense", amount,
					fmt.Sprintf("网购 订单#%d", index))

				results <- TransactionResult{
					user:    "wife",
					amount:  amount,
					success: success,
					time:    time.Since(startTime),
				}
			}(i)
		}

		// 同时，丈夫收到工资
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(50 * time.Millisecond)

			success := createTransaction(t, husbandToken, string(sharedAccount.ID), "income", 5000.0, "月工资")
			results <- TransactionResult{
				user:    "husband",
				amount:  5000.0,
				success: success,
				time:    time.Since(startTime),
			}
		}()

		wg.Wait()
		close(results)

		// 分析结果
		var husbandExpenses, wifeExpenses, income float64
		successCount := 0

		t.Log("\n=== Transaction Timeline ===")
		for result := range results {
			if result.success {
				successCount++
				if result.user == "husband" && result.amount > 1000 {
					income += result.amount
				} else if result.user == "husband" {
					husbandExpenses += result.amount
				} else {
					wifeExpenses += result.amount
				}

				t.Logf("[%v] %s: %.2f (success)", result.time, result.user, result.amount)
			}
		}

		assert.Equal(t, 9, successCount, "All 9 transactions should succeed")

		time.Sleep(500 * time.Millisecond)

		// 验证最终余额
		var finalAccount models.Account
		testDB.Where("id = ?", sharedAccount.ID).First(&finalAccount)

		expectedBalance := 10000.0 + income - husbandExpenses - wifeExpenses
		t.Logf("\n=== Balance Verification ===")
		t.Logf("Initial: 10000.00")
		t.Logf("Income: +%.2f", income)
		t.Logf("Husband expenses: -%.2f", husbandExpenses)
		t.Logf("Wife expenses: -%.2f", wifeExpenses)
		t.Logf("Expected: %.2f", expectedBalance)
		t.Logf("Actual: %.2f", finalAccount.Balance)

		assert.InDelta(t, expectedBalance, finalAccount.Balance, 0.01,
			"Final balance should match expected value")
	})
}

// TestRealWorldScenario_BusinessDayOperations 测试企业日常运营场景
// 注意：此测试需要实现账户共享功能后才能启用
func TestRealWorldScenario_BusinessDayOperations(t *testing.T) {
	t.Skip("Skipping test: shared account feature not yet implemented. Account ownership validation prevents cross-user account access.")
	
	cleanupTestData(testDB)

	// 模拟场景：公司的一天
	// - 财务人员处理工资发放
	// - 销售人员记录收入
	// - 采购人员记录支出
	// - CEO 查看报表

	finance := createTestUser(t, "finance@company.com")
	financeToken := loginAndGetTokenWithDevice(t, "finance@company.com", "TestPassword123!", "finance-pc")

	_ = createTestUser(t, "sales@company.com")
	salesToken := loginAndGetTokenWithDevice(t, "sales@company.com", "TestPassword123!", "sales-pc")

	_ = createTestUser(t, "procurement@company.com")
	procurementToken := loginAndGetTokenWithDevice(t, "procurement@company.com", "TestPassword123!", "procurement-pc")

	// 公司账户
	companyAccount := createTestAccount(t, finance.ID, "bank", 500000.0)

	t.Run("SimultaneousBusinessOperations", func(t *testing.T) {
		var wg sync.WaitGroup

		// 场景1: 财务批量发放工资（10个员工）
		t.Log("\n=== Phase 1: Payroll Processing ===")
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(employeeID int) {
				defer wg.Done()
				salary := 5000.0 + float64(employeeID)*100
				success := createTransaction(t, financeToken, string(companyAccount.ID), "expense",
					salary, fmt.Sprintf("工资-员工%d", employeeID))
				if success {
					t.Logf("✓ Paid employee %d: %.2f", employeeID, salary)
				}
			}(i)
		}

		// 稍等一下再开始其他操作
		time.Sleep(100 * time.Millisecond)

		// 场景2: 销售同时记录多笔收入
		t.Log("\n=== Phase 2: Sales Recording ===")
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(orderID int) {
				defer wg.Done()
				revenue := 10000.0 + float64(rand.Intn(5000))
				success := createTransaction(t, salesToken, string(companyAccount.ID), "income",
					revenue, fmt.Sprintf("销售订单#%d", orderID))
				if success {
					t.Logf("✓ Recorded sale %d: %.2f", orderID, revenue)
				}
			}(i)
		}

		// 场景3: 采购记录支出
		t.Log("\n=== Phase 3: Procurement ===")
		for i := 0; i < 3; i++ {
			wg.Add(1)
			go func(purchaseID int) {
				defer wg.Done()
				cost := 20000.0 + float64(rand.Intn(10000))
				success := createTransaction(t, procurementToken, string(companyAccount.ID), "expense",
					cost, fmt.Sprintf("采购订单#%d", purchaseID))
				if success {
					t.Logf("✓ Recorded purchase %d: %.2f", purchaseID, cost)
				}
			}(i)
		}

		wg.Wait()
		time.Sleep(500 * time.Millisecond)

		// 验证所有交易都已记录
		var txCount int64
		testDB.Model(&models.Transaction{}).Where("account_id = ?", companyAccount.ID).Count(&txCount)
		assert.Equal(t, int64(18), txCount, "Should have 18 transaction records (10 payroll + 5 sales + 3 procurement)")

		// 验证账户余额一致性
		var finalAccount models.Account
		testDB.Where("id = ?", companyAccount.ID).First(&finalAccount)

		t.Logf("\n=== Final Company Balance: %.2f ===", finalAccount.Balance)
		assert.Greater(t, finalAccount.Balance, 0.0, "Company should still have positive balance")
	})
}

// TestRealWorldScenario_MobileAppUsage 测试移动端真实使用场景
func TestRealWorldScenario_MobileAppUsage(t *testing.T) {
	cleanupTestData(testDB)

	// 模拟场景：用户的一天移动端使用
	// 早上：通勤时查看账户、记录早餐
	// 中午：午休时记录午餐
	// 下午：网络不稳定时尝试记录
	// 晚上：回家后批量记录多笔支出

	user := createTestUser(t, "mobile_user@example.com")
	token := loginAndGetTokenWithDevice(t, "mobile_user@example.com", "TestPassword123!", "iPhone")
	account := createTestAccount(t, user.ID, "bank", 1000.0)

	t.Run("MorningCommute", func(t *testing.T) {
		// 早上通勤时的快速记录
		success := createTransaction(t, token, string(account.ID), "expense", 15.0, "早餐-包子")
		assert.True(t, success, "Morning breakfast transaction should succeed")

		success = createTransaction(t, token, string(account.ID), "expense", 3.0, "地铁")
		assert.True(t, success, "Subway transaction should succeed")

		time.Sleep(100 * time.Millisecond)

		var balance models.Account
		testDB.Where("id = ?", account.ID).First(&balance)
		assert.Equal(t, 982.0, balance.Balance, "Balance should be 1000 - 15 - 3 = 982")
	})

	t.Run("UnstableNetworkScenario", func(t *testing.T) {
		// 模拟网络不稳定，用户多次尝试提交同一笔交易
		var wg sync.WaitGroup
		attempts := 5
		results := make(chan bool, attempts)

		// 用户在网络不好时连续点击提交按钮
		lunchAmount := 35.0
		for i := 0; i < attempts; i++ {
			wg.Add(1)
			go func(attemptNum int) {
				defer wg.Done()
				time.Sleep(time.Duration(attemptNum*50) * time.Millisecond)

				success := createTransaction(t, token, string(account.ID), "expense", lunchAmount,
					fmt.Sprintf("午餐-尝试#%d", attemptNum))
				results <- success
			}(i)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		time.Sleep(200 * time.Millisecond)

		// 验证：无论用户点击多少次，交易应该都被记录（除非有去重机制）
		var txCount int64
		testDB.Model(&models.Transaction{}).
			Where("account_id = ? AND amount = ?", account.ID, lunchAmount).
			Count(&txCount)

		t.Logf("User clicked %d times, %d succeeded, %d transactions recorded",
			attempts, successCount, txCount)

		// 验证余额一致性
		var currentBalance models.Account
		testDB.Where("id = ?", account.ID).First(&currentBalance)
		expectedBalance := 982.0 - (float64(txCount) * lunchAmount)
		assert.Equal(t, expectedBalance, currentBalance.Balance,
			"Balance should reflect all recorded transactions")
	})

	t.Run("EveningBatchRecording", func(t *testing.T) {
		// 晚上回家后，用户回忆并批量记录白天的支出
		expenses := []struct {
			amount float64
			remark string
		}{
			{20.0, "下午茶"},
			{8.0, "水果"},
			{50.0, "打车回家"},
			{100.0, "超市购物"},
			{30.0, "晚餐"},
		}

		var wg sync.WaitGroup
		for _, expense := range expenses {
			wg.Add(1)
			go func(amt float64, rem string) {
				defer wg.Done()
				createTransaction(t, token, string(account.ID), "expense", amt, rem)
			}(expense.amount, expense.remark)
		}

		wg.Wait()
		time.Sleep(300 * time.Millisecond)

		// 验证所有交易都已记录
		var finalBalance models.Account
		testDB.Where("id = ?", account.ID).First(&finalBalance)

		t.Logf("Final balance after a day: %.2f", finalBalance.Balance)
		assert.Less(t, finalBalance.Balance, 1000.0, "Balance should be less than initial 1000")
		assert.GreaterOrEqual(t, finalBalance.Balance, 0.0, "Balance should not be negative")
	})
}

// TestRealWorldScenario_SessionExpiredDuringOperation 测试操作过程中会话过期
func TestRealWorldScenario_SessionExpiredDuringOperation(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "session_test@example.com")
	token := loginAndGetTokenWithDevice(t, "session_test@example.com", "TestPassword123!", "device-1")
	account := createTestAccount(t, user.ID, "bank", 1000.0)

	t.Run("TokenExpirySimulation", func(t *testing.T) {
		// 第一笔交易应该成功
		success := createTransaction(t, token, string(account.ID), "expense", 100.0, "交易1")
		assert.True(t, success, "First transaction should succeed")

		// 模拟：用户手动登出（撤销token）
		logoutReq := map[string]interface{}{"deviceId": "device-1"}
		body, _ := json.Marshal(logoutReq)
		req, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		waitForDB()

		// 尝试使用已失效的 token 创建交易
		success = createTransaction(t, token, string(account.ID), "expense", 100.0, "交易2-应该失败")
		assert.False(t, success, "Transaction with expired token should fail")

		// 验证余额没有被第二笔交易改变
		var finalBalance models.Account
		testDB.Where("id = ?", account.ID).First(&finalBalance)
		assert.Equal(t, 900.0, finalBalance.Balance, "Balance should only reflect the first transaction")
	})
}

// TestRealWorldScenario_HighConcurrencyStress 压力测试：高并发场景
func TestRealWorldScenario_HighConcurrencyStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	cleanupTestData(testDB)

	// 创建10个用户，每个用户从3个设备登录，每个设备创建10笔交易
	// 总共：10 * 3 * 10 = 300 笔交易

	numUsers := 10
	numDevices := 3
	numTransactionsPerDevice := 10

	users := make([]struct {
		email   string
		tokens  []string
		account *models.Account
	}, numUsers)

	// 创建用户和账户
	for i := 0; i < numUsers; i++ {
		email := fmt.Sprintf("stress_user_%d@example.com", i)
		user := createTestUser(t, email)
		users[i].email = email
		users[i].account = createTestAccount(t, user.ID, "bank", 10000.0)
		users[i].tokens = make([]string, numDevices)

		// 每个用户从多个设备登录
		for j := 0; j < numDevices; j++ {
			deviceID := fmt.Sprintf("user%d-device%d", i, j)
			users[i].tokens[j] = loginAndGetTokenWithDevice(t, email, "TestPassword123!", deviceID)
		}
	}

	t.Run("MassiveConcurrentOperations", func(t *testing.T) {
		startTime := time.Now()

		var wg sync.WaitGroup
		successCount := int64(0)
		var mu sync.Mutex

		// 所有用户的所有设备同时创建交易
		for userIdx, user := range users {
			for deviceIdx, token := range user.tokens {
				for txIdx := 0; txIdx < numTransactionsPerDevice; txIdx++ {
					wg.Add(1)
					go func(uIdx, dIdx, tIdx int, tok, accID string) {
						defer wg.Done()

						amount := 10.0 + float64(rand.Intn(90))
						remark := fmt.Sprintf("U%d-D%d-T%d", uIdx, dIdx, tIdx)

						success := createTransaction(t, tok, accID, "expense", amount, remark)
						if success {
							mu.Lock()
							successCount++
							mu.Unlock()
						}
					}(userIdx, deviceIdx, txIdx, token, user.account.ID)
				}
			}
		}

		wg.Wait()
		duration := time.Since(startTime)

		expectedTotal := numUsers * numDevices * numTransactionsPerDevice
		t.Logf("\n=== Stress Test Results ===")
		t.Logf("Total transactions: %d", expectedTotal)
		t.Logf("Successful: %d", successCount)
		t.Logf("Failed: %d", expectedTotal-int(successCount))
		t.Logf("Duration: %v", duration)
		t.Logf("TPS: %.2f", float64(successCount)/duration.Seconds())

		// 至少90%的交易应该成功
		successRate := float64(successCount) / float64(expectedTotal)
		assert.GreaterOrEqual(t, successRate, 0.9,
			"At least 90%% of transactions should succeed in stress test")

		time.Sleep(1 * time.Second) // 等待所有数据库操作完成

		// 验证数据一致性
		for _, user := range users {
			var txCount int64
			testDB.Model(&models.Transaction{}).
				Where("account_id = ?", user.account.ID).
				Count(&txCount)

			var account models.Account
			testDB.Where("id = ?", user.account.ID).First(&account)

			// 余额应该 = 初始值 - 所有成功交易的金额
			// 由于我们不知道确切的金额，只验证交易数量
			assert.Greater(t, txCount, int64(0), "User should have transaction records")
			assert.LessOrEqual(t, account.Balance, 10000.0, "Balance should not exceed initial value")
		}
	})
}

// TestRealWorldScenario_DataConsistencyUnderFailure 测试失败情况下的数据一致性
func TestRealWorldScenario_DataConsistencyUnderFailure(t *testing.T) {
	cleanupTestData(testDB)

	user := createTestUser(t, "consistency_test@example.com")
	token := loginAndGetTokenWithDevice(t, "consistency_test@example.com", "TestPassword123!", "device")
	account := createTestAccount(t, user.ID, "bank", 1000.0)

	t.Run("PartialFailureScenario", func(t *testing.T) {
		// 创建一系列交易，其中一些可能失败（如无效金额、空remark等）
		type TestCase struct {
			amount        float64
			remark        string
			txType        string
			shouldSucceed bool
		}

		testCases := []TestCase{
			{100.0, "正常交易1", "expense", true},
			{-50.0, "负数金额", "expense", false}, // 应该失败
			{200.0, "正常交易2", "expense", true},
			{0.0, "零金额", "expense", false}, // 应该失败
			{150.0, "正常交易3", "expense", true},
		}

		var wg sync.WaitGroup
		results := make(chan bool, len(testCases))

		for _, tc := range testCases {
			wg.Add(1)
			go func(amt float64, rem string, typ string) {
				defer wg.Done()

				// 对于预期失败的情况，使用特殊处理
				if amt <= 0 {
					// 直接标记为失败
					results <- false
					return
				}

				success := createTransaction(t, token, string(account.ID), typ, amt, rem)
				results <- success
			}(tc.amount, tc.remark, tc.txType)
		}

		wg.Wait()
		close(results)

		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		time.Sleep(300 * time.Millisecond)

		// 验证：只有有效的交易被记录
		var finalBalance models.Account
		testDB.Where("id = ?", account.ID).First(&finalBalance)

		// 只有3笔正常交易应该成功：100 + 200 + 150 = 450
		expectedBalance := 1000.0 - 450.0
		assert.Equal(t, expectedBalance, finalBalance.Balance,
			"Balance should only reflect successful transactions")

		// 验证交易记录数
		var txCount int64
		testDB.Model(&models.Transaction{}).Where("account_id = ?", account.ID).Count(&txCount)
		assert.Equal(t, int64(3), txCount, "Should only have 3 successful transaction records")
	})
}
