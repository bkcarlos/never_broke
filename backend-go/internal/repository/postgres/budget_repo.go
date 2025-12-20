package postgres

import (
	"context"
	"neverbroke/internal/budget"
	"neverbroke/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pgBudgetRepository struct {
	db *gorm.DB
}

// NewPgBudgetRepository 创建新的 Postgres Budget 仓储
func NewPgBudgetRepository(db *gorm.DB) *pgBudgetRepository {
	return &pgBudgetRepository{db: db}
}

func (r *pgBudgetRepository) GetByMonth(ctx context.Context, userID, month string) (*budget.MonthlyBudget, error) {
	var dbBudget models.Budget

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND month = ?", userID, month).
		First(&dbBudget).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &budget.MonthlyBudget{
		Month:       dbBudget.Month,
		TotalBudget: dbBudget.TotalBudget,
	}, nil
}

func (r *pgBudgetRepository) Upsert(ctx context.Context, userID string, b *budget.MonthlyBudget) error {
	dbBudget := &models.Budget{
		ID:          uuid.NewString(),
		UserID:      userID,
		Month:       b.Month,
		TotalBudget: b.TotalBudget,
	}

	// 使用 upsert：如果 user_id + month 冲突，则更新 total_budget
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "month"}},
			DoUpdates: clause.AssignmentColumns([]string{"total_budget", "updated_at"}),
		}).
		Create(dbBudget).Error
}
