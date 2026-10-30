package ledger

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type dailyStore struct{ transactions []*Transaction }

func (s dailyStore) ListByUserBetween(context.Context, string, string, string) ([]*Transaction, error) {
	return s.transactions, nil
}
func (s dailyStore) Create(context.Context, string, *Transaction) error { return nil }

func TestDailyTransactionsListIsAlwaysAnArray(t *testing.T) {
	for _, tc := range []struct {
		name            string
		txs             []*Transaction
		income, expense float64
	}{
		{name: "nil repository result"},
		{name: "empty repository result", txs: []*Transaction{}},
		{name: "transactions", txs: []*Transaction{{ID: "expense", CreateTransactionRequest: CreateTransactionRequest{Type: TransactionTypeExpense, Amount: 12.5, OccurredAt: "2026-10-30T12:00:00+08:00"}}}, expense: 12.5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/daily", func(c *gin.Context) { c.Set("user_id", "1") }, NewHandler(dailyStore{tc.txs}, nil).GetDailyTransactions)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest("GET", "/daily?date=2026-10-30", nil))
			require.Equal(t, 200, w.Code)
			var response struct {
				Code int `json:"code"`
				Data struct {
					List  json.RawMessage `json:"list"`
					Stats struct {
						Income  float64 `json:"totalIncome"`
						Expense float64 `json:"totalExpense"`
					} `json:"stats"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
			require.Zero(t, response.Code)
			require.NotEmpty(t, response.Data.List)
			require.Equal(t, byte('['), response.Data.List[0], "list must be a JSON array, never null")
			var list []any
			require.NoError(t, json.Unmarshal(response.Data.List, &list))
			require.Len(t, list, len(tc.txs))
			require.Equal(t, tc.income, response.Data.Stats.Income)
			require.Equal(t, tc.expense, response.Data.Stats.Expense)
		})
	}
}
