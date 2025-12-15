package reports

// PeriodType 报表周期类型
type PeriodType string

const (
	PeriodMonthly PeriodType = "monthly"
	PeriodYearly  PeriodType = "yearly"
)

// CategorySummary 分类汇总
type CategorySummary struct {
	CategoryKey string  `json:"categoryKey"`
	Amount      float64 `json:"amount"`
	Ratio       float64 `json:"ratio"` // 在总支出中的占比（0~1）
}

// TrendPoint 趋势点
type TrendPoint struct {
	Date    string  `json:"date"` // YYYY-MM-DD 或 YYYY-MM
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}

// SummaryReportRequest 报表请求
type SummaryReportRequest struct {
	Period PeriodType `json:"period"` // monthly/yearly，MVP 用 monthly
	Year   int        `json:"year"`
	Month  int        `json:"month"` // 当 period=monthly 时有效
}

// CompareSummary 对比汇总
type CompareSummary struct {
	PrevYear          int     `json:"prevYear"`
	PrevMonth         int     `json:"prevMonth"`
	ExpenseChange     float64 `json:"expenseChange"`     // 本月支出 - 上月支出
	ExpenseChangeRate float64 `json:"expenseChangeRate"` // 相对上月变化比例
}

// SummaryReportResponse 报表响应
type SummaryReportResponse struct {
	Period               PeriodType        `json:"period"`
	Year                 int               `json:"year"`
	Month                int               `json:"month"`
	TotalIncome          float64           `json:"totalIncome"`
	TotalExpense         float64           `json:"totalExpense"`
	NetAmount            float64           `json:"netAmount"`
	Categories           []CategorySummary `json:"categories"`    // 只统计支出类
	Trend                []TrendPoint      `json:"trend"`         // 当月按天
	TopCategories        []CategorySummary `json:"topCategories"` // 前三大支出分类
	CompareWithPrevMonth *CompareSummary   `json:"compareWithPrevMonth,omitempty"`
}
