package reports

import (
	"context"
	"fmt"
	"sort"
	"time"

	"neverbroke/internal/ledger"
	"neverbroke/internal/repository"
)

// BuildMonthlySummary 构建月度汇总报表
func BuildMonthlySummary(
	ctx context.Context,
	store repository.Store,
	userID string,
	year int,
	month int,
) (SummaryReportResponse, error) {
	// 1. 计算本月的起止时间：[year-month-01, nextMonth-01)
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// 2. 使用 Transactions().ListByUserBetween 拉取该时间段内所有交易
	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return SummaryReportResponse{}, fmt.Errorf("failed to list transactions: %w", err)
	}

	// 3. 遍历交易进行统计
	var totalIncome float64
	var totalExpense float64
	categoryMap := make(map[string]float64) // categoryKey -> amount
	dailyMap := make(map[string]struct {
		income  float64
		expense float64
	}) // date -> {income, expense}

	for _, tx := range transactions {
		date := tx.OccurredAt[:10] // 提取 YYYY-MM-DD

		switch tx.Type {
		case ledger.TransactionTypeIncome:
			totalIncome += tx.Amount
			if daily, ok := dailyMap[date]; ok {
				daily.income += tx.Amount
				dailyMap[date] = daily
			} else {
				dailyMap[date] = struct {
					income  float64
					expense float64
				}{income: tx.Amount, expense: 0}
			}
		case ledger.TransactionTypeExpense:
			totalExpense += tx.Amount
			// 按分类聚合
			if tx.CategoryKey != "" {
				categoryMap[tx.CategoryKey] += tx.Amount
			}
			// 按日期聚合
			if daily, ok := dailyMap[date]; ok {
				daily.expense += tx.Amount
				dailyMap[date] = daily
			} else {
				dailyMap[date] = struct {
					income  float64
					expense float64
				}{income: 0, expense: tx.Amount}
			}
		}
	}

	// 4. 计算 NetAmount
	netAmount := totalIncome - totalExpense

	// 5. 构建分类汇总列表
	var categories []CategorySummary
	for categoryKey, amount := range categoryMap {
		ratio := 0.0
		if totalExpense > 0 {
			ratio = amount / totalExpense
		}
		categories = append(categories, CategorySummary{
			CategoryKey: categoryKey,
			Amount:      amount,
			Ratio:       ratio,
		})
	}

	// 6. 按支出金额降序排序，取前 3 个作为 TopCategories
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Amount > categories[j].Amount
	})
	topCategories := categories
	if len(topCategories) > 3 {
		topCategories = topCategories[:3]
	}

	// 7. 构建趋势点列表（按日期排序）
	var trend []TrendPoint
	var dates []string
	for date := range dailyMap {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	for _, date := range dates {
		daily := dailyMap[date]
		trend = append(trend, TrendPoint{
			Date:    date,
			Income:  daily.income,
			Expense: daily.expense,
		})
	}

	// 8. 上月对比
	var compareWithPrevMonth *CompareSummary
	prevYear, prevMonth := getPrevMonth(year, month)
	prevTotalExpense, err := getPrevMonthExpense(ctx, store, userID, prevYear, prevMonth)
	if err == nil {
		expenseChange := totalExpense - prevTotalExpense
		expenseChangeRate := 0.0
		if prevTotalExpense > 0 {
			expenseChangeRate = expenseChange / prevTotalExpense
		} else if totalExpense > 0 {
			expenseChangeRate = 1.0 // 从0到有支出，增长100%
		}
		compareWithPrevMonth = &CompareSummary{
			PrevYear:          prevYear,
			PrevMonth:         prevMonth,
			ExpenseChange:     expenseChange,
			ExpenseChangeRate: expenseChangeRate,
		}
	}

	return SummaryReportResponse{
		Period:               PeriodMonthly,
		Year:                 year,
		Month:                month,
		TotalIncome:          totalIncome,
		TotalExpense:         totalExpense,
		NetAmount:            netAmount,
		Categories:           categories,
		Trend:                trend,
		TopCategories:        topCategories,
		CompareWithPrevMonth: compareWithPrevMonth,
	}, nil
}

// getPrevMonth 计算上一个月的年月
func getPrevMonth(year, month int) (int, int) {
	if month == 1 {
		return year - 1, 12
	}
	return year, month - 1
}

// getPrevMonthExpense 获取上一个月的总支出
func getPrevMonthExpense(ctx context.Context, store repository.Store, userID string, year, month int) (float64, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	transactions, err := store.Transactions().ListByUserBetween(ctx, userID, startDateStr, endDateStr)
	if err != nil {
		return 0, err
	}

	var totalExpense float64
	for _, tx := range transactions {
		if tx.Type == ledger.TransactionTypeExpense {
			totalExpense += tx.Amount
		}
	}
	return totalExpense, nil
}
