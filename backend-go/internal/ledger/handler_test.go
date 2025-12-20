package ledger

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"neverbroke/internal/accounts"
	"neverbroke/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockTransactionStore 模拟交易存储
type MockTransactionStore struct {
	transactions []*Transaction
}

func NewMockTransactionStore() *MockTransactionStore {
	return &MockTransactionStore{
		transactions: make([]*Transaction, 0),
	}
}

func (m *MockTransactionStore) ListByUserBetween(ctx context.Context, userID, from, to string) ([]*Transaction, error) {
	fromTime, _ := time.Parse(time.RFC3339, from)
	toTime, _ := time.Parse(time.RFC3339, to)

	var result []*Transaction
	for _, tx := range m.transactions {
		occurredAt, err := time.Parse(time.RFC3339, tx.OccurredAt)
		if err != nil {
			continue
		}
		if (occurredAt.Equal(fromTime) || occurredAt.After(fromTime)) &&
			(occurredAt.Equal(toTime) || occurredAt.Before(toTime)) {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (m *MockTransactionStore) Create(ctx context.Context, userID string, tx *Transaction) error {
	m.transactions = append(m.transactions, tx)
	return nil
}

func (m *MockTransactionStore) AddTransaction(tx *Transaction) {
	m.transactions = append(m.transactions, tx)
}

func setupLedgerTestRouter() (*gin.Engine, *MockTransactionStore) {
	router := testutil.SetupTestRouter()

	// 添加模拟认证中间件，设置user_id到context
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user-123")
		c.Next()
	})

	mockStore := NewMockTransactionStore()
	handler := NewHandler(mockStore, nil) // nil DB for unit tests using memory accounts

	router.GET("/ledger/daily", handler.GetDailyTransactions)
	router.POST("/ledger/transactions", handler.CreateTransaction)

	return router, mockStore
}

func TestGetDailyTransactions(t *testing.T) {
	router, mockStore := setupLedgerTestRouter()

	// 准备测试数据
	date := "2024-01-15"
	tx1 := &Transaction{
		ID: "tx-1",
		CreateTransactionRequest: CreateTransactionRequest{
			Type:        TransactionTypeIncome,
			Amount:      5000,
			CategoryKey: "INCOME_SALARY",
			Remark:      "工资",
			OccurredAt:  "2024-01-15T10:00:00Z",
			AccountID:   "acc-1",
		},
	}
	tx2 := &Transaction{
		ID: "tx-2",
		CreateTransactionRequest: CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      100,
			CategoryKey: "EXPENSE_FOOD",
			Remark:      "午餐",
			OccurredAt:  "2024-01-15T12:30:00Z",
			AccountID:   "acc-1",
		},
	}
	tx3 := &Transaction{
		ID: "tx-3",
		CreateTransactionRequest: CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      50,
			CategoryKey: "EXPENSE_TRANSPORT",
			Remark:      "打车",
			OccurredAt:  "2024-01-15T18:00:00Z",
			AccountID:   "acc-1",
		},
	}

	mockStore.AddTransaction(tx1)
	mockStore.AddTransaction(tx2)
	mockStore.AddTransaction(tx3)

	t.Run("successful query", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/ledger/daily?date="+date, nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})

		assert.Equal(t, date, data["date"])

		stats := data["stats"].(map[string]interface{})
		assert.Equal(t, 5000.0, stats["totalIncome"])
		assert.Equal(t, 150.0, stats["totalExpense"])

		list := data["list"].([]interface{})
		assert.Len(t, list, 3)
	})

	t.Run("missing date parameter", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/ledger/daily", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "date parameter is required")
	})

	t.Run("invalid date format", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/ledger/daily?date=invalid-date", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid date format, expected YYYY-MM-DD")
	})

	t.Run("no transactions for date", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/ledger/daily?date=2024-01-20", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		stats := data["stats"].(map[string]interface{})
		assert.Equal(t, 0.0, stats["totalIncome"])
		assert.Equal(t, 0.0, stats["totalExpense"])

		list := data["list"]
		assert.Nil(t, list)
	})
}

func TestCreateTransaction(t *testing.T) {
	router, _ := setupLedgerTestRouter()

	// 准备测试账户
	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-1",
		Type:     accounts.AccountTypeBank,
		Balance:  10000,
		Currency: "CNY",
	})
	accounts.AddAccount(accounts.Account{
		ID:       "acc-2",
		Type:     accounts.AccountTypeBank,
		Balance:  5000,
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("create expense transaction", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      100,
			CategoryKey: "EXPENSE_FOOD",
			Remark:      "午餐",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "acc-1",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})

		assert.NotEmpty(t, data["id"])
		assert.Equal(t, string(TransactionTypeExpense), data["type"])
		assert.Equal(t, 100.0, data["amount"])

		// 验证账户余额已更新
		account, _ := accounts.FindAccountByID("acc-1")
		assert.Equal(t, 9900.0, account.Balance)
	})

	t.Run("create income transaction", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeIncome,
			Amount:      5000,
			CategoryKey: "INCOME_SALARY",
			Remark:      "工资",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "acc-1",
		}

		// 先重置账户
		accounts.ResetAccounts()
		accounts.AddAccount(accounts.Account{
			ID:       "acc-1",
			Type:     accounts.AccountTypeBank,
			Balance:  10000,
			Currency: "CNY",
		})

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, string(TransactionTypeIncome), data["type"])

		// 验证账户余额已更新
		account, _ := accounts.FindAccountByID("acc-1")
		assert.Equal(t, 15000.0, account.Balance)
	})

	t.Run("create transfer transaction", func(t *testing.T) {
		// 重置账户
		accounts.ResetAccounts()
		accounts.AddAccount(accounts.Account{
			ID:       "acc-1",
			Type:     accounts.AccountTypeBank,
			Balance:  10000,
			Currency: "CNY",
		})
		accounts.AddAccount(accounts.Account{
			ID:       "acc-2",
			Type:     accounts.AccountTypeBank,
			Balance:  5000,
			Currency: "CNY",
		})

		req := CreateTransactionRequest{
			Type:          TransactionTypeTransfer,
			Amount:        1000,
			Remark:        "转账",
			OccurredAt:    time.Now().Format(time.RFC3339),
			FromAccountID: "acc-1",
			ToAccountID:   "acc-2",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		// 验证账户余额已更新
		fromAccount, _ := accounts.FindAccountByID("acc-1")
		assert.Equal(t, 9000.0, fromAccount.Balance)

		toAccount, _ := accounts.FindAccountByID("acc-2")
		assert.Equal(t, 6000.0, toAccount.Balance)
	})

	t.Run("invalid amount", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      0,
			CategoryKey: "EXPENSE_FOOD",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "acc-1",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "amount must be > 0")
	})

	t.Run("missing account for expense", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      100,
			CategoryKey: "EXPENSE_FOOD",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "accountId is required")
	})

	t.Run("invalid transfer accounts", func(t *testing.T) {
		// 测试缺少 FromAccountID
		req1 := CreateTransactionRequest{
			Type:        TransactionTypeTransfer,
			Amount:      100,
			OccurredAt:  time.Now().Format(time.RFC3339),
			ToAccountID: "acc-2",
		}
		w1 := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req1)
		testutil.AssertErrorResponse(t, w1, http.StatusBadRequest, "invalid transfer accounts")

		// 测试缺少 ToAccountID
		req2 := CreateTransactionRequest{
			Type:          TransactionTypeTransfer,
			Amount:        100,
			OccurredAt:    time.Now().Format(time.RFC3339),
			FromAccountID: "acc-1",
		}
		w2 := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req2)
		testutil.AssertErrorResponse(t, w2, http.StatusBadRequest, "invalid transfer accounts")

		// 测试相同账户转账
		req3 := CreateTransactionRequest{
			Type:          TransactionTypeTransfer,
			Amount:        100,
			OccurredAt:    time.Now().Format(time.RFC3339),
			FromAccountID: "acc-1",
			ToAccountID:   "acc-1",
		}
		w3 := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req3)
		testutil.AssertErrorResponse(t, w3, http.StatusBadRequest, "invalid transfer accounts")
	})

	t.Run("account not found", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      100,
			CategoryKey: "EXPENSE_FOOD",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "non-existent",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "account not found")
	})

	t.Run("invalid transaction type", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        "invalid",
			Amount:      100,
			CategoryKey: "EXPENSE_FOOD",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "acc-1",
		}

		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid type")
	})

	t.Run("invalid json", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "POST", "/ledger/transactions", "invalid json")
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid request")
	})
}

func TestTransactionWithInstallment(t *testing.T) {
	router, _ := setupLedgerTestRouter()

	// 准备测试账户
	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-1",
		Type:     accounts.AccountTypeCredit,
		Balance:  0,
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("create transaction with installment info", func(t *testing.T) {
		req := CreateTransactionRequest{
			Type:        TransactionTypeExpense,
			Amount:      500,
			CategoryKey: "EXPENSE_SHOPPING",
			Remark:      "分期购物 第1期",
			OccurredAt:  time.Now().Format(time.RFC3339),
			AccountID:   "acc-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/ledger/transactions", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)

		data := response["data"].(map[string]interface{})
		assert.NotEmpty(t, data["id"])
		assert.Equal(t, "分期购物 第1期", data["remark"])
	})
}
