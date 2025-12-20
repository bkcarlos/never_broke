package postgres

import (
	"neverbroke/internal/repository"

	"gorm.io/gorm"
)

// PgStore Postgres 实现的 Store (使用 GORM)
type PgStore struct {
	db              *gorm.DB
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	budgetRepo      repository.BudgetRepository
}

// NewPgStore 创建新的 Postgres Store (使用 GORM)
func NewPgStore(db *gorm.DB) *PgStore {
	return &PgStore{
		db:              db,
		accountRepo:     NewPgAccountRepository(db),
		transactionRepo: NewPgTransactionRepository(db),
		budgetRepo:      NewPgBudgetRepository(db),
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

// Budgets 返回预算仓储
func (s *PgStore) Budgets() repository.BudgetRepository {
	return s.budgetRepo
}

// Close 关闭数据库连接
func (s *PgStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
