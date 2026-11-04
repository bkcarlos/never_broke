package cashflow

import (
	"context"
	"math"
	"time"

	"gorm.io/gorm"
	"neverbroke/internal/models"
	"neverbroke/internal/money"
	"neverbroke/internal/payroll"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
	"neverbroke/internal/workspace"
)

// BuildForecast supports repository-only callers without inventing verified
// balances or future plans that are available only in persistent storage.
func BuildForecast(ctx context.Context, store repository.Store, userID string, req CashflowForecastRequest) (CashflowForecastResponse, error) {
	start, err := validateRequest(req)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	sourceAccounts, err := store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	accounts := []models.Account{}
	for _, a := range sourceAccounts {
		accounts = append(accounts, models.Account{ID: a.ID, Type: string(a.Type), Currency: a.Currency, Balance: a.Balance, Archived: a.Archived})
	}
	first := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, first.AddDate(0, -3, 0).Format(time.DateOnly), start.AddDate(0, 0, 1).Format(time.DateOnly))
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	items := []workspace.Entry{}
	for _, tx := range transactions {
		if tx.Projected {
			continue
		}
		value := cents(tx.Amount)
		var valuePtr *int64 = &value
		if tx.AmountUnknown {
			valuePtr = nil
		}
		kind := string(tx.Type)
		if kind == "expense" && value < 0 {
			kind = "refund"
			value = -value
		}
		date := tx.OccurredAt
		if len(date) > 10 {
			date = date[:10]
		}
		txCurrency := tx.Currency
		if txCurrency == "" {
			txCurrency = money.Currency(ctx)
		}
		items = append(items, workspace.Entry{Key: tx.ID, Kind: kind, Date: date, Currency: txCurrency, AccountID: tx.AccountID, FromAccountID: tx.FromAccountID, ToAccountID: tx.ToAccountID, Cents: valuePtr, Description: tx.Remark, Metadata: map[string]any{"分期计划": tx.InstallmentPlanID}})
	}
	result, err := buildForecast(req, money.Currency(ctx), items, accounts, nil, nil, nil)
	result.Warnings = append(result.Warnings, "当前存储模式无法读取未来收入、固定支出和还款计划，收支估算不完整；请使用数据库模式。")
	return result, err
}

func BuildForecastWithDB(ctx context.Context, store repository.Store, db *gorm.DB, userID string, req CashflowForecastRequest) (CashflowForecastResponse, error) {
	start, err := validateRequest(req)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	db = db.WithContext(ctx)
	items, accounts, err := workspace.Load(db, userID)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	var settings []workspace.Setting
	if err = db.Where("user_id = ?", userID).Order("key").Find(&settings).Error; err != nil {
		return CashflowForecastResponse{}, err
	}
	var plans []models.RecurringIncomePlan
	if err = db.Where("user_id = ? AND active = ? AND currency = ?", userID, true, money.Currency(ctx)).Order("id").Find(&plans).Error; err != nil {
		return CashflowForecastResponse{}, err
	}
	incomes, err := incomeEvents(plans, start, start.AddDate(0, 0, req.Days), recurringincome.PolicyLookup(payroll.NewService(payroll.NewRepository(db)).GetRegionRuleAt))
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	budgets := map[string]int64{}
	first := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	if money.Currency(ctx) == "CNY" {
		for month := first; month.Before(start.AddDate(0, 0, req.Days)); month = month.AddDate(0, 1, 0) {
			b, err := store.Budgets().GetByMonth(ctx, userID, month.Format("2006-01"))
			if err != nil {
				return CashflowForecastResponse{}, err
			}
			if b != nil && b.TotalBudget > 0 {
				budgets[month.Format("2006-01")] = int64(math.Round(b.TotalBudget * 100))
			}
		}
	}
	return buildForecast(req, money.Currency(ctx), items, accounts, settings, incomes, budgets)
}

type incomeEvent struct {
	Date, Name, AccountID, Currency, Warning string
	Cents                                    int64
}

func incomeEvents(plans []models.RecurringIncomePlan, start, end time.Time, lookup recurringincome.PolicyLookup) ([]incomeEvent, error) {
	out := []incomeEvent{}
	for _, p := range plans {
		if !p.Active {
			continue
		}
		versions, err := recurringincome.SalaryVersions(p)
		if err != nil {
			return nil, err
		}
		if p.Type == "salary" && len(versions) > 0 {
			for year := start.Year(); year <= end.Year(); year++ {
				rows, err := recurringincome.PreviewSalaryHistory(versions, year, lookup)
				if err != nil {
					return nil, err
				}
				for _, r := range rows {
					if r.Date < start.Format(time.DateOnly) || r.Date >= end.Format(time.DateOnly) {
						continue
					}
					if p.LastFiredDate != nil && r.Date <= p.LastFiredDate.Format(time.DateOnly) {
						continue
					}
					out = append(out, incomeEvent{Date: r.Date, Name: p.Label, AccountID: r.Config.AccountID, Currency: r.Config.Currency, Cents: int64(math.Round(r.PostTax * 100)), Warning: r.Warning})
				}
			}
		} else {
			first := p.NextDueDate
			// Generate from the first day so an anchored 31st is not lost when
			// NextDueDate was clamped to the 28th in a previous short month.
			from := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
			for _, date := range workspace.RecurrenceDates(first, p.Frequency, "", from, end) {
				if p.DayOfMonth != nil && *p.DayOfMonth >= 1 && *p.DayOfMonth <= 31 && p.Frequency != "weekly" {
					d, _ := time.Parse(time.DateOnly, date)
					month := time.Date(d.Year(), d.Month(), 1, 0, 0, 0, 0, time.UTC)
					date = month.AddDate(0, 0, min(*p.DayOfMonth, month.AddDate(0, 1, -1).Day())-1).Format(time.DateOnly)
				}
				if date < start.Format(time.DateOnly) || date >= end.Format(time.DateOnly) {
					continue
				}
				if date < p.NextDueDate.Format(time.DateOnly) || (p.LastFiredDate != nil && date <= p.LastFiredDate.Format(time.DateOnly)) {
					continue
				}
				out = append(out, incomeEvent{Date: date, Name: p.Label, AccountID: p.AccountID, Currency: p.Currency, Cents: int64(math.Round(p.Amount * 100))})
			}
		}
	}
	return out, nil
}
