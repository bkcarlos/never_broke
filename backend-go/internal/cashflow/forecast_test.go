package cashflow

import (
	"github.com/stretchr/testify/require"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/workspace"
	"testing"
	"time"
)

func date(s string) time.Time { v, _ := time.Parse(time.DateOnly, s); return v }
func entry(key, kind, day string, value int64) workspace.Entry {
	return workspace.Entry{Key: key, Kind: kind, Date: day, Currency: "CNY", AccountID: "bank", Cents: &value, Description: key}
}
func fixture() ([]models.Account, []workspace.Setting) {
	return []models.Account{{ID: "bank", Currency: "CNY", Type: "bank", Balance: 1000}}, []workspace.Setting{{Kind: "balance", Match: "bank", Currency: "CNY", Enabled: true}}
}
func request() CashflowForecastRequest {
	return CashflowForecastRequest{StartDate: "2026-11-01", Days: 30, Base: ForecastBaseTotalAssets, SafeAmount: 200}
}

func TestMonthlySurplusCanHavePaymentDateGap(t *testing.T) {
	accounts, settings := fixture()
	settings = append(settings, workspace.Setting{Key: "rent", Kind: "fixed_expense", Name: "房租", Currency: "CNY", Cents: 300000, Date: "2026-11-10", Period: "monthly", Enabled: true})
	income := []incomeEvent{{Date: "2026-11-15", Name: "工资", AccountID: "bank", Currency: "CNY", Cents: 1000000}}
	result, err := buildForecast(request(), "CNY", nil, accounts, settings, income, map[string]int64{"2026-11": 300000})
	require.NoError(t, err)
	require.Len(t, result.Months, 1)
	m := result.Months[0]
	require.Equal(t, 7000.0, m.Net)
	require.Equal(t, "2026-11-10", m.FirstGapDate)
	require.Equal(t, 2000.0, *m.GapAmount)
	require.Equal(t, 8000.0, *m.ClosingBalance)
	require.Equal(t, ForecastStatusDanger, result.Status)
	require.Equal(t, "fixed", result.Points[9].Events[0].Kind)
}

func TestUnknownBalanceStillShowsMonthlyIncomeAndExpenses(t *testing.T) {
	accounts, _ := fixture()
	result, err := buildForecast(request(), "CNY", []workspace.Entry{entry("bill", "schedule", "2026-11-10", 3000)}, accounts, nil, []incomeEvent{{Date: "2026-11-15", Currency: "CNY", Cents: 100000}}, nil)
	require.NoError(t, err)
	require.False(t, result.BalanceKnown)
	require.Nil(t, result.Months[0].MinBalance)
	require.Nil(t, result.Months[0].GapAmount)
	require.Equal(t, 1000.0, result.Months[0].Income)
	require.Equal(t, 30.0, result.Months[0].Repayment)
	require.Empty(t, result.DangerDates)
	require.Equal(t, ForecastStatus("unknown"), result.Status)
}

func TestObservedSpendingDoesNotStopWhenBudgetExhausted(t *testing.T) {
	accounts, settings := fixture()
	req := request()
	req.StartDate = "2026-11-04"
	req.Days = 28
	items := []workspace.Entry{entry("food", "expense", "2026-10-01", 310000), entry("food2", "expense", "2026-11-02", 60000)}
	result, err := buildForecast(req, "CNY", items, accounts, settings, nil, map[string]int64{"2026-11": 10000})
	require.NoError(t, err)
	require.Equal(t, 130.0, result.DailyBudget) // 100/day history, 200/day this month.
	require.Equal(t, 130.0, result.Points[len(result.Points)-1].Outflow)
	require.Equal(t, "history", result.Months[0].SpendingSource)
	require.Equal(t, 600.0, result.Months[0].ActualOutflow)
	require.Equal(t, 870.0, *result.Points[0].ClosingBalance) // Actuals are already in the balance.
}

func TestFixedHistoryAndRefundsExcludedFromVariableEstimate(t *testing.T) {
	_, settings := fixture()
	settings = append(settings, workspace.Setting{Kind: "fixed_expense", Name: "rent", Currency: "CNY", Enabled: true})
	items := []workspace.Entry{entry("rent", "expense", "2026-10-01", 300000), entry("food", "expense", "2026-10-02", 62000), entry("food refund", "refund", "2026-10-03", 31000), entry("transfer", "transfer", "2026-10-02", 900000)}
	rate, source, days := spendingRate(items, settings, "CNY", "", date("2026-11-01"))
	require.Equal(t, int64(1000), rate)
	require.Equal(t, "history", source)
	require.Equal(t, 31, days)
}

func TestPaidDisabledOtherCurrencyAndUnknownSchedules(t *testing.T) {
	accounts, settings := fixture()
	paid := entry("paid", "schedule", "2026-11-05", 99999)
	paid.Edit.Status = "paid"
	replaced := paid
	replaced.Key = "replaced"
	replaced.Edit.Status = "replaced"
	usd := entry("usd", "schedule", "2026-11-05", 99999)
	usd.Currency = "USD"
	unknown := entry("unknown", "schedule", "2026-11-05", 0)
	unknown.Cents = nil
	settings = append(settings, workspace.Setting{Kind: "fixed_expense", Name: "off", Cents: 50000, Currency: "CNY", Date: "2026-11-10", Period: "monthly", Enabled: false})
	r, err := buildForecast(request(), "CNY", []workspace.Entry{paid, replaced, usd, unknown}, accounts, settings, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1, r.UnknownAmounts)
	require.False(t, r.BalanceKnown)
	require.Zero(t, r.Months[0].Outflow)
	require.Nil(t, r.Points[4].ClosingBalance)
}

func TestActualCreditConsumptionIsNotCountedAgainWithRepayment(t *testing.T) {
	accounts, settings := fixture()
	accounts = append(accounts, models.Account{ID: "credit", Type: "credit", Currency: "CNY"}, models.Account{ID: "bank2", Type: "bank", Currency: "CNY"})
	settings = append(settings, workspace.Setting{Kind: "balance", Match: "bank2"})
	req := request()
	req.StartDate = "2026-11-04"
	req.Days = 1
	purchase := entry("purchase", "expense", "2026-11-03", 100000)
	purchase.AccountID = "credit"
	transfer := entry("transfer", "transfer", "2026-11-03", 50000)
	transfer.FromAccountID = "bank"
	transfer.ToAccountID = "bank2"
	repayment := entry("repayment", "transfer", "2026-11-03", 100000)
	repayment.FromAccountID = "bank"
	repayment.ToAccountID = "credit"
	result, err := buildForecast(req, "CNY", []workspace.Entry{purchase, transfer, repayment}, accounts, settings, nil, nil)
	require.NoError(t, err)
	require.Equal(t, 1000.0, result.Months[0].ActualOutflow)
	require.Zero(t, result.Months[0].ActualIncome)
}

func TestRealizedTodayNotAppliedToBalanceTwice(t *testing.T) {
	accounts, settings := fixture()
	settings = append(settings, workspace.Setting{Kind: "fixed_expense", Key: "rent", Name: "rent", Cents: 30000, Currency: "CNY", AccountID: "bank", Date: "2026-11-01", Period: "monthly", Enabled: true})
	items := []workspace.Entry{entry("rent", "expense", "2026-11-01", 30000), entry("salary", "income", "2026-11-01", 100000)}
	req := request()
	req.Days = 1
	result, err := buildForecast(req, "CNY", items, accounts, settings, []incomeEvent{{Date: req.StartDate, Name: "salary", Currency: "CNY", AccountID: "bank", Cents: 100000}}, nil)
	require.NoError(t, err)
	require.Zero(t, result.Points[0].Inflow)
	require.Zero(t, result.Points[0].Outflow)
	require.Equal(t, 1000.0, *result.Points[0].ClosingBalance)
	require.Equal(t, 1000.0, result.Months[0].ActualIncome)
	require.Equal(t, 300.0, result.Months[0].ActualOutflow)
}

func TestSingleAccountDoesNotIncludeOtherIncomeAndUnassignedDebt(t *testing.T) {
	accounts, settings := fixture()
	req := request()
	req.Base = ForecastBaseAccount
	req.AccountID = "bank"
	incomes := []incomeEvent{{Date: req.StartDate, Currency: "CNY", AccountID: "other", Cents: 900000}}
	result, err := buildForecast(req, "CNY", []workspace.Entry{entry("card bill", "schedule", "2026-11-15", 30000)}, accounts, settings, incomes, nil)
	require.NoError(t, err)
	require.Zero(t, result.Months[0].Income)
	require.Equal(t, 300.0, result.Months[0].Repayment)
	debt := entry("card bill", "schedule", "2026-11-15", 30000)
	debt.AccountID = "credit"
	result, err = buildForecast(req, "CNY", []workspace.Entry{debt}, accounts, settings, incomes, nil)
	require.NoError(t, err)
	require.False(t, result.BalanceKnown)
	require.Zero(t, result.Months[0].Repayment)
}

func TestMonthlyBudgetFallbackUsesEachMonthAndReservesFixedCost(t *testing.T) {
	accounts, settings := fixture()
	settings = append(settings, workspace.Setting{Kind: "subscription", Name: "subscription", Cents: 3000, Currency: "CNY", Date: "2026-11-01", Period: "monthly", Enabled: true})
	req := request()
	req.Days = 61
	result, err := buildForecast(req, "CNY", nil, accounts, settings, nil, map[string]int64{"2026-11": 33000, "2026-12": 65000})
	require.NoError(t, err)
	require.Equal(t, 300.0, result.Months[0].DailyExpense)
	require.Equal(t, 620.0, result.Months[1].DailyExpense)
	require.Equal(t, 30.0, result.Months[1].FixedExpense)
	require.False(t, result.Months[1].Partial)
}

func TestNoHistoryDoesNotClaimSafe(t *testing.T) {
	accounts, settings := fixture()
	r, err := buildForecast(request(), "CNY", nil, accounts, settings, nil, nil)
	require.NoError(t, err)
	require.True(t, r.BalanceKnown)
	require.False(t, r.SpendingKnown)
	require.Equal(t, ForecastStatus("unknown"), r.Status)
}

func TestSalaryForecastUsesMonthlyNetPayAndResetsTaxAcrossYears(t *testing.T) {
	gross := 30000.0
	day := 15
	start := date("2026-01-01")
	rate := 0.1
	plan := models.RecurringIncomePlan{ID: "salary", Type: "salary", Label: "工资", Currency: "CNY", AccountID: "bank", Active: true, SalaryPreTaxMonthly: &gross, SalaryPayDay: &day, SalaryStartDate: &start, SalarySocialRate: &rate, Amount: 1}
	rows, err := incomeEvents([]models.RecurringIncomePlan{plan}, date("2026-12-01"), date("2027-02-01"), payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Len(t, rows, 2)
	require.Greater(t, rows[1].Cents, rows[0].Cents)
	require.NotEqual(t, int64(100), rows[0].Cents)
	versions, err := recurringincome.SalaryVersions(plan)
	require.NoError(t, err)
	preview, err := recurringincome.PreviewSalaryHistory(versions, 2026, payroll.NewService(nil).GetRegionRuleAt)
	require.NoError(t, err)
	require.Equal(t, cents(preview[11].PostTax), rows[0].Cents)
}

func TestRecurringIncomeRetainsMonthEndAndQuarterlyCadence(t *testing.T) {
	day := 31
	for _, frequency := range []string{"monthly", "quarterly"} {
		plan := models.RecurringIncomePlan{ID: "income", Active: true, Currency: "CNY", Frequency: frequency, NextDueDate: date("2027-02-28"), DayOfMonth: &day, Amount: 100}
		rows, err := incomeEvents([]models.RecurringIncomePlan{plan}, date("2027-03-29"), date("2027-06-01"), nil)
		require.NoError(t, err)
		if frequency == "monthly" {
			require.Len(t, rows, 3)
			require.Equal(t, "2027-03-31", rows[0].Date)
			require.Equal(t, "2027-04-30", rows[1].Date)
		} else {
			require.Len(t, rows, 1)
			require.Equal(t, "2027-05-31", rows[0].Date)
		}
	}
}

func TestSingleAccountDoesNotUseGlobalBudget(t *testing.T) {
	accounts, settings := fixture()
	req := request()
	req.Base = ForecastBaseAccount
	req.AccountID = "bank"
	result, err := buildForecast(req, "CNY", nil, accounts, settings, nil, map[string]int64{"2026-11": 300000})
	require.NoError(t, err)
	require.False(t, result.SpendingKnown)
	require.Zero(t, result.Months[0].DailyExpense)
	require.Equal(t, ForecastStatus("unknown"), result.Status)
}
