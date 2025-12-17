package report

import (
	"context"
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/repository"
	"neverbroke/internal/storage/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 报表处理器
type Handler struct {
	store repository.Store
}

// NewHandler 创建报表处理器
func NewHandler(store repository.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// GetMonthlyReport 获取月度报表
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		logger.Warn("Missing month parameter in monthly report request")
		response.BadRequest(c, "month parameter is required")
		return
	}

	// 使用固定用户 ID
	userID := "demo"
	ctx := context.Background()

	logger.Debug("Fetching monthly report",
		zap.String("month", month),
		zap.String("userID", userID))

	// 解析月份 (YYYY-MM)
	monthStart, err := time.Parse("2006-01", month)
	if err != nil {
		logger.Warn("Invalid month format", zap.String("month", month), zap.Error(err))
		response.BadRequest(c, "invalid month format, expected YYYY-MM")
		return
	}
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 从数据库获取本月交易
	txs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch monthly transactions",
			zap.String("month", month),
			zap.Error(err))
		response.InternalError(c, "Failed to fetch transactions")
		return
	}

	// 计算统计数据
	var totalIncome, totalExpense float64
	categoryMap := make(map[string]float64)
	dailyMap := make(map[string]*models.DailyTrendPoint)

	for _, tx := range txs {
		// 统计收支
		switch tx.Type {
		case "income":
			totalIncome += tx.Amount
		case "expense":
			totalExpense += tx.Amount
			// 统计分类
			if tx.CategoryKey != "" {
				categoryMap[tx.CategoryKey] += tx.Amount
			}
		}

		// 统计每日趋势
		occurredTime, _ := time.Parse(time.RFC3339, tx.OccurredAt)
		dateKey := occurredTime.Format("2006-01-02")
		if dailyMap[dateKey] == nil {
			dailyMap[dateKey] = &models.DailyTrendPoint{
				Date:    dateKey,
				Expense: 0,
				Income:  0,
			}
		}
		switch tx.Type {
		case "income":
			dailyMap[dateKey].Income += tx.Amount
		case "expense":
			dailyMap[dateKey].Expense += tx.Amount
		}
	}

	// 转换分类统计
	var categoryStats []models.CategoryReportItem
	for key, amount := range categoryMap {
		percent := 0.0
		if totalExpense > 0 {
			percent = (amount / totalExpense) * 100
		}
		categoryStats = append(categoryStats, models.CategoryReportItem{
			CategoryKey: key,
			Amount:      amount,
			Percent:     percent,
		})
	}

	// 按金额排序
	for i := 0; i < len(categoryStats); i++ {
		for j := i + 1; j < len(categoryStats); j++ {
			if categoryStats[j].Amount > categoryStats[i].Amount {
				categoryStats[i], categoryStats[j] = categoryStats[j], categoryStats[i]
			}
		}
	}

	// 设置排名
	for i := range categoryStats {
		categoryStats[i].Rank = i + 1
	}

	// 转换每日趋势
	var dailyTrend []models.DailyTrendPoint
	for _, point := range dailyMap {
		dailyTrend = append(dailyTrend, *point)
	}

	// 按日期排序
	for i := 0; i < len(dailyTrend); i++ {
		for j := i + 1; j < len(dailyTrend); j++ {
			if dailyTrend[j].Date < dailyTrend[i].Date {
				dailyTrend[i], dailyTrend[j] = dailyTrend[j], dailyTrend[i]
			}
		}
	}

	report := models.MonthlyReport{
		Month:         month,
		TotalIncome:   totalIncome,
		TotalExpense:  totalExpense,
		NetIncome:     totalIncome - totalExpense,
		CategoryStats: categoryStats,
		DailyTrend:    dailyTrend,
	}

	logger.Info("Monthly report fetched successfully",
		zap.String("month", month),
		zap.Int("categoryCount", len(report.CategoryStats)),
		zap.Int("trendPoints", len(report.DailyTrend)),
	)
	response.OK(c, report)
}
