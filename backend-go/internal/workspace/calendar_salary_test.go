package workspace

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"neverbroke/internal/models"
	"neverbroke/internal/recurringincome"
	"testing"
	"time"
)

func TestCalendarSalaryFiltersAndHistoricalCumulativeTax(t *testing.T) {
	start, _ := time.Parse(time.DateOnly, "2026-01-01")
	gross, rate, extra, day := 20000.0, 0.2, 1000.0, 31
	plan := models.RecurringIncomePlan{ID: "p", Type: "salary", Active: true, Currency: "CNY", AccountID: "a", Frequency: "monthly", SalaryPreTaxMonthly: &gross, SalaryStartDate: &start, SalaryPayDay: &day, SalarySocialRate: &rate, SalaryExtraDeductions: &extra}
	q := CalendarQuery{Month: "2026-02", Day: "2026-02-28", View: "spending", Filter: Filter{Currency: "CNY"}}
	result, err := buildCalendarSalaries([]models.RecurringIncomePlan{plan}, q)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "2026-02-28", result[0].Date)
	require.Equal(t, 40000.0, result[0].CumulativePreTax)
	require.Equal(t, 600.0, result[0].CumulativeTax)
	require.Equal(t, 15700.0, result[0].PostTax)
	require.Nil(t, result[0].Contributions)
	require.Equal(t, "salary:p:2026-02-28", result[0].Key)
	for _, change := range []func(*models.RecurringIncomePlan){
		func(p *models.RecurringIncomePlan) { p.Active = false },
		func(p *models.RecurringIncomePlan) { p.Type = "rent" },
		func(p *models.RecurringIncomePlan) { p.Currency = "USD" },
		func(p *models.RecurringIncomePlan) { p.SalaryPreTaxMonthly = nil },
		func(p *models.RecurringIncomePlan) { future := start.AddDate(0, 2, 0); p.SalaryStartDate = &future },
	} {
		other := plan
		change(&other)
		rows, err := buildCalendarSalaries([]models.RecurringIncomePlan{other}, q)
		require.NoError(t, err)
		require.Empty(t, rows)
		require.NotNil(t, rows)
	}
	q.Account = "another"
	rows, err := buildCalendarSalaries([]models.RecurringIncomePlan{plan}, q)
	require.NoError(t, err)
	require.Empty(t, rows)
	q.Account = "a"
	raw, _ := json.Marshal(recurringincome.SalaryContributions{SocialBase: 10000, HousingBase: 12000, PensionRate: 0.08, MedicalRate: 0.02, UnemploymentRate: 0.002, HousingRate: 0.12})
	config := string(raw)
	plan.SalaryContributions = &config
	rows, err = buildCalendarSalaries([]models.RecurringIncomePlan{plan}, q)
	require.NoError(t, err)
	require.Equal(t, 2460.0, rows[0].SocialSecurity)
	require.Equal(t, 1440.0, rows[0].Contributions.Housing)
	// Projected salaries never enter the spending/repayment accumulator.
	_, totals, _, err := BuildCalendar(nil, q)
	require.NoError(t, err)
	require.Equal(t, CalendarMetric{}, totals)
}
