package accounts

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

	router.GET("/accounts", handler.ListAccountsHandler)
	router.POST("/accounts", handler.CreateAccountHandler)
	router.PUT("/accounts/:id/credit-limits", handler.UpdateCreditLimitsHandler)

	return router
}

func TestListAccountsHandler(t *testing.T) {
	router := setupTestRouter()

	// Save and restore original accounts
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "acc1", Type: AccountTypeBank, Balance: 1000, Currency: "CNY"},
		{ID: "acc2", Type: AccountTypeCash, Balance: 500, Currency: "CNY"},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	w := testutil.MakeRequest(t, router, "GET", "/accounts", nil)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	assert.Contains(t, response, "data")
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestCreateAccountHandler_Cash(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{} // Start with empty accounts
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:     AccountTypeCash,
		Currency: "CNY",
		Balance:  500,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	assert.Contains(t, response, "data")
	data := response["data"].(map[string]interface{})
	assert.Equal(t, string(AccountTypeCash), data["type"])
	assert.Equal(t, 500.0, data["balance"])
}

func TestCreateAccountHandler_DuplicateCash(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "cash1", Type: AccountTypeCash, Archived: false},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:     AccountTypeCash,
		Currency: "CNY",
		Balance:  500,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "cash account already exists")
}

func TestCreateAccountHandler_Bank(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:        AccountTypeBank,
		Institution: AccountInstitutionICBC,
		CardLast4:   "1234",
		Currency:    "CNY",
		Balance:     5000,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, string(AccountTypeBank), data["type"])
	assert.Equal(t, string(AccountInstitutionICBC), data["institution"])
	assert.Equal(t, "1234", data["cardLast4"])
}

func TestCreateAccountHandler_BankMissingInstitution(t *testing.T) {
	router := setupTestRouter()

	req := CreateAccountRequest{
		Type:     AccountTypeBank,
		Currency: "CNY",
		Balance:  5000,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "institution is required for bank/credit accounts")
}

func TestCreateAccountHandler_BankInvalidCardLast4(t *testing.T) {
	router := setupTestRouter()

	tests := []struct {
		name      string
		cardLast4 string
	}{
		{"too short", "1"},
		{"too long", "12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := CreateAccountRequest{
				Type:        AccountTypeBank,
				Institution: AccountInstitutionICBC,
				CardLast4:   tt.cardLast4,
				Currency:    "CNY",
				Balance:     5000,
			}

			w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
			testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "cardLast4 must be 2-4 digits")
		})
	}
}

func TestCreateAccountHandler_Credit(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:        AccountTypeCredit,
		Institution: AccountInstitutionCMB,
		CardLast4:   "5678",
		Currency:    "CNY",
		Balance:     2000,
		CreditLimit: 20000,
		TempLimit:   3000,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, string(AccountTypeCredit), data["type"])
	assert.Equal(t, 20000.0, data["creditLimit"])
	assert.Equal(t, 3000.0, data["tempLimit"])
}

func TestCreateAccountHandler_Wallet(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:        AccountTypeWallet,
		Institution: AccountInstitutionWeChat,
		Currency:    "CNY",
		Balance:     1200,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, string(AccountTypeWallet), data["type"])
	assert.Equal(t, string(AccountInstitutionWeChat), data["institution"])
}

func TestCreateAccountHandler_WalletMissingInstitution(t *testing.T) {
	router := setupTestRouter()

	req := CreateAccountRequest{
		Type:     AccountTypeWallet,
		Currency: "CNY",
		Balance:  1200,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "institution is required for wallet accounts")
}

func TestCreateAccountHandler_InvalidType(t *testing.T) {
	router := setupTestRouter()

	req := CreateAccountRequest{
		Type:     "invalid",
		Currency: "CNY",
		Balance:  1000,
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid account type")
}

func TestCreateAccountHandler_DefaultCurrency(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	req := CreateAccountRequest{
		Type:    AccountTypeCash,
		Balance: 500,
		// Currency omitted
	}

	w := testutil.MakeRequest(t, router, "POST", "/accounts", req)
	testutil.AssertStatusCode(t, w, http.StatusOK)

	var response map[string]interface{}
	testutil.ParseResponse(t, w, &response)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "CNY", data["currency"])
}

func TestUpdateCreditLimitsHandler(t *testing.T) {
	router := setupTestRouter()

	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "credit1", Type: AccountTypeCredit, CreditLimit: 10000, TempLimit: 0},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	t.Run("update both limits", func(t *testing.T) {
		newCreditLimit := 15000.0
		newTempLimit := 2000.0
		req := UpdateCreditLimitRequest{
			CreditLimit: &newCreditLimit,
			TempLimit:   &newTempLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/credit1/credit-limits", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, 15000.0, data["creditLimit"])
		assert.Equal(t, 2000.0, data["tempLimit"])
	})

	t.Run("update credit limit only", func(t *testing.T) {
		newCreditLimit := 18000.0
		req := UpdateCreditLimitRequest{
			CreditLimit: &newCreditLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/credit1/credit-limits", req)
		testutil.AssertStatusCode(t, w, http.StatusOK)

		var response map[string]interface{}
		testutil.ParseResponse(t, w, &response)

		data := response["data"].(map[string]interface{})
		assert.Equal(t, 18000.0, data["creditLimit"])
	})

	t.Run("account not found", func(t *testing.T) {
		newCreditLimit := 15000.0
		req := UpdateCreditLimitRequest{
			CreditLimit: &newCreditLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/non-existent/credit-limits", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "account not found")
	})

	t.Run("not a credit account", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "bank1", Type: AccountTypeBank},
		}
		mu.Unlock()

		newCreditLimit := 15000.0
		req := UpdateCreditLimitRequest{
			CreditLimit: &newCreditLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/bank1/credit-limits", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "not a credit account")
	})

	t.Run("negative credit limit", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "credit1", Type: AccountTypeCredit, CreditLimit: 10000},
		}
		mu.Unlock()

		negativeCreditLimit := -1000.0
		req := UpdateCreditLimitRequest{
			CreditLimit: &negativeCreditLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/credit1/credit-limits", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "creditLimit must be >= 0")
	})

	t.Run("negative temp limit", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "credit1", Type: AccountTypeCredit, CreditLimit: 10000},
		}
		mu.Unlock()

		negativeTempLimit := -500.0
		req := UpdateCreditLimitRequest{
			TempLimit: &negativeTempLimit,
		}

		w := testutil.MakeRequest(t, router, "PUT", "/accounts/credit1/credit-limits", req)
		testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "tempLimit must be >= 0")
	})
}

func TestCreateAccountHandler_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	w := testutil.MakeRequest(t, router, "POST", "/accounts", "invalid json")
	testutil.AssertErrorResponse(t, w, http.StatusBadRequest, "invalid request")
}
