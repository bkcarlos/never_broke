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
}
