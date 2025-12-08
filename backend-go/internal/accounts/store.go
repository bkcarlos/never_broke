package accounts

import "sync"

var (
	mu       sync.RWMutex
	accounts = []Account{
		{
			ID:          "wallet_wechat",
			Type:        AccountTypeWallet,
			Institution: AccountInstitutionWeChat,
			Currency:    "CNY",
			Balance:     1200,
			Archived:    false,
		},
		{
			ID:          "bank_icbc_1234",
			Type:        AccountTypeBank,
			Institution: AccountInstitutionICBC,
			CardLast4:   "1234",
			Currency:    "CNY",
			Balance:     5000,
			Archived:    false,
		},
		{
			ID:          "credit_cmb_7788",
			Type:        AccountTypeCredit,
			Institution: AccountInstitutionCMB,
			CardLast4:   "7788",
			Currency:    "CNY",
			Balance:     2000,
			CreditLimit: 20000,
			TempLimit:   3000,
			Archived:    false,
		},
	}
)

// ListAccounts 列出所有账户
func ListAccounts() []Account {
	mu.RLock()
	defer mu.RUnlock()
	copied := make([]Account, len(accounts))
	copy(copied, accounts)
	return copied
}

// FindAccountByID 根据 ID 查找账户
func FindAccountByID(id string) (*Account, bool) {
	mu.RLock()
	defer mu.RUnlock()
	for i := range accounts {
		if accounts[i].ID == id {
			acc := accounts[i]
			return &acc, true
		}
	}
	return nil, false
}

// AddToBalance 更新账户余额（增加或减少）
func AddToBalance(id string, delta float64) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range accounts {
		if accounts[i].ID == id {
			accounts[i].Balance += delta
			return true
		}
	}
	return false
}

// AddAccount 添加账户
func AddAccount(acc Account) Account {
	mu.Lock()
	defer mu.Unlock()
	accounts = append(accounts, acc)
	return acc
}

// UpdateAccount 更新账户（用于更新额度等）
func UpdateAccount(id string, updater func(*Account)) bool {
	mu.Lock()
	defer mu.Unlock()
	for i := range accounts {
		if accounts[i].ID == id {
			updater(&accounts[i])
			return true
		}
	}
	return false
}
