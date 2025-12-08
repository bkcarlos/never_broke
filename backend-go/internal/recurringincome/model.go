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
	FrequencyMonthly    RecurringIncomeFrequency = "monthly"    // 每月
	FrequencyBimonthly  RecurringIncomeFrequency = "bimonthly"  // 每两个月
	FrequencyQuarterly  RecurringIncomeFrequency = "quarterly"  // 每季度
	FrequencySemiannual RecurringIncomeFrequency = "semiannual" // 每半年
	FrequencyWeekly     RecurringIncomeFrequency = "weekly"     // 每周
	FrequencyYearly     RecurringIncomeFrequency = "yearly"     // 每年
)

// RecurringIncomePlan 周期收入计划（固定收入配置）
type RecurringIncomePlan struct {
	ID            string                   `json:"id"`
	Type          RecurringIncomeType      `json:"type"`        // salary/rent/side/other
	Label         string                   `json:"label"`       // 展示名称，如"工资""房租"
	AccountID     string                   `json:"accountId"`   // 收入到账账户
	Amount        float64                  `json:"amount"`      // 每次收入金额（固定），工资为税前
	Currency      string                   `json:"currency"`    // CNY
	Frequency     RecurringIncomeFrequency `json:"frequency"`   // monthly/weekly/...
	DayOfMonth    int                      `json:"dayOfMonth"`  // 月频率时使用：1-31
	Weekday       int                      `json:"weekday"`     // 周频率时使用：1=周一...7=周日
	NextDueDate   string                   `json:"nextDueDate"` // 下次应收日期 YYYY-MM-DD
	LastFiredDate string                   `json:"lastFiredDate,omitempty"`
	CityCode      string                   `json:"cityCode,omitempty"` // 薪资：纳税城市，如 "shanghai"
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
	NextDueDate string                   `json:"nextDueDate"`        // 初始下次应收日
	CityCode    string                   `json:"cityCode,omitempty"` // 薪资：纳税城市
	Active      bool                     `json:"active"`
}

// SalaryPreviewRequest 工资预览请求：从 startDate 到当年年底
type SalaryPreviewRequest struct {
	PreTaxAmount float64                  `json:"preTaxAmount"` // 每次发薪税前金额
	CityCode     string                   `json:"cityCode"`     // 纳税城市，如 "beijing", "shanghai"
	StartDate    string                   `json:"startDate"`    // 开始执行日期 YYYY-MM-DD
	PayDay       int                      `json:"payDay"`       // 发薪日（1-31）
	Frequency    RecurringIncomeFrequency `json:"frequency"`    // 常用为 monthly
	Currency     string                   `json:"currency"`     // CNY
}

// SalaryPreviewItem 单次发薪预览
type SalaryPreviewItem struct {
	Date             string  `json:"date"`             // 发薪日
	PreTax           float64 `json:"preTax"`           // 税前
	SocialSecurity   float64 `json:"socialSecurity"`   // 五险一金个人部分（简化）
	TaxableIncome    float64 `json:"taxableIncome"`    // 应纳税所得额
	TaxAmount        float64 `json:"taxAmount"`        // 个税
	PostTax          float64 `json:"postTax"`          // 实际到手
	CumulativePreTax float64 `json:"cumulativePreTax"` // 累计税前（方便对照中国个税累积制）
	CumulativeTax    float64 `json:"cumulativeTax"`    // 累计个税
}

// SalaryPreviewResponse 工资预览响应
type SalaryPreviewResponse struct {
	CityCode     string              `json:"cityCode"`
	Year         int                 `json:"year"`
	Items        []SalaryPreviewItem `json:"items"`
	TotalPreTax  float64             `json:"totalPreTax"`
	TotalTax     float64             `json:"totalTax"`
	TotalPostTax float64             `json:"totalPostTax"`
}

// ProjectedTransaction 未来明细投射交易
type ProjectedTransaction struct {
	Date      string  `json:"date"`
	AccountID string  `json:"accountId"`
	Amount    float64 `json:"amount"`
	Type      string  `json:"type"`   // "income"
	Label     string  `json:"label"`  // 来源，如"工资""房租收入"等
	PlanID    string  `json:"planId"` // 对应周期计划ID
	IsSalary  bool    `json:"isSalary"`
}
