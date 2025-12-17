package postgres

import (
	"context"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"

	"gorm.io/gorm"
)

type pgTransactionRepository struct {
	db *gorm.DB
}

func NewPgTransactionRepository(db *gorm.DB) *pgTransactionRepository {
	return &pgTransactionRepository{db: db}
}

func (r *pgTransactionRepository) GetByID(ctx context.Context, userID, id string) (*ledger.Transaction, error) {
	var dbTx models.Transaction

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&dbTx).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return toTransactionModel(&dbTx), nil
}

func (r *pgTransactionRepository) ListByUserBetween(ctx context.Context, userID, from, to string) ([]*ledger.Transaction, error) {
	var dbTxs []models.Transaction

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND occurred_at >= ? AND occurred_at < ?", userID, from, to).
		Order("occurred_at DESC").
		Find(&dbTxs).Error

	if err != nil {
		return nil, err
	}

	result := make([]*ledger.Transaction, len(dbTxs))
	for i, dbTx := range dbTxs {
		result[i] = toTransactionModel(&dbTx)
	}

	return result, nil
}

func (r *pgTransactionRepository) Create(ctx context.Context, userID string, tx *ledger.Transaction) error {
	dbTx := toDBTransaction(userID, tx)
	return r.db.WithContext(ctx).Create(dbTx).Error
}

func (r *pgTransactionRepository) CreateBatch(ctx context.Context, userID string, txs []*ledger.Transaction) error {
	if len(txs) == 0 {
		return nil
	}

	dbTxs := make([]models.Transaction, len(txs))
	for i, tx := range txs {
		dbTxs[i] = *toDBTransaction(userID, tx)
	}

	return r.db.WithContext(ctx).Create(&dbTxs).Error
}

// 辅助函数：将业务模型转换为数据库模型
func toDBTransaction(userID string, tx *ledger.Transaction) *models.Transaction {
	dbTx := &models.Transaction{
		ID:         tx.ID,
		UserID:     userID,
		Type:       string(tx.Type),
		Amount:     tx.Amount,
		OccurredAt: tx.OccurredAt,
	}

	if tx.CategoryKey != "" {
		dbTx.CategoryKey = &tx.CategoryKey
	}
	if tx.Remark != "" {
		dbTx.Remark = &tx.Remark
	}
	if tx.AccountID != "" {
		dbTx.AccountID = &tx.AccountID
	}
	if tx.FromAccountID != "" {
		dbTx.FromAccountID = &tx.FromAccountID
	}
	if tx.ToAccountID != "" {
		dbTx.ToAccountID = &tx.ToAccountID
	}
	if tx.InstallmentPlanID != "" {
		dbTx.InstallmentPlanID = &tx.InstallmentPlanID
	}
	if tx.InstallmentPeriod > 0 {
		dbTx.InstallmentPeriod = &tx.InstallmentPeriod
	}
	if tx.InstallmentTotal > 0 {
		dbTx.InstallmentTotal = &tx.InstallmentTotal
	}

	return dbTx
}

// 辅助函数：将数据库模型转换为业务模型
func toTransactionModel(dbTx *models.Transaction) *ledger.Transaction {
	tx := &ledger.Transaction{
		ID: dbTx.ID,
		CreateTransactionRequest: ledger.CreateTransactionRequest{
			Type:       ledger.TransactionType(dbTx.Type),
			Amount:     dbTx.Amount,
			OccurredAt: dbTx.OccurredAt,
		},
	}

	if dbTx.CategoryKey != nil {
		tx.CategoryKey = *dbTx.CategoryKey
	}
	if dbTx.Remark != nil {
		tx.Remark = *dbTx.Remark
	}
	if dbTx.AccountID != nil {
		tx.AccountID = *dbTx.AccountID
	}
	if dbTx.FromAccountID != nil {
		tx.FromAccountID = *dbTx.FromAccountID
	}
	if dbTx.ToAccountID != nil {
		tx.ToAccountID = *dbTx.ToAccountID
	}
	if dbTx.InstallmentPlanID != nil {
		tx.InstallmentPlanID = *dbTx.InstallmentPlanID
	}
	if dbTx.InstallmentPeriod != nil {
		tx.InstallmentPeriod = *dbTx.InstallmentPeriod
	}
	if dbTx.InstallmentTotal != nil {
		tx.InstallmentTotal = *dbTx.InstallmentTotal
	}

	return tx
}
