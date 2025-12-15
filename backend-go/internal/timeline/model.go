package timeline

import "neverbroke/internal/ledger"

// TimelineMode 时间线模式
type TimelineMode string

const (
	TimelineModeHistory    TimelineMode = "history"
	TimelineModeWithFuture TimelineMode = "with_future"
)

// DailySummary 每日汇总
type DailySummary struct {
	Date      string  `json:"date"` // YYYY-MM-DD
	Income    float64 `json:"income"`
	Expense   float64 `json:"expense"`
	Net       float64 `json:"net"`
	HasFuture bool    `json:"hasFuture"` // 是否包含未来预计项目
}

// DailyTimelineResponse 每日时间线响应
type DailyTimelineResponse struct {
	FromDate string         `json:"fromDate"`
	ToDate   string         `json:"toDate"`
	Mode     TimelineMode   `json:"mode"`
	Days     []DailySummary `json:"days"`
}

// DailyDetailResponse 每日明细响应
type DailyDetailResponse struct {
	Date         string               `json:"date"`
	Mode         TimelineMode         `json:"mode"`
	Transactions []ledger.Transaction `json:"transactions"`
}
