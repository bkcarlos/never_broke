package memoryadapter

import (
	"context"
	"fmt"

	"neverbroke/internal/accounts"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/repository"
)

// MemoryStoreAdapter 内存存储适配器，实现 repository.Store 接口
type MemoryStoreAdapter struct{}

// NewMemoryStoreAdapter 创建内存存储适配器
func NewMemoryStoreAdapter() repository.Store {
	return &MemoryStoreAdapter{}
}

// Accounts 返回账户仓储
func (m *MemoryStoreAdapter) Accounts() repository.AccountRepository {
	return &memoryAccountRepository{}
}

// Transactions 返回交易仓储
func (m *MemoryStoreAdapter) Transactions() repository.TransactionRepository {
	return &memoryTransactionRepository{}
}

// Installments 返回分期计划仓储
func (m *MemoryStoreAdapter) Installments() repository.InstallmentRepository {
	return &memoryInstallmentRepository{}
}

// RecurringIncomes 返回周期收入计划仓储
func (m *MemoryStoreAdapter) RecurringIncomes() repository.RecurringIncomeRepository {
	return &memoryRecurringIncomeRepository{}
}

// Close 关闭（内存存储无需关闭）
func (m *MemoryStoreAdapter) Close() error {
	return nil
}

// memoryAccountRepository 内存账户仓储实现
type memoryAccountRepository struct{}

func (r *memoryAccountRepository) GetByID(ctx context.Context, userID, id string) (*accounts.Account, error) {
	acc, ok := accounts.FindAccountByID(id)
	if !ok {
		return nil, fmt.Errorf("account not found: %s", id)
	}
	return acc, nil
}

func (r *memoryAccountRepository) ListByUser(ctx context.Context, userID string) ([]*accounts.Account, error) {
	accs := accounts.ListAccounts()
	result := make([]*accounts.Account, len(accs))
	for i := range accs {
		result[i] = &accs[i]
	}
	return result, nil
}

func (r *memoryAccountRepository) Create(ctx context.Context, userID string, acc *accounts.Account) error {
	accounts.AddAccount(*acc)
	return nil
}

func (r *memoryAccountRepository) Update(ctx context.Context, userID string, acc *accounts.Account) error {
	// 简化实现：先删除再添加
	// 实际应该使用 UpdateAccount
	return nil
}

// memoryTransactionRepository 内存交易仓储实现
type memoryTransactionRepository struct{}

func (r *memoryTransactionRepository) GetByID(ctx context.Context, userID, id string) (*ledger.Transaction, error) {
	// 内存存储中没有实现，返回错误
	return nil, fmt.Errorf("not implemented")
}

func (r *memoryTransactionRepository) ListByUserBetween(ctx context.Context, userID, from, to string) ([]*ledger.Transaction, error) {
	// 内存存储中没有实现，返回空列表
	return []*ledger.Transaction{}, nil
}

func (r *memoryTransactionRepository) Create(ctx context.Context, userID string, tx *ledger.Transaction) error {
	// 内存存储中没有实现
	return nil
}

func (r *memoryTransactionRepository) CreateBatch(ctx context.Context, userID string, txs []*ledger.Transaction) error {
	// 内存存储中没有实现
	return nil
}

// memoryInstallmentRepository 内存分期计划仓储实现
type memoryInstallmentRepository struct{}

func (r *memoryInstallmentRepository) GetByID(ctx context.Context, userID, id string) (*installment.InstallmentPlan, error) {
	plan, ok := installment.FindPlanByID(id)
	if !ok {
		return nil, fmt.Errorf("installment plan not found: %s", id)
	}
	return plan, nil
}

func (r *memoryInstallmentRepository) ListByAccount(ctx context.Context, userID, accountID string) ([]*installment.InstallmentPlan, error) {
	plans := installment.ListPlansByAccount(accountID)
	result := make([]*installment.InstallmentPlan, len(plans))
	for i := range plans {
		result[i] = &plans[i]
	}
	return result, nil
}

func (r *memoryInstallmentRepository) ListByUser(ctx context.Context, userID string) ([]*installment.InstallmentPlan, error) {
	// 简化实现：返回所有分期计划（内存实现不区分用户）
	plans := installment.ListPlansByAccount("") // 空字符串返回所有
	result := make([]*installment.InstallmentPlan, 0, len(plans))
	for i := range plans {
		result = append(result, &plans[i])
	}
	return result, nil
}

func (r *memoryInstallmentRepository) Create(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	installment.AddPlan(*plan)
	return nil
}

func (r *memoryInstallmentRepository) Update(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	// 内存存储中没有实现
	return nil
}

// memoryRecurringIncomeRepository 内存周期收入计划仓储实现
type memoryRecurringIncomeRepository struct{}

func (r *memoryRecurringIncomeRepository) GetByID(ctx context.Context, userID, id string) (*recurringincome.RecurringIncomePlan, error) {
	plan := recurringincome.GetPlanByID(id)
	if plan == nil {
		return nil, fmt.Errorf("recurring income plan not found: %s", id)
	}
	return plan, nil
}

func (r *memoryRecurringIncomeRepository) ListByUser(ctx context.Context, userID string) ([]*recurringincome.RecurringIncomePlan, error) {
	plans := recurringincome.ListPlans()
	result := make([]*recurringincome.RecurringIncomePlan, len(plans))
	for i := range plans {
		result[i] = &plans[i]
	}
	return result, nil
}

func (r *memoryRecurringIncomeRepository) Create(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	recurringincome.SavePlan(*plan)
	return nil
}

func (r *memoryRecurringIncomeRepository) Update(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	recurringincome.SavePlan(*plan)
	return nil
}
