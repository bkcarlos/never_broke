package home

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/models"

	"github.com/gin-gonic/gin"
)

// Handler 首页处理器
type Handler struct{}

// NewHandler 创建首页处理器
func NewHandler() *Handler {
	return &Handler{}
}

// GetOverview 获取首页概览
func (h *Handler) GetOverview(c *gin.Context) {
	logger.Debug("Fetching home overview")

	// Mock 数据
	overview := models.HomeOverview{
		TodayStats: models.TodayStats{
			TodayExpense:   156.80,
			TodayAvailable: 343.20,
			MonthUsed:      3456.78,
			MonthBudget:    5000.00,
			MonthNet:       1543.22,
		},
		AssetSummary: models.AssetSummary{
			TotalAssets:      125000.00,
			TotalLiabilities: 35000.00,
			NetAssets:        90000.00,
		},
		TopCategories: []models.CategoryStat{
			{CategoryKey: "food", Amount: 1234.56, Percent: 35.7, Rank: 1},
			{CategoryKey: "transport", Amount: 876.54, Percent: 25.3, Rank: 2},
			{CategoryKey: "shopping", Amount: 654.32, Percent: 18.9, Rank: 3},
			{CategoryKey: "entertainment", Amount: 432.10, Percent: 12.5, Rank: 4},
			{CategoryKey: "other", Amount: 259.26, Percent: 7.6, Rank: 5},
		},
	}

	logger.Info("Home overview fetched successfully")
	response.OK(c, overview)
}
