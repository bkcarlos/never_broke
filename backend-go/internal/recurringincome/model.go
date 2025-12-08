package recurringincome

// RecurringIncomeType 周期收入类型
type RecurringIncomeType string

const (
	RecurringIncomeTypeSalary RecurringIncomeType = "salary" // 工资
	RecurringIncomeTypeRent   RecurringIncomeType = "rent"   // 房租收入
	RecurringIncomeTypeSide   RecurringIncomeType = "side"   // 固定副业
	RecurringIncomeTypeOther  RecurringIncomeType = "other"
)

// RecurringIncomeFrequency 周期收入频率
type RecurringIncomeFrequency string

const (
	FrequencyMonthly  RecurringIncomeFrequency = "monthly"
	FrequencyWeekly   RecurringIncomeFrequency = "weekly"
	FrequencyBiweekly RecurringIncomeFrequency = "biweekly"
	FrequencyYearly   RecurringIncomeFrequency = "yearly"
)

// RecurringIncomePlan 周期收入计划（固定收入配置）
type RecurringIncomePlan struct {
	ID            string                   `json:"id"`
	Type          RecurringIncomeType      `json:"type"`        // salary/rent/side/other
	Label         string                   `json:"label"`       // 展示名称，如"工资""房租"
	AccountID     string                   `json:"accountId"`   // 收入到账账户
	Amount        float64                  `json:"amount"`      // 每次收入金额（固定）
	Currency      string                   `json:"currency"`    // CNY
	Frequency     RecurringIncomeFrequency `json:"frequency"`   // monthly/weekly/...
	DayOfMonth    int                      `json:"dayOfMonth"`  // 月频率时使用：1-31
	Weekday       int                      `json:"weekday"`     // 周频率时使用：1=周一...7=周日
	NextDueDate   string                   `json:"nextDueDate"` // 下次应收日期 YYYY-MM-DD
	LastFiredDate string                   `json:"lastFiredDate,omitempty"`
	Active        bool                     `json:"active"`
	CreatedAt     string                   `json:"createdAt"`
}

// UpsertRecurringIncomeRequest 新建/更新请求
type UpsertRecurringIncomeRequest struct {
	ID          string                   `json:"id,omitempty"` // 为空则是新建
	Type        RecurringIncomeType      `json:"type"`
	Label       string                   `json:"label"`
	AccountID   string                   `json:"accountId"`
	Amount      float64                  `json:"amount"`
	Currency    string                   `json:"currency"`
	Frequency   RecurringIncomeFrequency `json:"frequency"`
	DayOfMonth  int                      `json:"dayOfMonth,omitempty"`
	Weekday     int                      `json:"weekday,omitempty"`
	NextDueDate string                   `json:"nextDueDate"` // 初始下次应收日
	Active      bool                     `json:"active"`
}
