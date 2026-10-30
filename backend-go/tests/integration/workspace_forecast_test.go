package integration

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
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
		require.NoError(t, testDB.Create(&models.CuratedImportRecord{ID: "forecast-" + key, UserID: uid, SourceKey: key, Kind: "schedule", Currency: "CNY", Date: record.Date, Payload: string(raw), Digest: key}).Error)
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
