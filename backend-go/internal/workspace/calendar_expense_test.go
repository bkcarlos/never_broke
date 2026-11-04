package workspace

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCalendarFixedExpensesExpandCadenceAndKeepAmountsSeparate(t *testing.T) {
	plans := []Setting{
		{Key: "rent", Kind: "fixed_expense", Name: "房租", Currency: "CNY", AccountID: "bank", Cents: 900000, Period: "quarterly", Date: "2026-11-05", EndDate: "2027-02-05", Enabled: true},
		{Key: "subscription", Kind: "subscription", Name: "订阅", Currency: "CNY", Cents: 3000, Period: "monthly", Date: "2026-01-31", Enabled: true},
		{Key: "insurance", Kind: "fixed_expense", Name: "保险", Currency: "CNY", Cents: 500000, Period: "yearly", Date: "2026-02-28", Enabled: true},
		{Key: "weekly", Kind: "fixed_expense", Name: "清洁", Currency: "CNY", Cents: 10000, Period: "weekly", Date: "2027-02-01", EndDate: "2027-02-15", Enabled: true},
	}
	q := CalendarQuery{Month: "2027-02", Day: "2027-02-05", View: "spending", Filter: Filter{Currency: "CNY"}}
	rows, err := buildCalendarFixedExpenses(plans, q)
	require.NoError(t, err)
	require.Len(t, rows, 6)
	require.Equal(t, "2027-02-01", rows[0].Date)
	require.Equal(t, "fixed:rent:2027-02-05", rows[1].Key)
	require.Equal(t, int64(900000), rows[1].Cents)
	require.Equal(t, "2027-02-28", rows[4].Date)
	require.Equal(t, "2027-02-28", rows[5].Date)
	q.View = "all"
	same, err := buildCalendarFixedExpenses(plans, q)
	require.NoError(t, err)
	require.Equal(t, rows, same, "repayment toggle must not hide fixed expenses")
	q.Account = "bank"
	filtered, err := buildCalendarFixedExpenses(plans, q)
	require.NoError(t, err)
	require.Len(t, filtered, 1)
	require.Equal(t, "rent", filtered[0].SettingKey)
	q.Account = ""
	q.Month = "2027-05"
	q.Day = "2027-05-05"
	rows, err = buildCalendarFixedExpenses(plans, q)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	require.Equal(t, "subscription", rows[0].SettingKey)
	// Rendering plans never adds them to ledger or repayment totals.
	_, totals, actual, err := BuildCalendar(nil, q)
	require.NoError(t, err)
	require.Equal(t, CalendarMetric{}, totals)
	require.Empty(t, actual)
}

func TestCalendarFixedExpensesFiltersEmptyAndInvalidSettings(t *testing.T) {
	q := CalendarQuery{Month: "2026-11", Day: "2026-11-05", View: "all", Filter: Filter{Currency: "CNY"}}
	plans := []Setting{
		{Kind: "fixed_expense", Currency: "CNY", Enabled: false},
		{Kind: "subscription", Currency: "USD", Enabled: true},
		{Kind: "budget", Currency: "CNY", Enabled: true},
	}
	rows, err := buildCalendarFixedExpenses(plans, q)
	require.NoError(t, err)
	require.Empty(t, rows)
	require.NotNil(t, rows)
	plans = append(plans, Setting{Kind: "fixed_expense", Currency: "CNY", Enabled: true, Name: "损坏的计划"})
	_, err = buildCalendarFixedExpenses(plans, q)
	require.ErrorContains(t, err, "损坏的计划")
	q.Month = "2026-13"
	_, err = buildCalendarFixedExpenses(nil, q)
	require.Error(t, err)
}
