package home

// TodayStats 今日统计
type TodayStats struct {
	TodayExpense   float64 `json:"todayExpense"`
	TodayAvailable float64 `json:"todayAvailable"`
	MonthUsed      float64 `json:"monthUsed"`
	MonthBudget    float64 `json:"monthBudget"`
	MonthNet       float64 `json:"monthNet"`
}

// PayCycleOverview 发薪周期概览
type PayCycleOverview struct {
	CycleName        string  `json:"cycleName"`
	DaysToNextPayday int     `json:"daysToNextPayday"`
	FreeToSpend      float64 `json:"freeToSpend"`
	CycleUsedPercent float64 `json:"cycleUsedPercent"`
}

// AssetSummary 资产总览
type AssetSummary struct {
	TotalAssets      float64 `json:"totalAssets"`      // 总资产（不含信用卡）
	AvailableBalance float64 `json:"availableBalance"` // 可用余额（总资产+信用卡可用额度）
	TotalLiabilities float64 `json:"totalLiabilities"` // 总负债（信用卡待还款）
	NetAssets        float64 `json:"netAssets"`        // 净资产（总资产-总负债）
}

// CategoryStat 分类统计
type CategoryStat struct {
	CategoryKey string  `json:"categoryKey"`
	Amount      float64 `json:"amount"`
	Percent     float64 `json:"percent"`
	Rank        int     `json:"rank"`
}

// HintType 提示类型
type HintType string

const (
	HintInfo    HintType = "info"
	HintWarning HintType = "warning"
	HintSuccess HintType = "success"
)

// HomeHint 首页提示
type HomeHint struct {
	Type    HintType `json:"type"`
	TextKey string   `json:"textKey"`
}

// HomeOverview 首页概览
type HomeOverview struct {
	TodayStats   TodayStats       `json:"todayStats"`
	PayCycle     PayCycleOverview `json:"payCycle"`
	AssetSummary AssetSummary     `json:"assetSummary"`
	CategoryTop  []CategoryStat   `json:"categoryTop"`
	Hint         HomeHint         `json:"hint"`
}
