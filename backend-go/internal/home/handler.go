package home

import (
	"context"
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/installment"
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

	// 获取分期计划用于计算负债
	installmentPlans, err := h.store.Installments().ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	assetSummary := h.calculateAssetSummary(ctx, accounts, installmentPlans)

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

	// 计算分类排行 - 初始化为空切片，确保 JSON 返回 [] 而不是 null
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
func (h *Handler) calculateAssetSummary(ctx context.Context, accs []*accounts.Account, installmentPlans []*installment.InstallmentPlan) AssetSummary {
	var totalAssets, totalLiabilities, creditAvailable float64

	// 1. 计算账户资产和信用卡负债
	for _, acc := range accs {
		if acc.Archived {
			continue
		}

		switch acc.Type {
		case accounts.AccountTypeCash, accounts.AccountTypeBank, accounts.AccountTypeWallet:
			// 总资产：现金、银行卡、电子钱包余额
			totalAssets += acc.Balance

		case accounts.AccountTypeCredit:
			// 信用卡：Balance 表示已用额度（正数）
			// 待还款 = 已用额度
			totalLiabilities += acc.Balance

			// 可用额度 = 总额度 - 已用额度（不含临时额度）
			if acc.CreditLimit > 0 {
				available := acc.CreditLimit - acc.Balance
				if available > 0 {
					creditAvailable += available
				}
			}
		}
	}

	// 2. 计算分期剩余待还金额（只统计 active 状态的分期）
	var installmentLiabilities float64
	for _, plan := range installmentPlans {
		if plan.Status == installment.InstallmentStatusActive {
			// 查询该分期已还款的金额
			paidAmount, err := h.calculateInstallmentPaid(ctx, plan.ID)
			if err != nil {
				logger.Error("Failed to calculate installment paid",
					zap.String("planID", plan.ID),
					zap.Error(err),
				)
				continue
			}

			// 剩余待还 = 总金额 - 已还金额
			remaining := plan.TotalAmount - paidAmount
			if remaining > 0 {
				installmentLiabilities += remaining
			}
		}
	}

	// 3. 总负债 = 信用卡负债 + 分期剩余待还
	totalLiabilities += installmentLiabilities

	// 4. 可用余额 = 总资产 + 信用卡可用额度
	availableBalance := totalAssets + creditAvailable

	return AssetSummary{
		TotalAssets:      totalAssets,
		AvailableBalance: availableBalance,
		TotalLiabilities: totalLiabilities,
		NetAssets:        totalAssets - totalLiabilities,
	}
}

// calculateInstallmentPaid 计算分期已还款金额
func (h *Handler) calculateInstallmentPaid(ctx context.Context, planID string) (float64, error) {
	// 查询所有与该分期关联的交易记录
	// 假设交易的 Remark 或其他字段包含分期ID，或者有专门的关联字段
	// 这里需要根据实际的数据模型来实现
	// 简化实现：如果交易有 InstallmentPlanID 字段，则累加金额

	// 由于 Transaction 模型中有 InstallmentPlanID，我们可以查询
	// 但这需要在 TransactionRepository 中添加按 InstallmentPlanID 查询的方法
	// 暂时返回 0，表示未实现
	return 0, nil
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

// GetAssetSummary 获取资产总览
func (h *Handler) GetAssetSummary(c *gin.Context) {
	userID := "demo"
	ctx := context.Background()

	logger.Debug("Fetching asset summary", zap.String("userID", userID))

	// 获取账户
	accounts, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch accounts", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch asset summary")
		return
	}

	// 获取分期计划
	installmentPlans, err := h.store.Installments().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch installments", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch asset summary")
		return
	}

	assetSummary := h.calculateAssetSummary(ctx, accounts, installmentPlans)
	logger.Info("Asset summary fetched successfully",
		zap.Float64("totalAssets", assetSummary.TotalAssets),
		zap.Float64("totalLiabilities", assetSummary.TotalLiabilities),
		zap.Int("installmentCount", len(installmentPlans)),
	)
	response.OK(c, assetSummary)
}

// GetBudgetStats 获取本月预算统计
func (h *Handler) GetBudgetStats(c *gin.Context) {
	userID := "demo"
	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching budget stats", zap.String("userID", userID))

	// 获取本月起止日期
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 获取本月交易记录
	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch budget stats")
		return
	}

	// 计算本月收支
	var monthExpense, monthIncome float64
	for _, tx := range monthTxs {
		switch tx.Type {
		case "expense":
			monthExpense += tx.Amount
		case "income":
			monthIncome += tx.Amount
		}
	}

	// 默认预算
	monthBudget := 3000.0

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
	userID := "demo"
	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching pay cycle", zap.String("userID", userID))

	// 获取账户信息
	accounts, err := h.store.Accounts().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch accounts", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch pay cycle")
		return
	}

	// 获取分期计划
	installmentPlans, err := h.store.Installments().ListByUser(ctx, userID)
	if err != nil {
		logger.Error("Failed to fetch installments", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch pay cycle")
		return
	}

	assetSummary := h.calculateAssetSummary(ctx, accounts, installmentPlans)

	payCycle := PayCycleOverview{
		CycleName:        now.Format("1月") + "发薪周期",
		DaysToNextPayday: calculateDaysToNextPayday(now),
		FreeToSpend:      assetSummary.NetAssets,
		CycleUsedPercent: 0,
	}

	logger.Info("Pay cycle fetched successfully", zap.Int("daysToNext", payCycle.DaysToNextPayday))
	response.OK(c, payCycle)
}

// GetCategoryStats 获取分类统计（本月消费分类）
func (h *Handler) GetCategoryStats(c *gin.Context) {
	userID := "demo"
	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching category stats", zap.String("userID", userID))

	// 获取本月起止日期
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 获取本月交易记录
	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch category stats")
		return
	}

	// 计算分类统计
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

	// 计算分类排行 - 初始化为空切片，确保 JSON 返回 [] 而不是 null
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

	logger.Info("Category stats fetched successfully", zap.Int("count", len(categoryTop)))
	response.OK(c, categoryTop)
}

// GetMonthOverview 获取本月概览
func (h *Handler) GetMonthOverview(c *gin.Context) {
	userID := "demo"
	ctx := context.Background()
	now := time.Now()

	logger.Debug("Fetching month overview", zap.String("userID", userID))

	// 获取本月起止日期
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Second)

	// 获取本月交易记录
	monthTxs, err := h.store.Transactions().ListByUserBetween(ctx, userID,
		monthStart.Format(time.RFC3339),
		monthEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch transactions", zap.String("userID", userID), zap.Error(err))
		response.InternalError(c, "Failed to fetch month overview")
		return
	}

	// 计算收支
	var monthIncome, monthExpense float64
	for _, tx := range monthTxs {
		switch tx.Type {
		case "income":
			monthIncome += tx.Amount
		case "expense":
			monthExpense += tx.Amount
		}
	}

	// 计算收支状态
	status := calculateStatus(monthIncome, monthExpense)

	// 生成建议
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

// calculateStatus 计算收支状态
func calculateStatus(income, expense float64) string {
	if income > expense {
		return "surplus" // 盈余
	} else if income < expense {
		return "deficit" // 赤字
	} else {
		return "balanced" // 平衡
	}
}

// generateSuggestion 生成消费建议
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
	} else {
		return "💰 储蓄率良好，继续保持"
	}
}
