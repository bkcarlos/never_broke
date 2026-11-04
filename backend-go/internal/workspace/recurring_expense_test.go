package workspace

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func recurringDate(s string) time.Time { v, _ := time.Parse(time.DateOnly, s); return v }
func TestRecurringExpenseMonthEndAndInclusiveEndDate(t *testing.T) {
	rows := RecurrenceDates(recurringDate("2028-01-31"), "monthly", "2028-03-31", recurringDate("2028-01-01"), recurringDate("2028-05-01"))
	require.Equal(t, []string{"2028-01-31", "2028-02-29", "2028-03-31"}, rows)
	require.Equal(t, []string{"2029-02-28", "2030-02-28"}, RecurrenceDates(recurringDate("2028-02-29"), "yearly", "", recurringDate("2029-01-01"), recurringDate("2031-01-01")))
	require.Equal(t, []string{"2026-11-06", "2026-11-13"}, RecurrenceDates(recurringDate("2026-10-30"), "weekly", "", recurringDate("2026-11-01"), recurringDate("2026-11-20")))
}
func TestValidateFixedExpense(t *testing.T) {
	good := Setting{Name: "Rent", Cents: 300000, Date: "2026-11-10", Period: "monthly"}
	require.NoError(t, ValidateRecurringExpense(good))
	for _, change := range []func(*Setting){func(s *Setting) { s.Name = " " }, func(s *Setting) { s.Cents = 0 }, func(s *Setting) { s.Period = "invalid" }, func(s *Setting) { s.Date = "2026-02-30" }, func(s *Setting) { s.EndDate = "2026-11-09" }} {
		bad := good
		change(&bad)
		require.Error(t, ValidateRecurringExpense(bad))
	}
}
