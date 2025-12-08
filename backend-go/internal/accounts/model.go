package accounts

// AccountType 账户类型
type AccountType string

const (
	AccountTypeCash   AccountType = "cash"
	AccountTypeBank   AccountType = "bank"
	AccountTypeWallet AccountType = "wallet"
	AccountTypeCredit AccountType = "credit"
)

// Account 账户模型
type Account struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     AccountType `json:"type"`
	Currency string      `json:"currency"`
	Balance  float64     `json:"balance"`
	Archived bool        `json:"archived"`
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	Name     string      `json:"name"`
	Type     AccountType `json:"type"`
	Currency string      `json:"currency"`
	Balance  float64     `json:"balance"`
}
