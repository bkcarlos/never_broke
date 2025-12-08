package accounts

// AccountType 账户类型
type AccountType string

const (
	AccountTypeCash   AccountType = "cash"
	AccountTypeBank   AccountType = "bank"
	AccountTypeWallet AccountType = "wallet"
	AccountTypeCredit AccountType = "credit"
)

// AccountInstitution 账户机构
type AccountInstitution string

const (
	AccountInstitutionICBC   AccountInstitution = "icbc"
	AccountInstitutionCCB    AccountInstitution = "ccb"
	AccountInstitutionABC    AccountInstitution = "abc"
	AccountInstitutionBOC    AccountInstitution = "boc"
	AccountInstitutionBOCOM  AccountInstitution = "bocom"
	AccountInstitutionCMB    AccountInstitution = "cmb"
	AccountInstitutionWeChat AccountInstitution = "wechat"
	AccountInstitutionAlipay AccountInstitution = "alipay"
	AccountInstitutionOther  AccountInstitution = "other"
)

// Account 账户模型
type Account struct {
	ID          string             `json:"id"`
	Type        AccountType        `json:"type"`                  // cash/bank/wallet/credit
	Institution AccountInstitution `json:"institution,omitempty"` // 工行/招行/微信/支付宝...
	CardLast4   string             `json:"cardLast4,omitempty"`   // 卡片尾号（最多4位）
	Currency    string             `json:"currency"`              // 例如 "CNY"
	Balance     float64            `json:"balance"`               // 对信用卡：代表已用额度
	CreditLimit float64            `json:"creditLimit,omitempty"` // 信用卡总额度
	TempLimit   float64            `json:"tempLimit,omitempty"`   // 信用卡临时额度
	Archived    bool               `json:"archived"`
}

// CreateAccountRequest 创建账户请求
type CreateAccountRequest struct {
	Type        AccountType        `json:"type"`
	Institution AccountInstitution `json:"institution,omitempty"`
	CardLast4   string             `json:"cardLast4,omitempty"`
	Currency    string             `json:"currency"`
	Balance     float64            `json:"balance"`
	CreditLimit float64            `json:"creditLimit,omitempty"`
	TempLimit   float64            `json:"tempLimit,omitempty"`
}

// UpdateCreditLimitRequest 更新信用卡额度请求
type UpdateCreditLimitRequest struct {
	CreditLimit *float64 `json:"creditLimit,omitempty"`
	TempLimit   *float64 `json:"tempLimit,omitempty"`
}
