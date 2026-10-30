package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"neverbroke/internal/categories"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	"neverbroke/internal/money"

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
	if len(from) < 10 || len(to) < 10 {
		return nil, fmt.Errorf("invalid date range")
	}
	var dbTxs []models.Transaction

	err := r.db.WithContext(ctx).
		Where("transactions.user_id = ? AND occurred_at >= ? AND occurred_at < ?", userID, from[:10], to[:10]).
		Where("account_id IN (SELECT id FROM accounts WHERE user_id = ? AND currency = ?) OR (account_id IS NULL AND ? = 'CNY')", userID, money.Currency(ctx), money.Currency(ctx)).
		Order("occurred_at DESC").
		Find(&dbTxs).Error

	if err != nil {
		return nil, err
	}

	result := make([]*ledger.Transaction, len(dbTxs))
	// Keep legacy reports and MCP readers consistent with reviewed workspace categories.
	var overlays []struct {
		TransactionID string
		Payload       string
	}
	categoryOverrides := map[string]string{}
	if r.db.Migrator().HasTable("workspace_states") {
		if err := r.db.WithContext(ctx).Table("workspace_states AS ws").
			Select("COALESCE(NULLIF(ci.transaction_id, ''), CASE WHEN ws.key LIKE 'tx:%' THEN substring(ws.key from 4) ELSE '' END) AS transaction_id, ws.payload").
			Joins("LEFT JOIN curated_import_records AS ci ON ci.user_id = ws.user_id AND ci.source_key = ws.key").
			Where("ws.user_id = ?", userID).Scan(&overlays).Error; err != nil {
			return nil, err
		}
		for _, row := range overlays {
			var edit struct {
				Category string `json:"category"`
			}
			if err := json.Unmarshal([]byte(row.Payload), &edit); err != nil {
				return nil, err
			}
			if edit.Category != "" {
				categoryOverrides[row.TransactionID] = edit.Category
			}
		}
	}
	for i, dbTx := range dbTxs {
		result[i] = toTransactionModel(&dbTx)
		if category, ok := categoryOverrides[dbTx.ID]; ok {
			result[i].CategoryKey = category
		}
		result[i].CategoryKey = categories.Resolve(result[i].CategoryKey).ID
		result[i].Currency = money.Currency(ctx)
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
