package timeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"

	"github.com/google/uuid"
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
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, fromDate, toDate)
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

	// mode = with_future 时，叠加未来项目
	if mode == TimelineModeWithFuture {
		// 叠加周期收入
		recurringPlans, err := store.RecurringIncomes().ListByUser(ctx, userID)
		if err == nil {
			for _, plan := range recurringPlans {
				if !plan.Active {
					continue
				}
				projectedDates := projectRecurringIncome(plan, from, to)
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

		// 叠加分期还款
		accounts, err := store.Accounts().ListByUser(ctx, userID)
		if err == nil {
			for _, acc := range accounts {
				if acc.Archived {
					continue
				}
				plans, err := store.Installments().ListByAccount(ctx, userID, acc.ID)
				if err != nil {
					continue
				}
				for _, plan := range plans {
					if plan.Status != installment.InstallmentStatusActive {
						continue
					}
					schedule := calculateInstallmentSchedule(plan, from, to)
					for _, item := range schedule {
						dateStr := item.DueDate
						if daily, ok := dailyMap[dateStr]; ok {
							daily.Expense += item.Total
							daily.Net = daily.Income - daily.Expense
							daily.HasFuture = true
						}
					}
				}
			}
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

// BuildDailyDetail 构建每日明细
func BuildDailyDetail(
	ctx context.Context,
	store repository.Store,
	userID string,
	date string,
	mode TimelineMode,
) (DailyDetailResponse, error) {
	// 解析日期
	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("invalid date: %w", err)
	}

	// 计算该日期的起止时间
	startTime := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, time.UTC)
	endTime := startTime.AddDate(0, 0, 1)
	startStr := startTime.Format("2006-01-02")
	endStr := endTime.Format("2006-01-02")

	// 获取真实交易
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, startStr, endStr)
	if err != nil {
		return DailyDetailResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	var result []ledger.Transaction

	// 过滤出该日期的交易
	for _, tx := range transactions {
		txDate, err := time.Parse("2006-01-02", tx.OccurredAt[:10])
		if err == nil && txDate.Equal(targetDate) {
			result = append(result, *tx)
		}
	}

	// mode = with_future 时，添加虚拟交易
	if mode == TimelineModeWithFuture {
		// 添加周期收入虚拟交易
		recurringPlans, err := store.RecurringIncomes().ListByUser(ctx, userID)
		if err == nil {
			for _, plan := range recurringPlans {
				if !plan.Active {
					continue
				}
				nextDue, err := time.Parse("2006-01-02", plan.NextDueDate)
				if err == nil && nextDue.Equal(targetDate) {
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

		// 添加分期还款虚拟交易
		accounts, err := store.Accounts().ListByUser(ctx, userID)
		if err == nil {
			for _, acc := range accounts {
				if acc.Archived {
					continue
				}
				plans, err := store.Installments().ListByAccount(ctx, userID, acc.ID)
				if err != nil {
					continue
				}
				for _, plan := range plans {
					if plan.Status != installment.InstallmentStatusActive {
						continue
					}
					schedule := calculateInstallmentSchedule(plan, startTime, endTime)
					for _, item := range schedule {
						if item.DueDate == date {
							virtualTx := ledger.Transaction{
								ID: uuid.NewString(),
								CreateTransactionRequest: ledger.CreateTransactionRequest{
									Type:        ledger.TransactionTypeExpense,
									Amount:      item.Total,
									CategoryKey: "",
									Remark:      fmt.Sprintf("预计支出：分期还款（第%d期）", item.Period),
									OccurredAt:  date + "T00:00:00Z",
									AccountID:   plan.AccountID,
								},
							}
							result = append(result, virtualTx)
						}
					}
				}
			}
		}
	}

	return DailyDetailResponse{
		Date:         date,
		Mode:         mode,
		Transactions: result,
	}, nil
}

// projectRecurringIncome 推算周期收入在未来日期范围内的收入日期
func projectRecurringIncome(plan *recurringincome.RecurringIncomePlan, startDate, endDate time.Time) []time.Time {
	var dates []time.Time
	nextDue, err := time.Parse("2006-01-02", plan.NextDueDate)
	if err != nil {
		return dates
	}

	current := nextDue
	for current.Before(endDate) || current.Equal(endDate) {
		if current.After(startDate) || current.Equal(startDate) {
			dates = append(dates, current)
		}

		// 计算下一次日期
		switch plan.Frequency {
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

// calculateInstallmentSchedule 计算分期计划在指定日期范围内的还款计划
func calculateInstallmentSchedule(plan *installment.InstallmentPlan, startDate, endDate time.Time) []installment.InstallmentPreviewItem {
	req := installment.PreviewInstallmentRequest{
		Principal:          plan.Principal,
		TermMonths:         plan.TermMonths,
		MonthlyRatePercent: plan.MonthlyRate * 100,
		RepayMethod:        plan.RepayMethod,
		FirstDueDate:       plan.FirstDueDate,
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
