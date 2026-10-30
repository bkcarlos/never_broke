package report

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"neverbroke/internal/ledger"
	"neverbroke/internal/repository"
)

type reportStore struct {
	repository.Store
	tx *reportTransactions
}

func (s reportStore) Transactions() repository.TransactionRepository { return s.tx }

type reportTransactions struct {
	repository.TransactionRepository
	users []string
}

func (r *reportTransactions) ListByUserBetween(_ context.Context, user, from, to string) ([]*ledger.Transaction, error) {
	r.users = append(r.users, user)
	return []*ledger.Transaction{}, nil
}
func TestMonthlyReportUsesAuthenticatedUser(t *testing.T) {
	tx := &reportTransactions{}
	h := NewHandler(reportStore{tx: tx})
	for _, user := range []string{"alice", "bob", ""} {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", "/reports/monthly?month=2026-10", nil)
		if user != "" {
			c.Set("user_id", user)
		}
		h.GetMonthlyReport(c)
		if user == "" {
			require.Equal(t, 401, c.Writer.Status())
		} else {
			require.Equal(t, 200, c.Writer.Status())
		}
	}
	require.Equal(t, []string{"alice", "bob"}, tx.users)
}
