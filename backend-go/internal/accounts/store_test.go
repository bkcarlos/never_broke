package accounts

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAccounts(t *testing.T) {
	// Reset accounts to known state
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "acc1", Type: AccountTypeBank, Balance: 1000},
		{ID: "acc2", Type: AccountTypeCash, Balance: 500},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	result := ListAccounts()
	assert.Len(t, result, 2)
	assert.Equal(t, "acc1", result[0].ID)
	assert.Equal(t, "acc2", result[1].ID)
}

func TestFindAccountByID(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "test-id", Type: AccountTypeBank, Balance: 1000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	t.Run("found", func(t *testing.T) {
		acc, found := FindAccountByID("test-id")
		require.True(t, found)
		assert.Equal(t, "test-id", acc.ID)
		assert.Equal(t, 1000.0, acc.Balance)
	})

	t.Run("not found", func(t *testing.T) {
		_, found := FindAccountByID("non-existent")
		assert.False(t, found)
	})
}

func TestAddToBalance(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "test-id", Type: AccountTypeBank, Balance: 1000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	t.Run("add positive amount", func(t *testing.T) {
		ok := AddToBalance("test-id", 500)
		require.True(t, ok)

		acc, _ := FindAccountByID("test-id")
		assert.Equal(t, 1500.0, acc.Balance)
	})

	t.Run("add negative amount", func(t *testing.T) {
		ok := AddToBalance("test-id", -200)
		require.True(t, ok)

		acc, _ := FindAccountByID("test-id")
		assert.Equal(t, 1300.0, acc.Balance)
	})

	t.Run("account not found", func(t *testing.T) {
		ok := AddToBalance("non-existent", 100)
		assert.False(t, ok)
	})
}

func TestHasCashAccount(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	t.Run("has active cash account", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "cash1", Type: AccountTypeCash, Archived: false},
		}
		mu.Unlock()

		assert.True(t, HasCashAccount())
	})

	t.Run("has archived cash account only", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "cash1", Type: AccountTypeCash, Archived: true},
		}
		mu.Unlock()

		assert.False(t, HasCashAccount())
	})

	t.Run("no cash account", func(t *testing.T) {
		mu.Lock()
		accounts = []Account{
			{ID: "bank1", Type: AccountTypeBank, Archived: false},
		}
		mu.Unlock()

		assert.False(t, HasCashAccount())
	})
}

func TestAddAccount(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	newAcc := Account{
		ID:       "new-acc",
		Type:     AccountTypeBank,
		Balance:  2000,
		Currency: "CNY",
	}

	result := AddAccount(newAcc)
	assert.Equal(t, "new-acc", result.ID)

	// Verify it was added
	list := ListAccounts()
	assert.Len(t, list, 1)
	assert.Equal(t, "new-acc", list[0].ID)
}

func TestUpdateAccount(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "test-id", Type: AccountTypeCredit, CreditLimit: 10000, TempLimit: 0},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	t.Run("update existing account", func(t *testing.T) {
		ok := UpdateAccount("test-id", func(acc *Account) {
			acc.CreditLimit = 15000
			acc.TempLimit = 2000
		})
		require.True(t, ok)

		acc, _ := FindAccountByID("test-id")
		assert.Equal(t, 15000.0, acc.CreditLimit)
		assert.Equal(t, 2000.0, acc.TempLimit)
	})

	t.Run("update non-existent account", func(t *testing.T) {
		ok := UpdateAccount("non-existent", func(acc *Account) {
			acc.Balance = 999
		})
		assert.False(t, ok)
	})
}

func TestConcurrentAccess(t *testing.T) {
	mu.Lock()
	originalAccounts := accounts
	accounts = []Account{
		{ID: "concurrent-test", Type: AccountTypeBank, Balance: 1000},
	}
	mu.Unlock()

	defer func() {
		mu.Lock()
		accounts = originalAccounts
		mu.Unlock()
	}()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = FindAccountByID("concurrent-test")
			_ = ListAccounts()
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(amount float64) {
			defer wg.Done()
			AddToBalance("concurrent-test", amount)
		}(10.0)
	}

	wg.Wait()

	// Verify final balance
	acc, found := FindAccountByID("concurrent-test")
	require.True(t, found)
	assert.Equal(t, 2000.0, acc.Balance) // 1000 + (100 * 10)
}
