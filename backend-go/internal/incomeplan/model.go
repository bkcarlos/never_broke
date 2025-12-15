package incomeplan

// IncomeFrequency 收入频率
type IncomeFrequency string

const (
	FrequencyMonthly  IncomeFrequency = "monthly"
	FrequencyWeekly   IncomeFrequency = "weekly"
	FrequencyBiweekly IncomeFrequency = "biweekly"
	FrequencyYearly   IncomeFrequency = "yearly"
)

// IncomePlan 收入计划
type IncomePlan struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`         // 名称：工资 / 副业 / 房租收入
	Amount       float64         `json:"amount"`       // 每次预计收入金额
	Frequency    IncomeFrequency `json:"frequency"`    // 发薪频率
	AccountID    string          `json:"accountId"`    // 到账账户
	FirstPayDate string          `json:"firstPayDate"` // 第一次发薪日期 YYYY-MM-DD
	EndDate      string          `json:"endDate"`      // 可选，到期日期，否则长期
	Active       bool            `json:"active"`
}

// CreateIncomePlanRequest 创建收入计划请求
type CreateIncomePlanRequest struct {
	Name         string          `json:"name"`
	Amount       float64         `json:"amount"`
	Frequency    IncomeFrequency `json:"frequency"`
	AccountID    string          `json:"accountId"`
	FirstPayDate string          `json:"firstPayDate"`
	EndDate      string          `json:"endDate"`
}



