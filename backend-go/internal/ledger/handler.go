package ledger

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 账本处理器
type Handler struct{}

// NewHandler 创建账本处理器
func NewHandler() *Handler {
	return &Handler{}
}

// GetDailyTransactions 获取每日账单
func (h *Handler) GetDailyTransactions(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		logger.Warn("Missing date parameter in daily transactions request")
		response.BadRequest(c, "date parameter is required")
		return
	}

	logger.Debug("Fetching daily transactions", zap.String("date", date))

	// Mock 数据
	dailyData := models.DailyTransactionResponse{
		Date: date,
		Stats: models.DailyStats{
			TotalIncome:  500.00,
			TotalExpense: 156.80,
		},
		List: []models.DailyTransaction{
			{
				ID:          "tx001",
				Type:        "expense",
				Amount:      45.50,
				CategoryKey: "food",
				Remark:      "午餐",
				Time:        "12:30",
			},
			{
				ID:          "tx002",
				Type:        "expense",
				Amount:      28.00,
				CategoryKey: "transport",
				Remark:      "地铁",
				Time:        "08:15",
			},
			{
				ID:          "tx003",
				Type:        "expense",
				Amount:      83.30,
				CategoryKey: "shopping",
				Remark:      "日用品",
				Time:        "19:45",
			},
			{
				ID:          "tx004",
				Type:        "income",
				Amount:      500.00,
				CategoryKey: "salary",
				Remark:      "工资",
				Time:        "10:00",
			},
		},
	}

	logger.Info("Daily transactions fetched successfully",
		zap.String("date", date),
		zap.Int("count", len(dailyData.List)),
	)
	response.OK(c, dailyData)
}

