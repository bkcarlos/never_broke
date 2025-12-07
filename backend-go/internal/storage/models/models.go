package models

// TodayStats 今日统计
type TodayStats struct {
	TodayExpense   float64 `json:"todayExpense"`
	TodayAvailable float64 `json:"todayAvailable"`
	MonthUsed      float64 `json:"monthUsed"`
	MonthBudget    float64 `json:"monthBudget"`
	MonthNet       float64 `json:"monthNet"`
}

// AssetSummary 资产总览
type AssetSummary struct {
	TotalAssets      float64 `json:"totalAssets"`
	TotalLiabilities float64 `json:"totalLiabilities"`
	NetAssets        float64 `json:"netAssets"`
}

// CategoryStat 分类统计
type CategoryStat struct {
	CategoryKey string  `json:"categoryKey"`
	Amount      float64 `json:"amount"`
	Percent     float64 `json:"percent"`
	Rank        int     `json:"rank"`
}

// HomeOverview 首页概览
type HomeOverview struct {
	TodayStats    TodayStats     `json:"todayStats"`
	AssetSummary  AssetSummary   `json:"assetSummary"`
	TopCategories []CategoryStat `json:"topCategories"`
}

// DailyTransaction 每日账单
type DailyTransaction struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"` // income, expense, transfer
	Amount      float64 `json:"amount"`
	CategoryKey string  `json:"categoryKey"`
	Remark      string  `json:"remark,omitempty"`
	Time        string  `json:"time"` // HH:mm
}

// DailyStats 每日统计
type DailyStats struct {
	TotalIncome  float64 `json:"totalIncome"`
	TotalExpense float64 `json:"totalExpense"`
}

// DailyTransactionResponse 每日账单响应
type DailyTransactionResponse struct {
	Date  string             `json:"date"` // YYYY-MM-DD
	Stats DailyStats         `json:"stats"`
	List  []DailyTransaction `json:"list"`
}

// CategoryReportItem 分类报表项
type CategoryReportItem struct {
	CategoryKey string  `json:"categoryKey"`
	Amount      float64 `json:"amount"`
	Percent     float64 `json:"percent"`
	Rank        int     `json:"rank"`
}

// DailyTrendPoint 每日趋势点
type DailyTrendPoint struct {
	Date    string  `json:"date"` // YYYY-MM-DD
	Expense float64 `json:"expense"`
	Income  float64 `json:"income"`
}

// MonthlyReport 月度报表
type MonthlyReport struct {
	Month         string               `json:"month"` // YYYY-MM
	TotalIncome   float64              `json:"totalIncome"`
	TotalExpense  float64              `json:"totalExpense"`
	NetIncome     float64              `json:"netIncome"`
	CategoryStats []CategoryReportItem `json:"categoryStats"`
	DailyTrend    []DailyTrendPoint    `json:"dailyTrend"`
}

// UserSettings 用户设置
type UserSettings struct {
	Locale     string `json:"locale"`
	ShowAmount bool   `json:"showAmount"`
}

