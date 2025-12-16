package budget

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

	router.GET("/budget", handler.GetMonthlyBudgetHandler)
	router.PUT("/budget", handler.PutMonthlyBudgetHandler)

	return router
}

func TestGetMonthlyBudgetHandler(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalRecords := records
	records = map[string]MonthlyBudget{
		"2024-01": {Month: "2024-01", TotalBudget: 5000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		records = originalRecords
		mu.Unlock()
	}()

	t.Run("existing budget", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/budget?month=2024-01", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "2024-01", data["month"])
		assert.Equal(t, 5000.0, data["totalBudget"])
	})

	t.Run("non-existing budget", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/budget?month=2024-02", nil)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "2024-02", data["month"])
		assert.Equal(t, 0.0, data["totalBudget"]) // Default to 0
	})

	t.Run("missing month parameter", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "GET", "/budget", nil)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "month parameter is required")
	})
}

func TestPutMonthlyBudgetHandler(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalRecords := records
	records = map[string]MonthlyBudget{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		records = originalRecords
		mu.Unlock()
	}()

	t.Run("create new budget", func(t *testing.T) {
		req := MonthlyBudget{
			Month:       "2024-01",
			TotalBudget: 6000,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/budget", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		assert.Contains(t, response, "data")
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "2024-01", data["month"])
		assert.Equal(t, 6000.0, data["totalBudget"])

		// Verify it was stored
		stored, exists := GetMonthlyBudget("2024-01")
		assert.True(t, exists)
		assert.Equal(t, 6000.0, stored.TotalBudget)
	})

	t.Run("update existing budget", func(t *testing.T) {
		// First create
		UpsertMonthlyBudget(MonthlyBudget{Month: "2024-02", TotalBudget: 5000})

		// Then update
		req := MonthlyBudget{
			Month:       "2024-02",
			TotalBudget: 7000,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/budget", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, 7000.0, data["totalBudget"])

		// Verify update
		stored, _ := GetMonthlyBudget("2024-02")
		assert.Equal(t, 7000.0, stored.TotalBudget)
	})

	t.Run("empty month", func(t *testing.T) {
		req := MonthlyBudget{
			Month:       "",
			TotalBudget: 5000,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/budget", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "month is required")
	})

	t.Run("negative budget", func(t *testing.T) {
		req := MonthlyBudget{
			Month:       "2024-03",
			TotalBudget: -1000,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/budget", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "totalBudget must be >= 0")
	})

	t.Run("zero budget", func(t *testing.T) {
		req := MonthlyBudget{
			Month:       "2024-04",
			TotalBudget: 0,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/budget", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, 0.0, data["totalBudget"])
	})

	t.Run("invalid JSON", func(t *testing.T) {
		w := testutil.MakeRequest(t, router, "PUT", "/budget", "invalid json")
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid request")
	})
}
