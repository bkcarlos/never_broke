package installment

import (
	"net/http"
	"testing"

	"neverbroke/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	router := testutil.SetupTestRouter()
	handler := NewHandler()

	router.POST("/installment/preview", handler.PreviewInstallmentHandler)
	router.POST("/installment/expense", handler.CreateExpenseInstallmentHandler)

	return router
}

func TestPreviewInstallmentHandler(t *testing.T) {
	router := setupTestRouter()

	t.Run("valid equal principal request", func(t *testing.T) {
		req := PreviewInstallmentRequest{
			Principal:          10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPrincipal,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/preview", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, 10000.0, data["principal"])
		assert.Equal(t, float64(12), data["termMonths"])
		assert.Contains(t, data, "schedule")
	})

	t.Run("valid equal payment request", func(t *testing.T) {
		req := PreviewInstallmentRequest{
			Principal:          10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/preview", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)
	})

	t.Run("invalid request body", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "POST", "/installment/preview", "invalid json")
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid request")
	})

	t.Run("invalid repay method", func(t *testing.T) {
		req := PreviewInstallmentRequest{
			Principal:          10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        "invalid",
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/preview", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "build schedule failed")
	})

	t.Run("negative principal", func(t *testing.T) {
		req := PreviewInstallmentRequest{
			Principal:          -10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPrincipal,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/preview", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "build schedule failed")
	})
}

func TestCreateExpenseInstallmentHandler(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalPlans := plans
	plans = []InstallmentPlan{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		plans = originalPlans
		mu.Unlock()
	}()

	t.Run("valid request", func(t *testing.T) {
		req := CreateExpenseInstallmentRequest{
			AccountID:          "acc1",
			Principal:          10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
			CategoryKey:        "shopping",
			Remark:             "Test installment",
			PurchaseDate:       "2024-01-01",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Contains(t, data, "plan")
		assert.Contains(t, data, "transactions")

		plan := data["plan"].(map[string]interface{})
		assert.Equal(t, "acc1", plan["accountId"])
		assert.Equal(t, 10000.0, plan["principal"])

		// Verify plan was stored
		list := ListPlansByAccount("acc1")
		assert.Len(t, list, 1)
	})

	t.Run("missing account ID", func(t *testing.T) {
		req := CreateExpenseInstallmentRequest{
			AccountID:          "",
			Principal:          10000,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid params")
	})

	t.Run("zero principal", func(t *testing.T) {
		req := CreateExpenseInstallmentRequest{
			AccountID:          "acc1",
			Principal:          0,
			TermMonths:         12,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid params")
	})

	t.Run("invalid term months", func(t *testing.T) {
		req := CreateExpenseInstallmentRequest{
			AccountID:          "acc1",
			Principal:          10000,
			TermMonths:         1, // Must be > 1
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			Currency:           "CNY",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid params")
	})

	t.Run("default currency", func(t *testing.T) {
		req := CreateExpenseInstallmentRequest{
			AccountID:          "acc2",
			Principal:          5000,
			TermMonths:         6,
			MonthlyRatePercent: 0.5,
			RepayMethod:        RepayMethodEqualPayment,
			FirstDueDate:       "2024-01-01",
			// Currency omitted
			PurchaseDate: "2024-01-01",
		}

		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		plan := data["plan"].(map[string]interface{})
		assert.Equal(t, "CNY", plan["currency"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "POST", "/installment/expense", "invalid json")
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid request")
	})
}
