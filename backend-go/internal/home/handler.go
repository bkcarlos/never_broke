package home

import (
	"context"
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"
	"neverbroke/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 首页处理器
type Handler struct {
	store repository.Store
	db    *gorm.DB
}

// NewHandler 创建首页处理器
func NewHandler(store repository.Store, db *gorm.DB) *Handler {
	return &Handler{
		store: store,
		db:    db,
	}
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// GetOverview 获取首页概览
func (h *Handler) GetOverview(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()

	logger.Debug("Fetching home overview", zap.String("userID", userID))

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

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayEnd := todayStart.AddDate(0, 0, 1).Add(-time.Second)

	// 1. 获取所有账户
	accs, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 获取分期计划用于计算负债
	var installmentPlans []models.InstallmentPlan
	if h.db != nil {
		h.db.Where("user_id = ? AND status = ?", userID, "active").Find(&installmentPlans)
	}

	assetSummary := h.calculateAssetSummary(ctx, accs, installmentPlans)

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

	for _, tx := range todayTxs {
		switch tx.Type {
		case "expense":
			todayExpense += tx.Amount
		case "income":
			todayIncome += tx.Amount
		}
	}

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

	// 获取用户预算
	monthBudget := 3000.0
	budget, err := h.store.Budgets().GetByMonth(ctx, userID, now.Format("2006-01"))
	if err == nil && budget != nil && budget.TotalBudget > 0 {
		monthBudget = budget.TotalBudget
	}

	todayStats := TodayStats{
		TodayExpense:   todayExpense,
		TodayAvailable: monthBudget - monthExpense,
		MonthUsed:      monthExpense,
		MonthBudget:    monthBudget,
		MonthNet:       monthIncome - monthExpense,
	}

	categoryTop := make([]CategoryStat, 0)
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

	for i := 0; i < len(categoryTop); i++ {
		for j := i + 1; j < len(categoryTop); j++ {
			if categoryTop[j].Amount > categoryTop[i].Amount {
				categoryTop[i], categoryTop[j] = categoryTop[j], categoryTop[i]
			}
		}
	}

	for i := range categoryTop {
		categoryTop[i].Rank = i + 1
	}
	if len(categoryTop) > 3 {
		categoryTop = categoryTop[:3]
	}

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
func (h *Handler) calculateAssetSummary(ctx context.Context, accs []*accounts.Account, installmentPlans []models.InstallmentPlan) AssetSummary {
	var totalAssets, totalLiabilities, creditAvailable float64

	for _, acc := range accs {
		if acc.Archived {
			continue
		}

		switch acc.Type {
		case accounts.AccountTypeCash, accounts.AccountTypeBank, accounts.AccountTypeWallet:
			totalAssets += acc.Balance
		case accounts.AccountTypeCredit:
			totalLiabilities += acc.Balance
			if acc.CreditLimit > 0 {
				available := acc.CreditLimit - acc.Balance
				if available > 0 {
					creditAvailable += available
				}
			}
		}
	}

	var installmentLiabilities float64
	for _, plan := range installmentPlans {
		if plan.Status == "active" {
			remaining := plan.TotalAmount
			if remaining > 0 {
				installmentLiabilities += remaining
			}
		}
	}

	totalLiabilities += installmentLiabilities
	availableBalance := totalAssets + creditAvailable

	return AssetSummary{
		TotalAssets:      totalAssets,
		AvailableBalance: availableBalance,
		TotalLiabilities: totalLiabilities,
		NetAssets:        totalAssets - totalLiabilities,
	}
}

// calculateDaysToNextPayday 计算距离下次发薪日的天数
func calculateDaysToNextPayday(now time.Time) int {
	payDay := 15
	currentDay := now.Day()

	if currentDay < payDay {
		return payDay - currentDay
	}

	nextMonth := now.AddDate(0, 1, 0)
	nextPayday := time.Date(nextMonth.Year(), nextMonth.Month(), payDay, 0, 0, 0, 0, time.Local)
	return int(nextPayday.Sub(now).Hours() / 24)
}

// GetAssetSummary 获取资产总览
func (h *Handler) GetAssetSummary(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()

	logger.Debug("Fetching asset summary", zap.String("userID", userID))

	accs, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch accounts", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch asset summary")
		return
	}

	var installmentPlans []models.InstallmentPlan
	if h.db != nil {
		h.db.Where("user_id = ? AND status = ?", userID, "active").Find(&installmentPlans)
	}

	assetSummary := h.calculateAssetSummary(ctx, accs, installmentPlans)
	logger.Info("Asset summary fetched successfully",
		zap.Float64("totalAssets", assetSummary.TotalAssets),
		zap.Float64("totalLiabilities", assetSummary.TotalLiabilities),
	)
	response.OK(c, assetSummary)
}

// GetBudgetStats 获取本月预算统计
func (h *Handler) GetBudgetStats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching budget stats", zap.String("userID", userID))

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch budget stats")
		return
	}

	var monthExpense, monthIncome float64
	for _, tx := range monthTxs {
		switch tx.Type {
		case "expense":
			monthExpense += tx.Amount
		case "income":
			monthIncome += tx.Amount
		}
	}

	monthBudget := 3000.0
	budget, err := h.store.Budgets().GetByMonth(ctx, userID, now.Format("2006-01"))
	if err == nil && budget != nil && budget.TotalBudget > 0 {
		monthBudget = budget.TotalBudget
	}

	budgetStats := TodayStats{
		MonthUsed:   monthExpense,
		MonthBudget: monthBudget,
		MonthNet:    monthIncome - monthExpense,
	}

	logger.Info("Budget stats fetched successfully", zap.Float64("monthExpense", monthExpense))
	response.OK(c, budgetStats)
}

// GetPayCycle 获取支付周期信息
func (h *Handler) GetPayCycle(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching pay cycle", zap.String("userID", userID))

	accs, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch accounts", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch pay cycle")
		return
	}

	var installmentPlans []models.InstallmentPlan
	if h.db != nil {
		h.db.Where("user_id = ? AND status = ?", userID, "active").Find(&installmentPlans)
	}

	assetSummary := h.calculateAssetSummary(ctx, accs, installmentPlans)

	payCycle := PayCycleOverview{
		CycleName:        now.Format("1月") + "发薪周期",
		DaysToNextPayday: calculateDaysToNextPayday(now),
		FreeToSpend:      assetSummary.NetAssets,
		CycleUsedPercent: 0,
	}

	logger.Info("Pay cycle fetched successfully", zap.Int("daysToNext", payCycle.DaysToNextPayday))
	response.OK(c, payCycle)
}

// GetCategoryStats 获取分类统计
func (h *Handler) GetCategoryStats(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching category stats", zap.String("userID", userID))

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch category stats")
		return
	}

	var monthExpense float64
	categoryMap := make(map[string]float64)

	for _, tx := range monthTxs {
		if tx.Type == "expense" {
			monthExpense += tx.Amount
			if tx.CategoryKey != "" {
				categoryMap[tx.CategoryKey] += tx.Amount
			}
		}
	}

	categoryTop := make([]CategoryStat, 0)
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

	for i := 0; i < len(categoryTop); i++ {
		for j := i + 1; j < len(categoryTop); j++ {
			if categoryTop[j].Amount > categoryTop[i].Amount {
				categoryTop[i], categoryTop[j] = categoryTop[j], categoryTop[i]
			}
		}
	}

	for i := range categoryTop {
		categoryTop[i].Rank = i + 1
	}
	if len(categoryTop) > 3 {
		categoryTop = categoryTop[:3]
	}

	logger.Info("Category stats fetched successfully", zap.Int("count", len(categoryTop)))
	response.OK(c, categoryTop)
}

// GetMonthOverview 获取本月概览
func (h *Handler) GetMonthOverview(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching month overview", zap.String("userID", userID))

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch month overview")
		return
	}

	var monthIncome, monthExpense float64
	for _, tx := range monthTxs {
		switch tx.Type {
		case "income":
			monthIncome += tx.Amount
		case "expense":
			monthExpense += tx.Amount
		}
	}

	status := calculateStatus(monthIncome, monthExpense)
	suggestion := generateSuggestion(monthIncome, monthExpense)

	overview := MonthOverview{
		MonthIncome:  monthIncome,
		MonthExpense: monthExpense,
		Status:       status,
		Suggestion:   suggestion,
	}

	logger.Info("Month overview fetched successfully",
		zap.Float64("income", monthIncome),
		zap.Float64("expense", monthExpense),
		zap.String("status", status),
	)
	response.OK(c, overview)
}

func calculateStatus(income, expense float64) string {
	if income > expense {
		return "surplus"
	} else if income < expense {
		return "deficit"
	}
	return "balanced"
}

func generateSuggestion(income, expense float64) string {
	if income == 0 {
		return "💡 暂无收入记录"
	}

	ratio := expense / income

	if ratio > 0.9 {
		return "⚠️ 支出过高（超过90%），建议控制消费"
	} else if ratio > 0.7 {
		return "⚡ 支出较高（超过70%），注意理性消费"
	} else if ratio > 0.5 {
		return "✅ 收支健康，保持良好习惯"
	}
	return "💰 储蓄率良好，继续保持"
}
