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
	AccountInstitutionICBC   AccountInstitution = "icbc"   // 工商银行
	AccountInstitutionCCB    AccountInstitution = "ccb"    // 建设银行
	AccountInstitutionABC    AccountInstitution = "abc"    // 农业银行
	AccountInstitutionBOC    AccountInstitution = "boc"    // 中国银行
	AccountInstitutionBOCOM  AccountInstitution = "bocom"  // 交通银行
	AccountInstitutionCMB    AccountInstitution = "cmb"    // 招商银行
	AccountInstitutionSPDB   AccountInstitution = "spdb"   // 浦发银行
	AccountInstitutionCITIC  AccountInstitution = "citic"  // 中信银行
	AccountInstitutionCEB    AccountInstitution = "ceb"    // 光大银行
	AccountInstitutionHXB    AccountInstitution = "hxb"    // 华夏银行
	AccountInstitutionCMBC   AccountInstitution = "cmbc"   // 民生银行
	AccountInstitutionCGB    AccountInstitution = "cgb"    // 广发银行
	AccountInstitutionPAB    AccountInstitution = "pab"    // 平安银行
	AccountInstitutionCIB    AccountInstitution = "cib"    // 兴业银行
	AccountInstitutionPSBC   AccountInstitution = "psbc"   // 邮储银行
	AccountInstitutionWeChat AccountInstitution = "wechat" // 微信
	AccountInstitutionAlipay AccountInstitution = "alipay" // 支付宝
	AccountInstitutionOther  AccountInstitution = "other"  // 其他
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
