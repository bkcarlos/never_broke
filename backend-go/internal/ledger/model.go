package ledger

// TransactionType 交易类型
type TransactionType string

const (
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeTransfer TransactionType = "transfer"
)

// CreateTransactionRequest 创建交易请求
type CreateTransactionRequest struct {
	Type          TransactionType `json:"type"`
	Amount        float64         `json:"amount"`
	CategoryKey   string          `json:"categoryKey,omitempty"`
	Remark        string          `json:"remark,omitempty"`
	OccurredAt    string          `json:"occurredAt"`
	AccountID     string          `json:"accountId,omitempty"`     // expense/income
	FromAccountID string          `json:"fromAccountId,omitempty"` // transfer
	ToAccountID   string          `json:"toAccountId,omitempty"`   // transfer
}

// Transaction 交易模型
type Transaction struct {
	ID string `json:"id"`
	CreateTransactionRequest

	// 新增：分期关联信息（如果是分期每一期的还款）
	InstallmentPlanID string `json:"installmentPlanId,omitempty"` // 所属分期计划
	InstallmentPeriod int    `json:"installmentPeriod,omitempty"` // 第几期（从1开始）
	InstallmentTotal  int    `json:"installmentTotal,omitempty"`  // 总期数
}
