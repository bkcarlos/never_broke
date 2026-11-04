package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"neverbroke/internal/cashflow"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/models"
	"neverbroke/internal/workspace"
	"strconv"
	"testing"
)

func TestForecastUsesReviewedPlansAndVerifiedBalances(t *testing.T) {
	user := createTestUser(t, "forecast-workspace@example.com")
	uid := strconv.Itoa(user.ID)
	account := createTestAccount(t, user.ID, "bank", 100)
	for _, key := range []string{"future", "paid"} {
		amount := int64(1000)
		record := curatedimport.Record{Key: key, Kind: "schedule", Currency: "CNY", Date: "2026-10-30", Description: key, AmountCents: &amount}
		raw, err := json.Marshal(record)
		require.NoError(t, err)
		require.NoError(t, testDB.Create(&models.CuratedImportRecord{ID: "forecast-" + uid + "-" + key, UserID: uid, SourceKey: key, Kind: "schedule", Currency: "CNY", Date: record.Date, Payload: string(raw), Digest: key}).Error)
	}
	require.NoError(t, testDB.Create(&workspace.State{UserID: uid, Key: "paid", Version: 1, Payload: `{"status":"paid"}`}).Error)
	req := cashflow.CashflowForecastRequest{StartDate: "2026-10-30", Days: 2, Base: cashflow.ForecastBaseTotalAssets, SafeAmount: 0}
	result, err := cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, uid, req)
	require.NoError(t, err)
	require.False(t, result.BalanceKnown)
	require.Nil(t, result.Points[0].ClosingBalance)
	require.Equal(t, cashflow.ForecastStatus("unknown"), result.Status)
	require.Equal(t, 10.0, result.Points[0].Outflow)
	require.NoError(t, testDB.Create(&workspace.Setting{UserID: uid, Key: "balance:" + account.ID, Kind: "balance", Match: account.ID, Currency: "CNY", Cents: 10000, Enabled: true, Version: 1}).Error)
	result, err = cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, uid, req)
	require.NoError(t, err)
	require.True(t, result.BalanceKnown)
	require.Equal(t, 90.0, *result.Points[0].ClosingBalance)
}

func TestFixedExpensesPersistAndDriveMonthlyForecast(t *testing.T) {
	user := createTestUser(t, "fixed-expense@example.com")
	other := createTestUser(t, "fixed-expense-other@example.com")
	uid, otherID := strconv.Itoa(user.ID), strconv.Itoa(other.ID)
	bank := createTestAccount(t, user.ID, "bank", 1000)
	foreign := createTestAccount(t, other.ID, "bank", 1000)
	engine := gin.New()
	group := engine.Group("/api", func(c *gin.Context) { c.Set("user_id", c.GetHeader("X-Test-User")) })
	workspace.Register(group, testDB)
	save := func(who string, s workspace.Setting) (int, workspace.Setting) {
		t.Helper()
		raw, err := json.Marshal(s)
		require.NoError(t, err)
		req := httptest.NewRequest("POST", "/api/workspace/settings", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-User", who)
		writer := httptest.NewRecorder()
		engine.ServeHTTP(writer, req)
		var result struct {
			Code int
			Data workspace.Setting
		}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &result))
		return result.Code, result.Data
	}
	code, _ := save(uid, workspace.Setting{Key: "balance:" + bank.ID, Kind: "balance", Name: "银行余额", Currency: "CNY", Match: bank.ID, Date: "2026-11-01", Cents: 100000, Enabled: true})
	require.Zero(t, code)
	setting := workspace.Setting{Kind: "fixed_expense", Name: "房租", Currency: "CNY", AccountID: foreign.ID, Cents: 300000, Period: "monthly", Date: "2026-11-10", EndDate: "2026-12-10", Enabled: true}
	code, _ = save(uid, setting)
	require.NotZero(t, code, "must reject another user's account")
	setting.AccountID = bank.ID
	code, saved := save(uid, setting)
	require.Zero(t, code)
	require.Equal(t, 1, saved.Version)
	require.NotEmpty(t, saved.Key)
	code, _ = save(uid, workspace.Setting{Key: saved.Key, Kind: "fixed_expense", Name: "房租", Currency: "CNY", AccountID: bank.ID, Cents: 1, Period: "monthly", Date: "2026-11-10", Enabled: true})
	require.NotZero(t, code, "stale version must fail")
	req := cashflow.CashflowForecastRequest{StartDate: "2026-11-01", Days: 92, Base: cashflow.ForecastBaseTotalAssets, SafeAmount: 0}
	result, err := cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, uid, req)
	require.NoError(t, err)
	require.Len(t, result.Months, 3)
	require.Equal(t, 3000.0, result.Months[0].FixedExpense)
	require.Equal(t, 3000.0, result.Months[1].FixedExpense)
	require.Zero(t, result.Months[2].FixedExpense)
	require.Equal(t, "2026-11-10", result.Months[0].FirstGapDate)
	otherResult, err := cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, otherID, req)
	require.NoError(t, err)
	require.Zero(t, otherResult.Months[0].FixedExpense)
	saved.Enabled = false
	code, saved = save(uid, saved)
	require.Zero(t, code)
	require.Equal(t, 2, saved.Version)
	result, err = cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, uid, req)
	require.NoError(t, err)
	require.Zero(t, result.Months[0].FixedExpense)
	saved.Enabled = true
	saved.Cents = 500000
	code, _ = save(uid, saved)
	require.Zero(t, code)
	result, err = cashflow.BuildForecastWithDB(context.Background(), testStore, testDB, uid, req)
	require.NoError(t, err)
	require.Equal(t, 5000.0, result.Months[0].FixedExpense)
	var count int64
	require.NoError(t, testDB.Model(&models.Transaction{}).Where("user_id = ?", uid).Count(&count).Error)
	require.Zero(t, count, "forecast and settings must not post ledger transactions")
}
