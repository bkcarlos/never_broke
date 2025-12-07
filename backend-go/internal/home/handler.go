package home

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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

	// TODO: 未来可以从 userId、数据库中计算；现在用 mock 数据
	overview := HomeOverview{
		TodayStats: TodayStats{
			TodayExpense:   126.0,
			TodayAvailable: 230.0,
			MonthUsed:      2340.0,
			MonthBudget:    3000.0,
			MonthNet:       660.0,
		},
		PayCycle: PayCycleOverview{
			CycleName:        "12月发薪周期",
			DaysToNextPayday: 12,
			FreeToSpend:      3420.0,
			CycleUsedPercent: 38.5,
		},
		AssetSummary: AssetSummary{
			TotalAssets:      12380.0,
			TotalLiabilities: 3000.0,
			NetAssets:        9380.0,
		},
		CategoryTop: []CategoryStat{
			{CategoryKey: "FOOD_DINING", Amount: 860.0, Percent: 36.0, Rank: 1},
			{CategoryKey: "RENT", Amount: 1800.0, Percent: 30.0, Rank: 2},
			{CategoryKey: "TRANSPORT", Amount: 320.0, Percent: 12.0, Rank: 3},
		},
		Hint: HomeHint{
			Type:    HintInfo,
			TextKey: "home.hint.default",
		},
	}

	logger.Info("Home overview fetched successfully",
		zap.Int("categoryCount", len(overview.CategoryTop)),
		zap.Float64("freeToSpend", overview.PayCycle.FreeToSpend),
	)
	response.OK(c, overview)
}
