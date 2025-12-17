package postgres

import (
	"context"
	"neverbroke/internal/accounts"
	"neverbroke/internal/models"

	"gorm.io/gorm"
)

type pgAccountRepository struct {
	db *gorm.DB
}

func NewPgAccountRepository(db *gorm.DB) *pgAccountRepository {
	return &pgAccountRepository{db: db}
}

func (r *pgAccountRepository) GetByID(ctx context.Context, userID, id string) (*accounts.Account, error) {
	var dbAccount models.Account

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&dbAccount).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 转换为业务模型
	return toAccountModel(&dbAccount), nil
}

func (r *pgAccountRepository) ListByUser(ctx context.Context, userID string) ([]*accounts.Account, error) {
	var dbAccounts []models.Account

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("archived ASC, id ASC").
		Find(&dbAccounts).Error

	if err != nil {
		return nil, err
	}

	// 转换为业务模型
	result := make([]*accounts.Account, len(dbAccounts))
	for i, dbAcc := range dbAccounts {
		result[i] = toAccountModel(&dbAcc)
	}

	return result, nil
}

func (r *pgAccountRepository) Create(ctx context.Context, userID string, acc *accounts.Account) error {
	dbAccount := toDBAccount(userID, acc)
	return r.db.WithContext(ctx).Create(dbAccount).Error
}

func (r *pgAccountRepository) Update(ctx context.Context, userID string, acc *accounts.Account) error {
	dbAccount := toDBAccount(userID, acc)
	return r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, acc.ID).
		Updates(dbAccount).Error
}

// 辅助函数：将业务模型转换为数据库模型
func toDBAccount(userID string, acc *accounts.Account) *models.Account {
	dbAcc := &models.Account{
		ID:       acc.ID,
		UserID:   userID,
		Type:     string(acc.Type),
		Currency: acc.Currency,
		Balance:  acc.Balance,
		Archived: acc.Archived,
	}

	if acc.Institution != "" {
		inst := string(acc.Institution)
		dbAcc.Institution = &inst
	}
	if acc.CardLast4 != "" {
		dbAcc.CardLast4 = &acc.CardLast4
	}
	if acc.CreditLimit > 0 {
		dbAcc.CreditLimit = &acc.CreditLimit
	}
	if acc.TempLimit > 0 {
		dbAcc.TempLimit = &acc.TempLimit
	}

	return dbAcc
}

// 辅助函数：将数据库模型转换为业务模型
func toAccountModel(dbAcc *models.Account) *accounts.Account {
	acc := &accounts.Account{
		ID:       dbAcc.ID,
		Type:     accounts.AccountType(dbAcc.Type),
		Currency: dbAcc.Currency,
		Balance:  dbAcc.Balance,
		Archived: dbAcc.Archived,
	}

	if dbAcc.Institution != nil {
		acc.Institution = accounts.AccountInstitution(*dbAcc.Institution)
	}
	if dbAcc.CardLast4 != nil {
		acc.CardLast4 = *dbAcc.CardLast4
	}
	if dbAcc.CreditLimit != nil {
		acc.CreditLimit = *dbAcc.CreditLimit
	}
	if dbAcc.TempLimit != nil {
		acc.TempLimit = *dbAcc.TempLimit
	}

	return acc
}
