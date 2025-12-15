package cashflow

import (
	"context"
	"fmt"
	"time"

	"neverbroke/internal/budget"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
)

// BuildForecast 构建现金流预测
func BuildForecast(ctx context.Context, store repository.Store, userID string, req CashflowForecastRequest) (CashflowForecastResponse, error) {
	// 1. 计算起始余额
	var startingBalance float64
	if req.Base == ForecastBaseTotalAssets {
		accounts, err := store.Accounts().ListByUser(ctx, userID)
		if err != nil {
			return CashflowForecastResponse{}, fmt.Errorf("failed to list accounts: %w", err)
		}
		for _, acc := range accounts {
			if !acc.Archived {
				startingBalance += acc.Balance
			}
		}
	} else if req.Base == ForecastBaseAccount {
		if req.AccountID == "" {
			return CashflowForecastResponse{}, fmt.Errorf("accountId is required when base is account")
		}
		acc, err := store.Accounts().GetByID(ctx, userID, req.AccountID)
		if err != nil {
			return CashflowForecastResponse{}, fmt.Errorf("failed to get account: %w", err)
		}
		startingBalance = acc.Balance
	} else {
		return CashflowForecastResponse{}, fmt.Errorf("invalid base: %s", req.Base)
	}

	// 2. 解析起始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return CashflowForecastResponse{}, fmt.Errorf("invalid startDate: %w", err)
	}

	// 3. 计算结束日期
	endDate := startDate.AddDate(0, 0, req.Days)

	// 4. 获取周期收入计划
	recurringPlans, err := store.RecurringIncomes().ListByUser(ctx, userID)
	if err != nil {
		return CashflowForecastResponse{}, fmt.Errorf("failed to list recurring incomes: %w", err)
	}

	// 5. 获取分期计划
	// 如果 base 是 account，只获取该账户的分期；否则获取所有账户的分期
	var installmentPlans []*installment.InstallmentPlan
	if req.Base == ForecastBaseAccount {
		plans, err := store.Installments().ListByAccount(ctx, userID, req.AccountID)
		if err != nil {
			return CashflowForecastResponse{}, fmt.Errorf("failed to list installments: %w", err)
		}
		for i := range plans {
			installmentPlans = append(installmentPlans, plans[i])
		}
	} else {
		// 获取所有账户的分期计划
		accounts, err := store.Accounts().ListByUser(ctx, userID)
		if err != nil {
			return CashflowForecastResponse{}, fmt.Errorf("failed to list accounts: %w", err)
		}
		for _, acc := range accounts {
			if !acc.Archived {
				plans, err := store.Installments().ListByAccount(ctx, userID, acc.ID)
				if err != nil {
					continue // 忽略错误，继续处理其他账户
				}
				for i := range plans {
					installmentPlans = append(installmentPlans, plans[i])
				}
			}
		}
	}

	// 6. 获取预算信息（用于日常消费预估）
	now := time.Now()
	currentMonth := fmt.Sprintf("%04d-%02d", now.Year(), int(now.Month()))
	monthlyBudget, _ := budget.GetMonthlyBudget(currentMonth)
	// 计算本月已消费金额（简化：从交易中统计）
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	monthStartStr := monthStart.Format("2006-01-02")
	monthEndStr := monthEnd.Format("2006-01-02")
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, monthStartStr, monthEndStr)
	var monthExpense float64
	if err == nil {
		for _, tx := range transactions {
			if tx.Type == ledger.TransactionTypeExpense {
				monthExpense += tx.Amount
			}
		}
	}
	// 计算本月剩余天数
	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	remainingDaysInMonth := lastDayOfMonth - now.Day()
	if remainingDaysInMonth <= 0 {
		remainingDaysInMonth = 1
	}
	dailyBudget := (monthlyBudget.TotalBudget - monthExpense) / float64(remainingDaysInMonth)
	if dailyBudget < 0 {
		dailyBudget = 0
	}

	// 7. 构建日期到流入/流出的映射
	type dayFlow struct {
		inflow  float64
		outflow float64
	}
	dayFlows := make(map[string]*dayFlow)

	// 7.1 处理周期收入
	for _, plan := range recurringPlans {
		if !plan.Active {
			continue
		}
		projectedDates := projectRecurringIncome(plan, startDate, endDate)
		for _, date := range projectedDates {
			dateStr := date.Format("2006-01-02")
			if dayFlows[dateStr] == nil {
				dayFlows[dateStr] = &dayFlow{}
			}
			// 如果是工资，需要计算税后金额（简化：直接使用 amount）
			dayFlows[dateStr].inflow += plan.Amount
		}
	}

	// 7.2 处理分期还款
	for _, plan := range installmentPlans {
		if plan.Status != installment.InstallmentStatusActive {
			continue
		}
		schedule := calculateInstallmentSchedule(plan, startDate, endDate)
		for _, item := range schedule {
			dateStr := item.DueDate
			if dayFlows[dateStr] == nil {
				dayFlows[dateStr] = &dayFlow{}
			}
			dayFlows[dateStr].outflow += item.Total
		}
	}

	// 7.3 处理日常消费预估（在整个预测期内按每天均摊）
	current := startDate
	for current.Before(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayFlows[dateStr] == nil {
			dayFlows[dateStr] = &dayFlow{}
		}
		dayFlows[dateStr].outflow += dailyBudget
		current = current.AddDate(0, 0, 1)
	}

	// 8. 逐日累积计算
	var points []DailyForecastPoint
	currentBalance := startingBalance
	current = startDate
	var dangerDates []string
	var warningDates []string

	for current.Before(endDate) {
		dateStr := current.Format("2006-01-02")
		flow := dayFlows[dateStr]
		if flow == nil {
			flow = &dayFlow{}
		}

		openingBalance := currentBalance
		inflow := flow.inflow
		outflow := flow.outflow
		closingBalance := openingBalance + inflow - outflow

		point := DailyForecastPoint{
			Date:           dateStr,
			OpeningBalance: openingBalance,
			Inflow:         inflow,
			Outflow:        outflow,
			ClosingBalance: closingBalance,
		}
		points = append(points, point)

		// 检查状态
		if closingBalance < 0 {
			dangerDates = append(dangerDates, dateStr)
		} else if closingBalance < req.SafeAmount {
			warningDates = append(warningDates, dateStr)
		}

		currentBalance = closingBalance
		current = current.AddDate(0, 0, 1)
	}

	// 9. 确定整体状态
	var status ForecastStatus
	if len(dangerDates) > 0 {
		status = ForecastStatusDanger
	} else if len(warningDates) > 0 {
		status = ForecastStatusWarning
	} else {
		status = ForecastStatusSafe
	}

	return CashflowForecastResponse{
		StartDate:    req.StartDate,
		EndDate:      endDate.Format("2006-01-02"),
		Base:         req.Base,
		AccountID:    req.AccountID,
		SafeAmount:   req.SafeAmount,
		Points:       points,
		Status:       status,
		DangerDates:  dangerDates,
		WarningDates: warningDates,
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
			return dates // 未知频率，停止
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
