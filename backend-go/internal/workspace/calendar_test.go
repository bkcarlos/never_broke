package workspace

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCalendarSeparatesSpendingRefundsAndActiveRepayments(t *testing.T) {
	rows := []Entry{
		{Key: "expense", Kind: "expense", Date: "2024-02-29", Currency: "CNY", AccountID: "a", Cents: cents(2600)},
		{Key: "refund", Kind: "refund", Date: "2024-02-29", Currency: "CNY", AccountID: "a", Cents: cents(600)},
		{Key: "billed", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", AccountID: "a", Cents: cents(4500), Edit: Edit{Status: "billed"}},
		{Key: "estimated", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", AccountID: "a", Cents: cents(1300), Edit: Edit{Status: "estimated"}},
		{Key: "unknown", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", AccountID: "a", Edit: Edit{Status: "billed"}},
		{Key: "paid", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", Cents: cents(90000), Edit: Edit{Status: "paid"}},
		{Key: "cancelled", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", Cents: cents(90000), Edit: Edit{Status: "cancelled"}},
		{Key: "replaced", Kind: "schedule", Date: "2024-02-29", Currency: "CNY", Cents: cents(90000), Edit: Edit{Status: "replaced"}},
		{Key: "receipt", Kind: "receipt", Date: "2024-02-29", Currency: "CNY", Cents: cents(2600)},
		{Key: "transfer", Kind: "transfer", Date: "2024-02-29", Currency: "CNY", Cents: cents(90000)},
		{Key: "income", Kind: "income", Date: "2024-02-29", Currency: "CNY", Cents: cents(90000)},
		{Key: "usd", Kind: "expense", Date: "2024-02-29", Currency: "USD", Cents: cents(99000)},
		{Key: "next", Kind: "expense", Date: "2024-03-01", Currency: "CNY", Cents: cents(99000)},
	}
	q := CalendarQuery{Month: "2024-02", Day: "2024-02-29", View: "all", Filter: Filter{Currency: "CNY"}}
	days, total, detail, err := BuildCalendar(rows, q)
	require.NoError(t, err)
	require.Len(t, days, 29)
	require.Zero(t, days[0].Gross)
	require.Equal(t, int64(2600), total.Gross)
	require.Equal(t, int64(600), total.Refund)
	require.Equal(t, int64(2000), total.Net)
	require.Equal(t, int64(4500), total.Billed)
	require.Equal(t, int64(1300), total.Estimated)
	require.Equal(t, 2, total.BilledCount)
	require.Equal(t, 1, total.UnknownRepayments)
	require.Equal(t, total, days[28].CalendarMetric)
	require.Len(t, detail, 5)

	q.View = "repayment"
	_, total, detail, err = BuildCalendar(rows, q)
	require.NoError(t, err)
	require.Zero(t, total.Gross)
	require.Len(t, detail, 3)
	q.View = "spending"
	_, total, detail, err = BuildCalendar(rows, q)
	require.NoError(t, err)
	require.Zero(t, total.Billed)
	require.Len(t, detail, 2)
	q.Account = "another-account"
	_, total, detail, err = BuildCalendar(rows, q)
	require.NoError(t, err)
	require.Equal(t, CalendarMetric{}, total)
	require.Empty(t, detail)
	require.NotNil(t, detail)
}

func TestCalendarEmptyMonthsAndValidation(t *testing.T) {
	q := CalendarQuery{Month: "2026-02", Day: "2026-02-01", View: "all", Filter: Filter{Currency: "CNY"}}
	days, _, rows, err := BuildCalendar(nil, q)
	require.NoError(t, err)
	require.Len(t, days, 28)
	require.NotNil(t, rows)
	for _, change := range []func(*CalendarQuery){
		func(q *CalendarQuery) { q.Month = "2026-13" },
		func(q *CalendarQuery) { q.Day = "2026-02-29" },
		func(q *CalendarQuery) { q.Day = "2026-03-01" },
		func(q *CalendarQuery) { q.Currency = "EUR" },
		func(q *CalendarQuery) { q.View = "history" },
	} {
		bad := q
		change(&bad)
		_, _, _, err := BuildCalendar(nil, bad)
		require.Error(t, err)
	}
}
