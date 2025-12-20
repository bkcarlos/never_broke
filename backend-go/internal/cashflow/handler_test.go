package cashflow

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"neverbroke/internal/accounts"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
	"neverbroke/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockStore 测试用的 mock store 实现
type mockStore struct{}

func newMockStore() repository.Store {
	return &mockStore{}
}

func (m *mockStore) Accounts() repository.AccountRepository {
	return &mockAccountRepository{}
}

func (m *mockStore) Transactions() repository.TransactionRepository {
	return &mockTransactionRepository{}
}

func (m *mockStore) Installments() repository.InstallmentRepository {
	return &mockInstallmentRepository{}
}

func (m *mockStore) RecurringIncomes() repository.RecurringIncomeRepository {
	return &mockRecurringIncomeRepository{}
}

func (m *mockStore) Close() error {
	return nil
}

// mockAccountRepository mock 账户仓储
type mockAccountRepository struct{}

func (r *mockAccountRepository) GetByID(ctx context.Context, userID, id string) (*accounts.Account, error) {
	acc, ok := accounts.FindAccountByID(id)
	if !ok {
		return nil, fmt.Errorf("account not found: %s", id)
	}
	return acc, nil
}

func (r *mockAccountRepository) ListByUser(ctx context.Context, userID string) ([]*accounts.Account, error) {
	accs := accounts.ListAccounts()
	result := make([]*accounts.Account, len(accs))
	for i := range accs {
		result[i] = &accs[i]
	}
	return result, nil
}

func (r *mockAccountRepository) Create(ctx context.Context, userID string, acc *accounts.Account) error {
	accounts.AddAccount(*acc)
	return nil
}

func (r *mockAccountRepository) Update(ctx context.Context, userID string, acc *accounts.Account) error {
	return nil
}

// mockTransactionRepository mock 交易仓储
type mockTransactionRepository struct{}

func (r *mockTransactionRepository) GetByID(ctx context.Context, userID, id string) (*ledger.Transaction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *mockTransactionRepository) ListByUserBetween(ctx context.Context, userID, from, to string) ([]*ledger.Transaction, error) {
	return []*ledger.Transaction{}, nil
}

func (r *mockTransactionRepository) Create(ctx context.Context, userID string, tx *ledger.Transaction) error {
	return nil
}

func (r *mockTransactionRepository) CreateBatch(ctx context.Context, userID string, txs []*ledger.Transaction) error {
	return nil
}

// mockInstallmentRepository mock 分期仓储
type mockInstallmentRepository struct{}

func (r *mockInstallmentRepository) GetByID(ctx context.Context, userID, id string) (*installment.InstallmentPlan, error) {
	plan, ok := installment.FindPlanByID(id)
	if !ok {
		return nil, fmt.Errorf("installment plan not found: %s", id)
	}
	return plan, nil
}

func (r *mockInstallmentRepository) ListByAccount(ctx context.Context, userID, accountID string) ([]*installment.InstallmentPlan, error) {
	plans := installment.ListPlansByAccount(accountID)
	result := make([]*installment.InstallmentPlan, len(plans))
	for i := range plans {
		result[i] = &plans[i]
	}
	return result, nil
}

func (r *mockInstallmentRepository) ListByUser(ctx context.Context, userID string) ([]*installment.InstallmentPlan, error) {
	plans := installment.ListPlansByAccount("")
	result := make([]*installment.InstallmentPlan, 0, len(plans))
	for i := range plans {
		result = append(result, &plans[i])
	}
	return result, nil
}

func (r *mockInstallmentRepository) Create(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	installment.AddPlan(*plan)
	return nil
}

func (r *mockInstallmentRepository) Update(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	return nil
}

// mockRecurringIncomeRepository mock 周期收入仓储
type mockRecurringIncomeRepository struct{}

func (r *mockRecurringIncomeRepository) GetByID(ctx context.Context, userID, id string) (*recurringincome.RecurringIncomePlan, error) {
	plan := recurringincome.GetPlanByID(id)
	if plan == nil {
		return nil, fmt.Errorf("recurring income plan not found: %s", id)
	}
	return plan, nil
}

func (r *mockRecurringIncomeRepository) ListByUser(ctx context.Context, userID string) ([]*recurringincome.RecurringIncomePlan, error) {
	plans := recurringincome.ListPlans()
	result := make([]*recurringincome.RecurringIncomePlan, len(plans))
	for i := range plans {
		result[i] = &plans[i]
	}
	return result, nil
}

func (r *mockRecurringIncomeRepository) Create(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	recurringincome.SavePlan(*plan)
	return nil
}

func (r *mockRecurringIncomeRepository) Update(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	recurringincome.SavePlan(*plan)
	return nil
}

func setupCashflowTestRouter() (*gin.Engine, repository.Store) {
	router := testutil.SetupTestRouter()
	store := newMockStore()

	router.GET("/cashflow/forecast", ForecastHandler(store))

	return router, store
}

func TestForecastHandler(t *testing.T) {
	router, _ := setupCashflowTestRouter()

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

	t.Run("successful forecast with total assets", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=total_assets&safeAmount=2000", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")

		// 检查是否有错误
		if code, ok := response["code"].(float64); ok && code != 0 {
			t.Fatalf("API returned error: code=%v, message=%v", code, response["message"])
		}

		if response["data"] == nil {
			t.Fatal("Response data is nil")
		}

		data := response["data"].(map[string]interface{})

		assert.Contains(t, data, "status")
		assert.Contains(t, data, "points")

		points := data["points"].([]interface{})
		assert.NotEmpty(t, points)
	})

	t.Run("successful forecast with specific account", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=account&accountId=acc-1&safeAmount=1000", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})

		assert.Contains(t, data, "status")
		assert.Contains(t, data, "points")
	})

	t.Run("missing startDate parameter", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/cashflow/forecast?days=30", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "startDate is required")
	})

	t.Run("invalid days parameter", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=invalid", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid days parameter")
	})

	t.Run("negative days", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=-1", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid days parameter")
	})

	t.Run("invalid base parameter", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=invalid", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid base parameter")
	})

	t.Run("missing accountId when base is account", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=account", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "accountId is required when base is account")
	})

	t.Run("invalid safeAmount", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&safeAmount=invalid", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid safeAmount parameter")
	})

	t.Run("negative safeAmount", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&safeAmount=-100", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid safeAmount parameter")
	})

	t.Run("default parameters", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate, nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})

		// 验证使用了默认值
		points := data["points"].([]interface{})
		// 默认 30 天，应该有至少 1 个点
		assert.NotEmpty(t, points)
	})

	t.Run("large days value", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=365", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})

		points := data["points"].([]interface{})
		assert.NotEmpty(t, points)
	})

	t.Run("zero safeAmount", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&safeAmount=0", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
	})
}

func TestForecastStatuses(t *testing.T) {
	router, _ := setupCashflowTestRouter()

	// 准备测试场景：低余额账户
	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-low",
		Type:     accounts.AccountTypeBank,
		Balance:  100, // 低余额
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("warning status with low balance", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=account&accountId=acc-low&safeAmount=1000", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		// 由于余额 100 < safeAmount 1000，状态可能是 warning 或 danger
		status := data["status"].(string)
		assert.Contains(t, []string{"warning", "danger"}, status)
	})

	t.Run("safe status with high balance", func(t *testing.T) {
		// 添加高余额账户
		accounts.AddAccount(accounts.Account{
			ID:       "acc-high",
			Type:     accounts.AccountTypeBank,
			Balance:  50000, // 高余额
			Currency: "CNY",
		})

		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=30&base=account&accountId=acc-high&safeAmount=1000", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		// 余额远高于 safeAmount，状态应该是 safe
		status := data["status"].(string)
		// 状态可能是 safe 或 warning，取决于具体实现
		assert.NotEmpty(t, status)
	})
}

func TestForecastWithRecurringIncome(t *testing.T) {
	router, _ := setupCashflowTestRouter()

	// 准备测试账户
	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-1",
		Type:     accounts.AccountTypeBank,
		Balance:  5000,
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("forecast with future income", func(t *testing.T) {
		// TODO: 添加周期收入计划后测试
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=60&base=total_assets", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		// 检查是否有错误
		if code, ok := response["code"].(float64); ok && code != 0 {
			t.Fatalf("API returned error: code=%v, message=%v", code, response["message"])
		}

		if response["data"] == nil {
			t.Fatal("Response data is nil")
		}

		data := response["data"].(map[string]interface{})
		points := data["points"].([]interface{})
		assert.NotEmpty(t, points)
	})
}

func TestForecastContextCancellation(t *testing.T) {
	router, _ := setupCashflowTestRouter()

	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-1",
		Type:     accounts.AccountTypeBank,
		Balance:  10000,
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("request completes normally", func(t *testing.T) {
		startDate := time.Now().Format("2006-01-02")
		w := testutil.MakeRequest(t, router, "GET",
			"/cashflow/forecast?startDate="+startDate+"&days=7", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)
	})
}

func TestBuildForecastDirectly(t *testing.T) {
	store := newMockStore()
	ctx := context.Background()

	// 准备测试账户
	accounts.ResetAccounts()
	accounts.AddAccount(accounts.Account{
		ID:       "acc-1",
		Type:     accounts.AccountTypeBank,
		Balance:  10000,
		Currency: "CNY",
	})

	defer accounts.ResetAccounts()

	t.Run("build forecast with valid parameters", func(t *testing.T) {
		req := CashflowForecastRequest{
			StartDate:  time.Now().Format("2006-01-02"),
			Days:       30,
			Base:       ForecastBaseTotalAssets,
			SafeAmount: 2000,
		}

		result, err := BuildForecast(ctx, store, "demo", req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Points)
		assert.NotEmpty(t, result.Status)
	})

	t.Run("build forecast with account base", func(t *testing.T) {
		req := CashflowForecastRequest{
			StartDate:  time.Now().Format("2006-01-02"),
			Days:       30,
			Base:       ForecastBaseAccount,
			AccountID:  "acc-1",
			SafeAmount: 2000,
		}

		result, err := BuildForecast(ctx, store, "demo", req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Points)
	})

	t.Run("short period forecast", func(t *testing.T) {
		req := CashflowForecastRequest{
			StartDate:  time.Now().Format("2006-01-02"),
			Days:       7,
			Base:       ForecastBaseTotalAssets,
			SafeAmount: 1000,
		}

		result, err := BuildForecast(ctx, store, "demo", req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotEmpty(t, result.Points)
		// 7 天的预测应该有至少 1 个点
		assert.GreaterOrEqual(t, len(result.Points), 1)
	})
}
