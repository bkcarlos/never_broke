package installment

// InstallmentPlanStatus 分期计划状态
type InstallmentPlanStatus string

const (
	InstallmentStatusActive   InstallmentPlanStatus = "active"
	InstallmentStatusFinished InstallmentPlanStatus = "finished"
	InstallmentStatusCanceled InstallmentPlanStatus = "canceled"
)

// InstallmentRepayMethod 还款方式
type InstallmentRepayMethod string

const (
	RepayMethodEqualPrincipal InstallmentRepayMethod = "equal_principal" // 等额本金
	RepayMethodEqualPayment   InstallmentRepayMethod = "equal_payment"   // 等额本息
)

// InstallmentPlan 分期计划
type InstallmentPlan struct {
	ID            string                 `json:"id"`
	AccountID     string                 `json:"accountId"`     // 一般是信用卡账户
	TotalAmount   float64                `json:"totalAmount"`   // 分期总金额（本金+利息）
	Principal     float64                `json:"principal"`     // 本金
	TotalInterest float64                `json:"totalInterest"` // 总利息
	TermMonths    int                    `json:"termMonths"`    // 分期期数
	MonthlyRate   float64                `json:"monthlyRate"`   // 月利率（小数）
	RepayMethod   InstallmentRepayMethod `json:"repayMethod"`   // 还款方式
	Currency      string                 `json:"currency"`
	FirstDueDate  string                 `json:"firstDueDate"` // 第一期还款日 YYYY-MM-DD
	Status        InstallmentPlanStatus  `json:"status"`
	CreatedAt     string                 `json:"createdAt"`
}

// PreviewInstallmentRequest 预览分期请求
type PreviewInstallmentRequest struct {
	Principal          float64                `json:"principal"`          // 分期本金
	TermMonths         int                    `json:"termMonths"`         // 分期期数
	MonthlyRatePercent float64                `json:"monthlyRatePercent"` // 每月利率（百分比）
	RepayMethod        InstallmentRepayMethod `json:"repayMethod"`        // equal_principal / equal_payment
	FirstDueDate       string                 `json:"firstDueDate"`       // 第一期还款日 YYYY-MM-DD
	Currency           string                 `json:"currency"`           // 例如 CNY
}

// InstallmentPreviewItem 分期预览单期明细
type InstallmentPreviewItem struct {
	Period             int     `json:"period"`
	DueDate            string  `json:"dueDate"`
	Principal          float64 `json:"principal"`
	Interest           float64 `json:"interest"`
	Total              float64 `json:"total"`
	RemainingPrincipal float64 `json:"remainingPrincipal"`
}

// PreviewInstallmentResponse 预览分期响应
type PreviewInstallmentResponse struct {
	Principal     float64                  `json:"principal"`
	TotalInterest float64                  `json:"totalInterest"`
	TotalAmount   float64                  `json:"totalAmount"`
	TermMonths    int                      `json:"termMonths"`
	MonthlyRate   float64                  `json:"monthlyRate"` // 小数
	RepayMethod   InstallmentRepayMethod   `json:"repayMethod"`
	Schedule      []InstallmentPreviewItem `json:"schedule"`
}

// CreateExpenseInstallmentRequest 创建支出分期请求
type CreateExpenseInstallmentRequest struct {
	AccountID          string                 `json:"accountId"`
	Principal          float64                `json:"principal"`
	TermMonths         int                    `json:"termMonths"`
	MonthlyRatePercent float64                `json:"monthlyRatePercent"`
	RepayMethod        InstallmentRepayMethod `json:"repayMethod"`
	FirstDueDate       string                 `json:"firstDueDate"`
	Currency           string                 `json:"currency"`

	CategoryKey  string `json:"categoryKey,omitempty"`
	Remark       string `json:"remark,omitempty"`
	PurchaseDate string `json:"purchaseDate"` // 原始消费日期
}

// CreateExpenseInstallmentResponse 创建支出分期响应
// 注意：这里使用 ledger.Transaction，需要在 handler 中导入 ledger 包
type CreateExpenseInstallmentResponse struct {
	Plan         InstallmentPlan `json:"plan"`
	Transactions interface{}     `json:"transactions"` // 实际类型是 []ledger.Transaction，在 handler 中定义
}
