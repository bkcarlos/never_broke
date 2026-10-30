package cashflow

import (
	"context"
	"fmt"
	"time"

	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	"neverbroke/internal/money"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
	"neverbroke/internal/workspace"

	"gorm.io/gorm"
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
			if acc.Currency != "" && acc.Currency != money.Currency(ctx) {
				continue
			}
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
		if acc == nil || (acc.Currency != "" && acc.Currency != money.Currency(ctx)) {
			return CashflowForecastResponse{}, fmt.Errorf("account currency does not match forecast currency")
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

	// 4. 这里简化处理：不再获取周期收入和分期计划
	// 仅基于账户余额和预算计算

	// 5. 获取预算信息（用于日常消费预估）
	now := time.Now()
	var monthlyBudget float64 = 3000.0 // 默认预算
	budgetData, err := store.Budgets().GetByMonth(ctx, userID, now.Format("2006-01"))
	if err == nil && budgetData != nil && budgetData.TotalBudget > 0 {
		monthlyBudget = budgetData.TotalBudget
	}

	// 计算本月已消费金额
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
	dailyBudget := (monthlyBudget - monthExpense) / float64(remainingDaysInMonth)
	if dailyBudget < 0 {
		dailyBudget = 0
	}

	// 6. 构建日期到流入/流出的映射
	type dayFlow struct {
		inflow  float64
		outflow float64
	}
	dayFlows := make(map[string]*dayFlow)

	// 6.1 处理日常消费预估（在整个预测期内按每天均摊）
	current := startDate
	for current.Before(endDate) {
		dateStr := current.Format("2006-01-02")
		if dayFlows[dateStr] == nil {
			dayFlows[dateStr] = &dayFlow{}
		}
		dayFlows[dateStr].outflow += dailyBudget
		current = current.AddDate(0, 0, 1)
	}

	// 7. 逐日累积计算
	var points []DailyForecastPoint
	currentBalance := startingBalance
	current = startDate
	dangerDates := []string{}
	warningDates := []string{}

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
			OpeningBalance: &openingBalance,
			Inflow:         inflow,
			Outflow:        outflow,
			ClosingBalance: &closingBalance,
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

	// 8. 确定整体状态
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
		BalanceKnown: true, DailyBudget: dailyBudget,
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

// BuildForecastWithDB 构建现金流预测（使用db查询周期收入和分期）
func BuildForecastWithDB(ctx context.Context, store repository.Store, db *gorm.DB, userID string, req CashflowForecastRequest) (CashflowForecastResponse, error) {
	items, sourceAccounts, err := workspace.Load(db, userID)
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	if req.Base == ForecastBaseAccount {
		selected := []models.Account{}
		for _, a := range sourceAccounts {
			if a.ID == req.AccountID && a.Currency == money.Currency(ctx) && a.Type != "credit" {
				selected = append(selected, a)
			}
		}
		if len(selected) != 1 {
			return CashflowForecastResponse{}, fmt.Errorf("请选择同币种的现金类账户；信用卡欠款不作为可用余额")
		}
		sourceAccounts = selected
	}
	var settings []workspace.Setting
	if err := db.Where("user_id = ?", userID).Find(&settings).Error; err != nil {
		return CashflowForecastResponse{}, err
	}
	projection := workspace.Project(items, sourceAccounts, settings, money.Currency(ctx), req.StartDate, req.Days)
	balanceKnown := projection.Starting != nil && projection.UnknownAmounts == 0
	var startingBalance float64
	if projection.Starting != nil {
		startingBalance = float64(*projection.Starting) / 100
	}

	// 2. 解析起始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {

		return CashflowForecastResponse{}, fmt.Errorf("invalid startDate: %w", err)
	}

	// 3. 计算结束日期
	endDate := startDate.AddDate(0, 0, req.Days)

	// 4. 获取周期收入计划
	var recurringPlans []models.RecurringIncomePlan
	if db != nil {
		if err := db.Where("user_id = ? AND active = ? AND currency = ?", userID, true, money.Currency(ctx)).Find(&recurringPlans).Error; err != nil {
			return CashflowForecastResponse{}, err
		}
	}

	// 6. 获取预算信息
	now := time.Now()
	var monthlyBudget float64
	budgetData, err := store.Budgets().GetByMonth(ctx, userID, now.Format("2006-01"))
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	if money.Currency(ctx) == "CNY" && budgetData != nil && budgetData.TotalBudget > 0 {
		monthlyBudget = budgetData.TotalBudget
	}

	// 计算本月已消费金额
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0)
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, monthStart.Format(time.RFC3339), monthEnd.Format(time.RFC3339))
	if err != nil {
		return CashflowForecastResponse{}, err
	}
	var monthExpense float64
	for _, tx := range transactions {
		if tx.Type == ledger.TransactionTypeExpense {
			monthExpense += tx.Amount
		}
	}

	lastDayOfMonth := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	remainingDaysInMonth := lastDayOfMonth - now.Day()
	if remainingDaysInMonth <= 0 {
		remainingDaysInMonth = 1
	}
	dailyBudget := (monthlyBudget - monthExpense) / float64(remainingDaysInMonth)
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
		projectedDates := projectRecurringIncomeFromDB(&plan, startDate, endDate)
		for _, date := range projectedDates {
			dateStr := date.Format("2006-01-02")
			if dayFlows[dateStr] == nil {
				dayFlows[dateStr] = &dayFlow{}
			}
			dayFlows[dateStr].inflow += plan.Amount
		}
	}

	// Include source bills and native installments exactly once, honoring review states.
	for _, p := range projection.Points {
		if dayFlows[p.Date] == nil {
			dayFlows[p.Date] = &dayFlow{}
		}
		dayFlows[p.Date].outflow += float64(p.Billed+p.Estimated) / 100
	}

	// 7.3 处理日常消费预估
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
	dangerDates := []string{}
	warningDates := []string{}

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

		points = append(points, DailyForecastPoint{
			Date:           dateStr,
			OpeningBalance: &openingBalance,
			Inflow:         inflow,
			Outflow:        outflow,
			ClosingBalance: &closingBalance,
		})

		if closingBalance < 0 {
			dangerDates = append(dangerDates, dateStr)
		} else if closingBalance < req.SafeAmount {
			warningDates = append(warningDates, dateStr)
		}

		currentBalance = closingBalance
		current = current.AddDate(0, 0, 1)
	}

	var status ForecastStatus
	if len(dangerDates) > 0 {
		status = ForecastStatusDanger
	} else if len(warningDates) > 0 {
		status = ForecastStatusWarning
	} else {
		status = ForecastStatusSafe
	}

	if !balanceKnown {
		status = ForecastStatus("unknown")
		dangerDates = []string{}
		warningDates = []string{}
		for i := range points {
			points[i].OpeningBalance = nil
			points[i].ClosingBalance = nil
		}
	}
	return CashflowForecastResponse{
		StartDate:    req.StartDate,
		BalanceKnown: balanceKnown, UnknownAmounts: projection.UnknownAmounts, DailyBudget: dailyBudget,
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

// projectRecurringIncomeFromDB 推算周期收入在未来日期范围内的收入日期
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

// calculateInstallmentScheduleFromDB 计算分期计划在指定日期范围内的还款计划
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
