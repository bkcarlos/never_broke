package home

import (
	"context"
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 首页处理器
type Handler struct {
	store repository.Store
}

// NewHandler 创建首页处理器
func NewHandler(store repository.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// GetOverview 获取首页概览
func (h *Handler) GetOverview(c *gin.Context) {
	// 使用固定用户 ID
	userID := "demo"
	ctx := context.Background()

	logger.Debug("Fetching home overview", zap.String("userID", userID))

	// 从数据库获取真实数据
	overview, err := h.calculateOverview(ctx, userID)
	if err != nil {
		logger.Error("Failed to calculate home overview",
			zap.String("userID", userID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to fetch home overview")
		return
	}

	logger.Info("Home overview fetched successfully",
		zap.Int("categoryCount", len(overview.CategoryTop)),
		zap.Float64("freeToSpend", overview.PayCycle.FreeToSpend),
	)
	response.OK(c, overview)
}

// calculateOverview 计算首页概览数据
func (h *Handler) calculateOverview(ctx context.Context, userID string) (*HomeOverview, error) {
	now := time.Now()

	// 获取本月起止日期
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 获取今日起止日期
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayEnd := todayStart.AddDate(0, 0, 1).Add(-time.Second)

	// 1. 获取所有账户，计算资产总览
	accounts, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	assetSummary := calculateAssetSummary(accounts)

	// 2. 获取本月交易记录
	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	// 3. 获取今日交易记录
	todayTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		todayStart.Format(time.RFC3339),
		todayEnd.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	// 4. 计算统计数据
	var todayExpense, todayIncome, monthExpense, monthIncome float64
	categoryMap := make(map[string]float64)

	// 计算今日数据
	for _, tx := range todayTxs {
		switch tx.Type {
		case "expense":
			todayExpense += tx.Amount
		case "income":
			todayIncome += tx.Amount
		}
	}

	// 计算本月数据和分类统计
	for _, tx := range monthTxs {
		switch tx.Type {
		case "expense":
			monthExpense += tx.Amount
			if tx.CategoryKey != "" {
				categoryMap[tx.CategoryKey] += tx.Amount
			}
		case "income":
			monthIncome += tx.Amount
		}
	}

	// 默认预算
	monthBudget := 3000.0

	todayStats := TodayStats{
		TodayExpense:   todayExpense,
		TodayAvailable: monthBudget - monthExpense,
		MonthUsed:      monthExpense,
		MonthBudget:    monthBudget,
		MonthNet:       monthIncome - monthExpense,
	}

	// 计算分类排行
	var categoryTop []CategoryStat
	for key, amount := range categoryMap {
		percent := 0.0
		if monthExpense > 0 {
			percent = (amount / monthExpense) * 100
		}
		categoryTop = append(categoryTop, CategoryStat{
			CategoryKey: key,
			Amount:      amount,
			Percent:     percent,
		})
	}

	// 按金额排序
	for i := 0; i < len(categoryTop); i++ {
		for j := i + 1; j < len(categoryTop); j++ {
			if categoryTop[j].Amount > categoryTop[i].Amount {
				categoryTop[i], categoryTop[j] = categoryTop[j], categoryTop[i]
			}
		}
	}

	// 设置排名，只返回前3名
	for i := range categoryTop {
		categoryTop[i].Rank = i + 1
	}
	if len(categoryTop) > 3 {
		categoryTop = categoryTop[:3]
	}

	// 5. 计算发薪周期（简化版本，暂时使用固定值）
	payCycle := PayCycleOverview{
		CycleName:        now.Format("1月") + "发薪周期",
		DaysToNextPayday: calculateDaysToNextPayday(now),
		FreeToSpend:      assetSummary.NetAssets,
		CycleUsedPercent: 0,
	}

	return &HomeOverview{
		TodayStats:   todayStats,
		PayCycle:     payCycle,
		AssetSummary: assetSummary,
		CategoryTop:  categoryTop,
		Hint: HomeHint{
			Type:    HintInfo,
			TextKey: "home.hint.default",
		},
	}, nil
}

// calculateAssetSummary 计算资产总览
func calculateAssetSummary(accs []*accounts.Account) AssetSummary {
	var totalAssets, totalLiabilities float64

	for _, acc := range accs {
		if acc.Archived {
			continue
		}

		switch acc.Type {
		case accounts.AccountTypeCash, accounts.AccountTypeBank, accounts.AccountTypeWallet:
			totalAssets += acc.Balance
		case accounts.AccountTypeCredit:
			// 信用卡余额为负数表示欠款
			if acc.Balance < 0 {
				totalLiabilities += -acc.Balance
			} else {
				totalAssets += acc.Balance
			}
		}
	}

	return AssetSummary{
		TotalAssets:      totalAssets,
		TotalLiabilities: totalLiabilities,
		NetAssets:        totalAssets - totalLiabilities,
	}
}

// calculateDaysToNextPayday 计算距离下次发薪日的天数
func calculateDaysToNextPayday(now time.Time) int {
	// 假设每月15号发薪
	payDay := 15
	currentDay := now.Day()

	if currentDay < payDay {
		return payDay - currentDay
	}

	// 下个月的发薪日
	nextMonth := now.AddDate(0, 1, 0)
	nextPayday := time.Date(nextMonth.Year(), nextMonth.Month(), payDay, 0, 0, 0, 0, time.Local)
	return int(nextPayday.Sub(now).Hours() / 24)
}
