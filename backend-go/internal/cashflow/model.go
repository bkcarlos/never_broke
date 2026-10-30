package cashflow

// ForecastBase 预测基准
type ForecastBase string

const (
	ForecastBaseTotalAssets ForecastBase = "total_assets" // 所有账户合计
	ForecastBaseAccount     ForecastBase = "account"      // 指定单个账户
)

// DailyForecastPoint 每日预测点
type DailyForecastPoint struct {
	Date           string   `json:"date"`           // YYYY-MM-DD
	OpeningBalance *float64 `json:"openingBalance"` // 期初余额
	Inflow         float64  `json:"inflow"`         // 当天流入（预计）
	Outflow        float64  `json:"outflow"`        // 当天流出（预计）
	ClosingBalance *float64 `json:"closingBalance"` // 期末余额
}

// ForecastStatus 预测状态
type ForecastStatus string

const (
	ForecastStatusSafe    ForecastStatus = "safe"
	ForecastStatusWarning ForecastStatus = "warning"
	ForecastStatusDanger  ForecastStatus = "danger"
)

// CashflowForecastRequest 现金流预测请求
type CashflowForecastRequest struct {
	StartDate  string       `json:"startDate"`  // YYYY-MM-DD
	Days       int          `json:"days"`       // 30/60/90
	Base       ForecastBase `json:"base"`       // total_assets / account
	AccountID  string       `json:"accountId"`  // Base 为 account 时必填
	SafeAmount float64      `json:"safeAmount"` // 用户设定的安全垫金额
}

// CashflowForecastResponse 现金流预测响应
type CashflowForecastResponse struct {
	BalanceKnown   bool                 `json:"balanceKnown"`
	UnknownAmounts int                  `json:"unknownAmounts"`
	DailyBudget    float64              `json:"dailyBudget"`
	StartDate      string               `json:"startDate"`
	EndDate        string               `json:"endDate"`
	Base           ForecastBase         `json:"base"`
	AccountID      string               `json:"accountId,omitempty"`
	SafeAmount     float64              `json:"safeAmount"`
	Points         []DailyForecastPoint `json:"points"`
	Status         ForecastStatus       `json:"status"`
	DangerDates    []string             `json:"dangerDates"`  // <0 的日期
	WarningDates   []string             `json:"warningDates"` // <safeAmount 且 >=0
}
