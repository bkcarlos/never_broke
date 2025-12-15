package postgres

import (
	"database/sql"
	"neverbroke/internal/repository"
)

// PgStore Postgres 实现的 Store
type PgStore struct {
	db                  *sql.DB
	accountRepo         repository.AccountRepository
	transactionRepo     repository.TransactionRepository
	installmentRepo     repository.InstallmentRepository
	recurringIncomeRepo repository.RecurringIncomeRepository
}

// NewPgStore 创建新的 Postgres Store
func NewPgStore(db *sql.DB) *PgStore {
	return &PgStore{
		db:                  db,
		accountRepo:         NewPgAccountRepository(db),
		transactionRepo:     NewPgTransactionRepository(db),
		installmentRepo:     NewPgInstallmentRepository(db),
		recurringIncomeRepo: NewPgRecurringIncomeRepository(db),
	}
}

// Accounts 返回账户仓储
func (s *PgStore) Accounts() repository.AccountRepository {
	return s.accountRepo
}

// Transactions 返回交易仓储
func (s *PgStore) Transactions() repository.TransactionRepository {
	return s.transactionRepo
}

// Installments 返回分期计划仓储
func (s *PgStore) Installments() repository.InstallmentRepository {
	return s.installmentRepo
}

// RecurringIncomes 返回周期收入计划仓储
func (s *PgStore) RecurringIncomes() repository.RecurringIncomeRepository {
	return s.recurringIncomeRepo
}

// Close 关闭数据库连接
func (s *PgStore) Close() error {
	return s.db.Close()
}



