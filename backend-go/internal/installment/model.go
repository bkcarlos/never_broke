package installment

// InstallmentPlanStatus 分期计划状态
type InstallmentPlanStatus string

const (
	InstallmentStatusActive   InstallmentPlanStatus = "active"
	InstallmentStatusFinished InstallmentPlanStatus = "finished"
	InstallmentStatusCanceled InstallmentPlanStatus = "canceled"
)

// InstallmentPlan 分期计划
type InstallmentPlan struct {
	ID            string                `json:"id"`
	AccountID     string                `json:"accountId"`     // 一般是信用卡账户
	TotalAmount   float64               `json:"totalAmount"`   // 分期总金额（本金+手续费）
	Principal     float64               `json:"principal"`     // 本金（可以等于 totalAmount）
	FeeTotal      float64               `json:"feeTotal"`      // 手续费总额
	TermMonths    int                   `json:"termMonths"`    // 分期期数
	MonthlyAmount float64               `json:"monthlyAmount"` // 每期金额（预计算）
	Currency      string                `json:"currency"`
	FirstDueDate  string                `json:"firstDueDate"` // 第一期还款日 YYYY-MM-DD
	Status        InstallmentPlanStatus `json:"status"`
	CreatedAt     string                `json:"createdAt"`
}

// CreateExpenseInstallmentRequest 创建支出分期请求
type CreateExpenseInstallmentRequest struct {
	AccountID    string  `json:"accountId"`    // 哪个账户（一般信用卡）
	TotalAmount  float64 `json:"totalAmount"`  // 分期总金额（本金+手续费）
	Principal    float64 `json:"principal"`    // 本金
	FeeTotal     float64 `json:"feeTotal"`     // 手续费总额，可以为0
	TermMonths   int     `json:"termMonths"`   // 分期期数
	Currency     string  `json:"currency"`     // 例如 CNY
	FirstDueDate string  `json:"firstDueDate"` // 第一期扣款日 YYYY-MM-DD

	// 用于生成每期明细的标签
	CategoryKey  string `json:"categoryKey,omitempty"` // 使用原支出分类
	Remark       string `json:"remark,omitempty"`      // 原支出备注
	PurchaseDate string `json:"purchaseDate"`          // 原消费发生日期 YYYY-MM-DD
}

// CreateExpenseInstallmentResponse 创建支出分期响应
// 注意：这里使用 ledger.Transaction，需要在 handler 中导入 ledger 包
type CreateExpenseInstallmentResponse struct {
	Plan         InstallmentPlan `json:"plan"`
	Transactions interface{}     `json:"transactions"` // 实际类型是 []ledger.Transaction，在 handler 中定义
}
