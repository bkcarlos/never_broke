package payroll

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestOfficialBaseSelectionAndNoSilentFallback(t *testing.T) {
	service := NewService(nil)
	for _, tc := range []struct{ code, date, ssMin, ssMax, hfMin, hfMax string }{
		{"CN-BJ", "2026-06-30", "7162", "35811", "2540", "35811"},
		{"CN-BJ", "2026-07-01", "7270", "36348", "2540", "36348"},
		{"CN-BJ", "2027-06-30", "7270", "36348", "2540", "36348"},
		{"CN-SH", "2026-11-04", "7546", "37731", "2740", "37731"},
	} {
		rule, err := service.GetRegionRuleAt(tc.code, tc.date)
		require.NoError(t, err)
		require.Equal(t, tc.ssMin, rule.SSBaseMin.String())
		require.Equal(t, tc.ssMax, rule.SSBaseMax.String())
		require.Equal(t, tc.hfMin, rule.HFBaseMin.String())
		require.Equal(t, tc.hfMax, rule.HFBaseMax.String())
		require.Len(t, rule.Sources, 2)
		require.Equal(t, tc.date, rule.AsOf)
		require.NotEmpty(t, rule.VerifiedAt)
	}
	for _, code := range []string{"CN-DEFAULT", "CN-GD-SZ", "CN-GD-GZ", "unknown"} {
		_, err := service.GetRegionRuleAt(code, "2026-11-04")
		require.ErrorIs(t, err, ErrRuleNotFound)
	}
	for _, date := range []string{"2027-07-01", "2025-08-31"} {
		_, err := service.GetRegionRuleAt("CN-BJ", date)
		require.ErrorIs(t, err, ErrRuleNotFound)
	}
	_, err := service.GetRegionRuleAt("CN-BJ", "2026-02-30")
	require.Error(t, err)
}

func TestOfficialRuleEndpointsWorkWithoutPayrollTables(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHandler(nil)
	router.GET("/regions", handler.ListRegions)
	router.GET("/regions/:code/rule", handler.GetRegionRule)
	for _, path := range []string{"/regions", "/regions/CN-BJ/rule?date=2026-11-04", "/regions/CN-BJ/rule?year=2026"} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		require.Equal(t, 200, w.Code, w.Body.String())
		var result map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))
		require.Equal(t, "0", string(result["code"]))
	}
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/regions/CN-DEFAULT/rule?date=2026-11-04", 404},
		{"/regions/CN-BJ/rule?date=2027-07-01", 404},
		{"/regions/CN-BJ/rule?date=invalid", 400},
	} {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest("GET", tc.path, nil))
		require.Equal(t, tc.status, w.Code)
		var body map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
		require.NotEmpty(t, body["message"])
	}
}

func TestSalaryCalculatorUsesOfficialCapsWithoutSeedDatabase(t *testing.T) {
	result, err := NewService(nil).CalculateSalary(&CalcSalaryRequest{RegionCode: "CN-BJ", Year: 2026, RuleDate: "2026-11-04", GrossMonthly: decimal.NewFromInt(50000), SSEnabled: true, HFEnabled: true, HFRatePerson: decimal.NewFromFloat(.12)})
	require.NoError(t, err)
	require.Equal(t, "3816.54", result.Summary.SSPersonMonthly.String())
	require.Equal(t, "4361.76", result.Summary.HFPersonMonthly.String())
	require.Len(t, result.Months, 12)
	for _, month := range result.Months {
		require.True(t, month.Gross.Equal(month.SSPerson.Add(month.HFPerson).Add(month.Tax).Add(month.Net)))
	}
}
