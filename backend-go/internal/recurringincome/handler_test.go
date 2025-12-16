package recurringincome

import (
	"net/http"
	"testing"

	"neverbroke/internal/accounts"
	"neverbroke/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	router := testutil.SetupTestRouter()
	handler := NewHandler()

	router.GET("/recurring-income", handler.ListPlansHandler)
	router.POST("/recurring-income", handler.UpsertPlanHandler)

	return router
}

func TestListPlansHandler(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{
		{ID: "plan1", Amount: 10000, Active: true},
		{ID: "plan2", Amount: 20000, Active: true},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	w := testutil.MakeRequest(t, router, "GET", "/recurring-income", nil)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestUpsertPlanHandler(t *testing.T) {
	router := setupTestRouter()

	// Setup test accounts
	accounts.AddAccount(accounts.Account{ID: "test-acc", Type: accounts.AccountTypeBank})

	mu.Lock()
	originalPlans := plans
	plans = []RecurringIncomePlan{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("create new plan", func(t *testing.T) {
		req := UpsertRecurringIncomeRequest{
			Type:        RecurringIncomeTypeSalary,
			Label:       "Monthly Salary",
			AccountID:   "test-acc",
			Amount:      15000,
			Currency:    "CNY",
			Frequency:   FrequencyMonthly,
			NextDueDate: "2024-01-25",
			Active:      true,
		}

		w := testutil.MakeRequest(t, router, "POST", "/recurring-income", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, 15000.0, data["amount"])
	})

	t.Run("invalid amount", func(t *testing.T) {
		req := UpsertRecurringIncomeRequest{
			AccountID:   "test-acc",
			Amount:      0,
			NextDueDate: "2024-01-25",
		}

		w := testutil.MakeRequest(t, router, "POST", "/recurring-income", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "amount must be > 0")
	})

	t.Run("missing account ID", func(t *testing.T) {
		req := UpsertRecurringIncomeRequest{
			Amount:      15000,
			NextDueDate: "2024-01-25",
		}

		w := testutil.MakeRequest(t, router, "POST", "/recurring-income", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "accountId is required")
	})

	t.Run("missing next due date", func(t *testing.T) {
		req := UpsertRecurringIncomeRequest{
			AccountID: "test-acc",
			Amount:    15000,
		}

		w := testutil.MakeRequest(t, router, "POST", "/recurring-income", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "nextDueDate is required")
	})
}
