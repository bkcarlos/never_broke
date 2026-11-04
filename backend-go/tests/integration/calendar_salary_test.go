package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/workspace"
)

func TestCalendarSalarySavedConfigurationAndIsolation(t *testing.T) {
	user := createTestUser(t, "calendar-salary@example.com")
	token := loginAndGetToken(t, user.Email, "TestPassword123!")
	account := createTestAccount(t, user.ID, "bank", 100)
	other := createTestUser(t, "calendar-salary-other@example.com")
	otherToken := loginAndGetToken(t, other.Email, "TestPassword123!")
	call := func(method, path, accessToken string, payload any, target any) {
		t.Helper()
		raw, err := json.Marshal(payload)
		require.NoError(t, err)
		req := httptest.NewRequest(method, "/api"+path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		var response struct {
			Code int
			Data json.RawMessage
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		require.Zero(t, response.Code, w.Body.String())
		require.NoError(t, json.Unmarshal(response.Data, target))
	}
	config := recurringincome.UpsertSalaryPlanRequest{PreTaxMonthly: 20000, CityCode: "beijing", StartDate: "2026-01-01", PayDay: 31, SocialRate: .222, AccountID: account.ID, Currency: "CNY", Contributions: &recurringincome.SalaryContributions{SocialBase: 20000, HousingBase: 20000, PensionRate: .08, MedicalRate: .02, UnemploymentRate: .002, HousingRate: .12}}
	var saved recurringincome.UpsertSalaryPlanResponse
	call("POST", "/me/salary-plan", token, config, &saved)
	require.Equal(t, config.Contributions, saved.Plan.SalaryContributions)
	var fetched recurringincome.GetSalaryPlanResponse
	call("GET", "/me/salary-plan", token, nil, &fetched)
	require.NotNil(t, fetched.Plan)
	require.Equal(t, config.Contributions, fetched.Plan.SalaryContributions)
	var calendar struct {
		Salaries []workspace.CalendarSalary
		Totals   workspace.CalendarMetric
	}
	path := "/workspace/calendar?month=2026-02&day=2026-02-28&currency=CNY&view=spending"
	call("GET", path, token, nil, &calendar)
	require.Len(t, calendar.Salaries, 1)
	require.Equal(t, "2026-02-28", calendar.Salaries[0].Date)
	require.Equal(t, 4440.0, calendar.Salaries[0].SocialSecurity)
	require.Equal(t, 633.6, calendar.Salaries[0].CumulativeTax)
	require.Equal(t, 15243.2, calendar.Salaries[0].PostTax)
	require.Zero(t, calendar.Totals.Income)
	call("GET", path+"&account=another", token, nil, &calendar)
	require.Empty(t, calendar.Salaries)
	call("GET", path, otherToken, nil, &calendar)
	require.Empty(t, calendar.Salaries)
	// Replacing a plan preserves itemized settings and exposes only the new plan.
	config.PreTaxMonthly = 21000
	call("POST", "/me/salary-plan", token, config, &saved)
	var active int64
	require.NoError(t, testDB.Model(&models.RecurringIncomePlan{}).Where("user_id = ? AND active = ?", strconv.Itoa(user.ID), true).Count(&active).Error)
	require.Equal(t, int64(1), active)
	call("GET", path, token, nil, &calendar)
	require.Len(t, calendar.Salaries, 1)
	require.Equal(t, 21000.0, calendar.Salaries[0].PreTax)
	// Adding a July raise retains January-June and cumulative tax across the boundary.
	config.StartDate = "2026-07-01"
	config.PreTaxMonthly = 50000
	config.BaseMode = "policy"
	call("POST", "/me/salary-plan", token, config, &saved)
	require.Len(t, saved.Plan.SalaryVersions, 2)
	call("GET", path, token, nil, &calendar)
	require.Equal(t, 21000.0, calendar.Salaries[0].PreTax)
	call("GET", "/workspace/calendar?month=2026-07&day=2026-07-01&currency=CNY&view=spending", token, nil, &calendar)
	require.Len(t, calendar.Salaries, 1)
	require.Equal(t, 50000.0, calendar.Salaries[0].PreTax)
	require.Equal(t, 176000.0, calendar.Salaries[0].CumulativePreTax)
	require.Equal(t, 36348.0, calendar.Salaries[0].Configuration.SocialBase)
	// Previewing a future change must not persist it.
	config.StartDate = "2026-10-01"
	call("POST", "/me/salary-plan?preview=true", token, config, &saved)
	call("GET", "/me/salary-plan", token, nil, &fetched)
	require.Len(t, fetched.Plan.SalaryVersions, 2)
	var refreshed models.Account
	require.NoError(t, testDB.First(&refreshed, "id = ?", account.ID).Error)
	require.Equal(t, 100.0, refreshed.Balance)
}

func TestPayrollHistoryPersistence(t *testing.T) {
	require.NoError(t, payroll.SeedPolicyHistory(testDB))
	var before, after int64
	require.NoError(t, testDB.Model(&payroll.PolicyHistory{}).Count(&before).Error)
	require.NoError(t, payroll.SeedPolicyHistory(testDB))
	require.NoError(t, testDB.Model(&payroll.PolicyHistory{}).Count(&after).Error)
	require.Equal(t, before, after)
	service := payroll.NewService(payroll.NewRepository(testDB))
	for _, tc := range []struct{ date, ss, hf string }{
		{"2025-06-30", "6821", "2420"}, {"2025-07-01", "7162", "2420"}, {"2025-09-01", "7162", "2540"}, {"2026-07-01", "7270", "2540"},
	} {
		rule, err := service.GetRegionRuleAt("CN-BJ", tc.date)
		require.NoError(t, err)
		require.Equal(t, tc.ss, rule.SSBaseMin.String())
		require.Equal(t, tc.hf, rule.HFBaseMin.String())
	}
	tx := testDB.Begin()
	defer tx.Rollback()
	require.NoError(t, tx.Delete(&payroll.PolicyHistory{}, "region_code = ?", "CN-BJ").Error)
	_, err := payroll.NewService(payroll.NewRepository(tx)).GetRegionRuleAt("CN-BJ", "2026-07-01")
	require.ErrorIs(t, err, payroll.ErrRuleNotFound) // Production reads DB, never silently falls back to seeds.
}
