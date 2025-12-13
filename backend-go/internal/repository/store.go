package repository

import (
	"context"

	"neverbroke/internal/accounts"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
)

// Store 统一的数据存储接口
type Store interface {
	Accounts() AccountRepository
	Transactions() TransactionRepository
	Installments() InstallmentRepository
	RecurringIncomes() RecurringIncomeRepository
	Close() error
}

// AccountRepository 账户仓储接口
type AccountRepository interface {
	GetByID(ctx context.Context, userID, id string) (*accounts.Account, error)
	ListByUser(ctx context.Context, userID string) ([]*accounts.Account, error)
	Create(ctx context.Context, acc *accounts.Account) error
	Update(ctx context.Context, acc *accounts.Account) error
}

// TransactionRepository 交易仓储接口
type TransactionRepository interface {
	GetByID(ctx context.Context, userID, id string) (*ledger.Transaction, error)
	ListByUserBetween(ctx context.Context, userID, from, to string) ([]*ledger.Transaction, error)
	Create(ctx context.Context, tx *ledger.Transaction) error
	CreateBatch(ctx context.Context, txs []*ledger.Transaction) error
}

// InstallmentRepository 分期计划仓储接口
type InstallmentRepository interface {
	GetByID(ctx context.Context, userID, id string) (*installment.InstallmentPlan, error)
	ListByAccount(ctx context.Context, userID, accountID string) ([]*installment.InstallmentPlan, error)
	Create(ctx context.Context, plan *installment.InstallmentPlan) error
	Update(ctx context.Context, plan *installment.InstallmentPlan) error
}

// RecurringIncomeRepository 周期收入计划仓储接口
type RecurringIncomeRepository interface {
	GetByID(ctx context.Context, userID, id string) (*recurringincome.RecurringIncomePlan, error)
	ListByUser(ctx context.Context, userID string) ([]*recurringincome.RecurringIncomePlan, error)
	Create(ctx context.Context, plan *recurringincome.RecurringIncomePlan) error
	Update(ctx context.Context, plan *recurringincome.RecurringIncomePlan) error
}


