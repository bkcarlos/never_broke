package timeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	"neverbroke/internal/curatedimport"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	"neverbroke/internal/money"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
	"neverbroke/internal/workspace"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BuildDailySummary 构建每日汇总
func BuildDailySummary(
	ctx context.Context,
	store repository.Store,
	userID string,
	fromDate string,
	toDate string,
	mode TimelineMode,
) (DailyTimelineResponse, error) {
	// 解析日期
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("invalid fromDate: %w", err)
	}
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("invalid toDate: %w", err)
	}

	// 从 transactions 中取 [from, to] 的所有真实交易
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, fromDate, to.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	// 构建日期到汇总的映射
	dailyMap := make(map[string]*DailySummary)
	current := from
	for current.Before(to) || current.Equal(to) {
		dateStr := current.Format("2006-01-02")
		dailyMap[dateStr] = &DailySummary{
			Date:      dateStr,
			Income:    0,
			Expense:   0,
			Net:       0,
			HasFuture: false,
		}
		current = current.AddDate(0, 0, 1)
	}

	// 处理真实交易
	for _, tx := range transactions {
		dateStr := tx.OccurredAt[:10] // 提取 YYYY-MM-DD
		if daily, ok := dailyMap[dateStr]; ok {
			switch tx.Type {
			case ledger.TransactionTypeIncome:
				daily.Income += tx.Amount
			case ledger.TransactionTypeExpense:
				daily.Expense += tx.Amount
			}
			daily.Net = daily.Income - daily.Expense
		}
	}

	// 转换为列表并排序
	var days []DailySummary
	var dates []string
	for date := range dailyMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		days = append(days, *dailyMap[date])
	}

	return DailyTimelineResponse{
		FromDate: fromDate,
		ToDate:   toDate,
		Mode:     mode,
		Days:     days,
	}, nil
}

// BuildDailySummaryWithDB 构建每日汇总（支持未来预测）
func BuildDailySummaryWithDB(
	ctx context.Context,
	store repository.Store,
	db *gorm.DB,
	userID string,
	fromDate string,
	toDate string,
	mode TimelineMode,
) (DailyTimelineResponse, error) {
	from, err := time.Parse("2006-01-02", fromDate)
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("invalid fromDate: %w", err)
	}
	to, err := time.Parse("2006-01-02", toDate)
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("invalid toDate: %w", err)
	}

	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, fromDate, to.AddDate(0, 0, 1).Format("2006-01-02"))
	if err != nil {
		return DailyTimelineResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	dailyMap := make(map[string]*DailySummary)
	current := from
	for current.Before(to) || current.Equal(to) {
		dateStr := current.Format("2006-01-02")
		dailyMap[dateStr] = &DailySummary{
			Date:      dateStr,
			Income:    0,
			Expense:   0,
			Net:       0,
			HasFuture: false,
		}
		current = current.AddDate(0, 0, 1)
	}

	for _, tx := range transactions {
		dateStr := tx.OccurredAt[:10]
		if daily, ok := dailyMap[dateStr]; ok {
			switch tx.Type {
			case ledger.TransactionTypeIncome:
				daily.Income += tx.Amount
			case ledger.TransactionTypeExpense:
				daily.Expense += tx.Amount
			}
			daily.Net = daily.Income - daily.Expense
		}
	}

	// Imported bills are repayment obligations, separate from consumption.
	if mode == TimelineModeWithFuture && db != nil {
		schedules, err := importedSchedules(ctx, db, userID, fromDate, toDate)
		if err != nil {
			return DailyTimelineResponse{}, err
		}
		for _, r := range schedules {
			if d := dailyMap[r.Date]; d != nil {
				d.HasFuture = true
				if r.AmountCents != nil {
					d.Repayment += float64(*r.AmountCents) / 100
				}
			}
		}
	}
	// 未来模式：叠加周期收入和分期还款
	if mode == TimelineModeWithFuture && db != nil {
		// 叠加周期收入
		var recurringPlans []models.RecurringIncomePlan
		db.Where("user_id = ? AND active = ? AND currency = ?", userID, true, money.Currency(ctx)).Find(&recurringPlans)
		for _, plan := range recurringPlans {
			projectedDates := projectRecurringIncomeFromDB(&plan, from, to)
			for _, date := range projectedDates {
				dateStr := date.Format("2006-01-02")
				if daily, ok := dailyMap[dateStr]; ok {
					daily.Income += plan.Amount
					daily.Net = daily.Income - daily.Expense
					daily.HasFuture = true
				}
			}
		}

	}

	var days []DailySummary
	var dates []string
	for date := range dailyMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		days = append(days, *dailyMap[date])
	}

	return DailyTimelineResponse{
		FromDate: fromDate,
		ToDate:   toDate,
		Mode:     mode,
		Days:     days,
	}, nil
}

// BuildDailyDetail 构建每日明细
func BuildDailyDetail(
	ctx context.Context,
	store repository.Store,
	userID string,
	date string,
	mode TimelineMode,
) (DailyDetailResponse, error) {
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("invalid date: %w", err)
	}

	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 0, 1)
	startStr := startTime.Format("2006-01-02")
	endStr := endTime.Format("2006-01-02")

	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, startStr, endStr)
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	var result []ledger.Transaction
	for _, tx := range transactions {
		txDate, err := time.Parse("2006-01-02", tx.OccurredAt[:10])
		if err == nil && txDate.Equal(targetDate) {
			result = append(result, *tx)
		}
	}

	return DailyDetailResponse{
		Date:         date,
		Mode:         mode,
		Transactions: result,
	}, nil
}

// BuildDailyDetailWithDB 构建每日明细（支持未来预测）
func BuildDailyDetailWithDB(
	ctx context.Context,
	store repository.Store,
	db *gorm.DB,
	userID string,
	date string,
	mode TimelineMode,
) (DailyDetailResponse, error) {
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("invalid date: %w", err)
	}

	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 0, 1)

	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	var result []ledger.Transaction
	for _, tx := range transactions {
		txDate, err := time.Parse("2006-01-02", tx.OccurredAt[:10])
		if err == nil && txDate.Equal(targetDate) {
			result = append(result, *tx)
		}
	}

	if mode == TimelineModeWithFuture && db != nil {
		schedules, err := importedSchedules(ctx, db, userID, date, date)
		if err != nil {
			return DailyDetailResponse{}, err
		}
		for _, r := range schedules {
			var amount float64
			if r.AmountCents != nil {
				amount = float64(*r.AmountCents) / 100
			}
			result = append(result, ledger.Transaction{ID: "schedule:" + r.Key, Currency: r.Currency, Projected: true, AmountUnknown: r.AmountCents == nil, CreateTransactionRequest: ledger.CreateTransactionRequest{Type: ledger.TransactionTypeTransfer, Amount: amount, OccurredAt: date + "T00:00:00+08:00", Remark: r.Description + "\n" + r.Note}})
		}
	}
	// 未来模式：添加虚拟交易
	if mode == TimelineModeWithFuture && db != nil {
		// 添加周期收入虚拟交易
		var recurringPlans []models.RecurringIncomePlan
		db.Where("user_id = ? AND active = ? AND currency = ?", userID, true, money.Currency(ctx)).Find(&recurringPlans)
		for _, plan := range recurringPlans {
			if plan.NextDueDate.Equal(targetDate) {
				virtualTx := ledger.Transaction{
					ID: uuid.NewString(),
					CreateTransactionRequest: ledger.CreateTransactionRequest{
						Type:        ledger.TransactionTypeIncome,
						Amount:      plan.Amount,
						CategoryKey: "",
						Remark:      fmt.Sprintf("预计收入：%s", plan.Label),
						OccurredAt:  date + "T00:00:00Z",
						AccountID:   plan.AccountID,
					},
				}
				result = append(result, virtualTx)
			}
		}

	}

	return DailyDetailResponse{
		Date:         date,
		Mode:         mode,
		Transactions: result,
	}, nil
}

// projectRecurringIncomeFromDB 推算周期收入日期
func projectRecurringIncomeFromDB(plan *models.RecurringIncomePlan, startDate, endDate time.Time) []time.Time {
	var dates []time.Time
	current := plan.NextDueDate

	for current.Before(endDate) || current.Equal(endDate) {
		if current.After(startDate) || current.Equal(startDate) {
			dates = append(dates, current)
		}

		switch recurringincome.RecurringIncomeFrequency(plan.Frequency) {
		case recurringincome.FrequencyMonthly:
			current = current.AddDate(0, 1, 0)
		case recurringincome.FrequencyBimonthly:
			current = current.AddDate(0, 2, 0)
		case recurringincome.FrequencyQuarterly:
			current = current.AddDate(0, 3, 0)
		case recurringincome.FrequencySemiannual:
			current = current.AddDate(0, 6, 0)
		case recurringincome.FrequencyWeekly:
			current = current.AddDate(0, 0, 7)
		case recurringincome.FrequencyYearly:
			current = current.AddDate(1, 0, 0)
		default:
			return dates
		}
	}

	return dates
}

// calculateInstallmentScheduleFromDB 计算分期还款计划
func calculateInstallmentScheduleFromDB(plan *models.InstallmentPlan, startDate, endDate time.Time) []installment.InstallmentPreviewItem {
	req := installment.PreviewInstallmentRequest{
		Principal:          plan.Principal,
		TermMonths:         plan.TermMonths,
		MonthlyRatePercent: plan.MonthlyRate * 100,
		RepayMethod:        installment.InstallmentRepayMethod(plan.RepayMethod),
		FirstDueDate:       plan.FirstDueDate.Format("2006-01-02"),
		Currency:           plan.Currency,
	}

	preview, err := installment.BuildSchedule(req)
	if err != nil {
		return nil
	}

	var result []installment.InstallmentPreviewItem
	for _, item := range preview.Schedule {
		dueDate, err := time.Parse("2006-01-02", item.DueDate)
		if err != nil {
			continue
		}
		if (dueDate.After(startDate) || dueDate.Equal(startDate)) && dueDate.Before(endDate) {
			result = append(result, item)
		}
	}

	return result
}

// importedSchedules keeps estimates and payment-status-unknown bills out of
// expense totals. They are visible only when the user requests future items.
func importedSchedules(ctx context.Context, db *gorm.DB, user, from, to string) ([]curatedimport.Record, error) {
	items, _, err := workspace.Load(db.WithContext(ctx), user)
	if err != nil {
		return nil, err
	}
	out := []curatedimport.Record{}
	for _, x := range items {
		if x.Kind != "schedule" || x.Currency != money.Currency(ctx) || x.Date < from || x.Date > to || x.Edit.Status == "paid" || x.Edit.Status == "cancelled" || x.Edit.Status == "replaced" {
			continue
		}
		out = append(out, curatedimport.Record{Key: x.Key, Kind: "schedule", Currency: x.Currency, Date: x.Date, AmountCents: x.Cents, Description: x.Description, Note: x.Note})
	}
	return out, nil
}
