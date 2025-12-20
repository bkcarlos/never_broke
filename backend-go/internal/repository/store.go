package repository

import (
	"context"

	"neverbroke/internal/accounts"
	"neverbroke/internal/budget"
	"neverbroke/internal/ledger"
)

// Store 统一的数据存储接口
type Store interface {
	Accounts() AccountRepository
	Transactions() TransactionRepository
	Budgets() BudgetRepository
	Close() error
}

// AccountRepository 账户仓储接口
type AccountRepository interface {
	GetByID(ctx context.Context, userID, id string) (*accounts.Account, error)
	ListByUser(ctx context.Context, userID string) ([]*accounts.Account, error)
	Create(ctx context.Context, userID string, acc *accounts.Account) error
	Update(ctx context.Context, userID string, acc *accounts.Account) error
}

// TransactionRepository 交易仓储接口
type TransactionRepository interface {
	GetByID(ctx context.Context, userID, id string) (*ledger.Transaction, error)
	ListByUserBetween(ctx context.Context, userID, from, to string) ([]*ledger.Transaction, error)
	Create(ctx context.Context, userID string, tx *ledger.Transaction) error
	CreateBatch(ctx context.Context, userID string, txs []*ledger.Transaction) error
}

// BudgetRepository 预算仓储接口
type BudgetRepository interface {
	GetByMonth(ctx context.Context, userID, month string) (*budget.MonthlyBudget, error)
	Upsert(ctx context.Context, userID string, b *budget.MonthlyBudget) error
}
