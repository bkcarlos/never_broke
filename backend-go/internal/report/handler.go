package report

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 报表处理器
type Handler struct{}

// NewHandler 创建报表处理器
func NewHandler() *Handler {
	return &Handler{}
}

// GetMonthlyReport 获取月度报表
func (h *Handler) GetMonthlyReport(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		logger.Warn("Missing month parameter in monthly report request")
		response.BadRequest(c, "month parameter is required")
		return
	}

	logger.Debug("Fetching monthly report", zap.String("month", month))

	// Mock 数据
	report := models.MonthlyReport{
		Month:        month,
		TotalIncome:  15000.00,
		TotalExpense: 8543.22,
		NetIncome:    6456.78,
		CategoryStats: []models.CategoryReportItem{
			{CategoryKey: "food", Amount: 2345.67, Percent: 27.4, Rank: 1},
			{CategoryKey: "housing", Amount: 2000.00, Percent: 23.4, Rank: 2},
			{CategoryKey: "transport", Amount: 1234.56, Percent: 14.5, Rank: 3},
			{CategoryKey: "shopping", Amount: 987.65, Percent: 11.6, Rank: 4},
			{CategoryKey: "entertainment", Amount: 876.54, Percent: 10.3, Rank: 5},
			{CategoryKey: "healthcare", Amount: 654.32, Percent: 7.7, Rank: 6},
			{CategoryKey: "other", Amount: 444.48, Percent: 5.1, Rank: 7},
		},
		DailyTrend: []models.DailyTrendPoint{
			{Date: month + "-01", Expense: 234.56, Income: 500.00},
			{Date: month + "-02", Expense: 345.67, Income: 0.00},
			{Date: month + "-03", Expense: 123.45, Income: 0.00},
			{Date: month + "-04", Expense: 456.78, Income: 0.00},
			{Date: month + "-05", Expense: 234.56, Income: 0.00},
			{Date: month + "-06", Expense: 567.89, Income: 0.00},
			{Date: month + "-07", Expense: 345.67, Income: 0.00},
			{Date: month + "-08", Expense: 234.56, Income: 0.00},
			{Date: month + "-09", Expense: 456.78, Income: 0.00},
			{Date: month + "-10", Expense: 678.90, Income: 14500.00},
		},
	}

	logger.Info("Monthly report fetched successfully",
		zap.String("month", month),
		zap.Int("categoryCount", len(report.CategoryStats)),
		zap.Int("trendPoints", len(report.DailyTrend)),
	)
	response.OK(c, report)
}

