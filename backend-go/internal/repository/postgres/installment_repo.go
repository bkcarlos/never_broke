package postgres

import (
	"context"
	"neverbroke/internal/installment"
	"neverbroke/internal/models"
	"time"

	"gorm.io/gorm"
)

type pgInstallmentRepository struct {
	db *gorm.DB
}

func NewPgInstallmentRepository(db *gorm.DB) *pgInstallmentRepository {
	return &pgInstallmentRepository{db: db}
}

func (r *pgInstallmentRepository) GetByID(ctx context.Context, userID, id string) (*installment.InstallmentPlan, error) {
	var dbPlan models.InstallmentPlan

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&dbPlan).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return toInstallmentModel(&dbPlan), nil
}

func (r *pgInstallmentRepository) ListByAccount(ctx context.Context, userID, accountID string) ([]*installment.InstallmentPlan, error) {
	var dbPlans []models.InstallmentPlan

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND account_id = ?", userID, accountID).
		Order("created_at DESC").
		Find(&dbPlans).Error

	if err != nil {
		return nil, err
	}

	result := make([]*installment.InstallmentPlan, len(dbPlans))
	for i, dbPlan := range dbPlans {
		result[i] = toInstallmentModel(&dbPlan)
	}

	return result, nil
}

func (r *pgInstallmentRepository) ListByUser(ctx context.Context, userID string) ([]*installment.InstallmentPlan, error) {
	var dbPlans []models.InstallmentPlan

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&dbPlans).Error

	if err != nil {
		return nil, err
	}

	result := make([]*installment.InstallmentPlan, len(dbPlans))
	for i, dbPlan := range dbPlans {
		result[i] = toInstallmentModel(&dbPlan)
	}

	return result, nil
}

func (r *pgInstallmentRepository) Create(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	dbPlan, err := toDBInstallment(userID, plan)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(dbPlan).Error
}

func (r *pgInstallmentRepository) Update(ctx context.Context, userID string, plan *installment.InstallmentPlan) error {
	dbPlan, err := toDBInstallment(userID, plan)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, plan.ID).
		Updates(dbPlan).Error
}

// 辅助函数：将业务模型转换为数据库模型
func toDBInstallment(userID string, plan *installment.InstallmentPlan) (*models.InstallmentPlan, error) {
	firstDueDate, err := time.Parse("2006-01-02", plan.FirstDueDate)
	if err != nil {
		return nil, err
	}

	createdAt := time.Now()
	if plan.CreatedAt != "" {
		createdAt, err = time.Parse(time.RFC3339, plan.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	return &models.InstallmentPlan{
		ID:            plan.ID,
		UserID:        userID,
		AccountID:     plan.AccountID,
		Principal:     plan.Principal,
		TotalInterest: plan.TotalInterest,
		TotalAmount:   plan.TotalAmount,
		TermMonths:    plan.TermMonths,
		MonthlyRate:   plan.MonthlyRate,
		RepayMethod:   string(plan.RepayMethod),
		Currency:      plan.Currency,
		FirstDueDate:  firstDueDate,
		Status:        string(plan.Status),
		CreatedAt:     createdAt,
	}, nil
}

// 辅助函数：将数据库模型转换为业务模型
func toInstallmentModel(dbPlan *models.InstallmentPlan) *installment.InstallmentPlan {
	return &installment.InstallmentPlan{
		ID:            dbPlan.ID,
		AccountID:     dbPlan.AccountID,
		Principal:     dbPlan.Principal,
		TotalInterest: dbPlan.TotalInterest,
		TotalAmount:   dbPlan.TotalAmount,
		TermMonths:    dbPlan.TermMonths,
		MonthlyRate:   dbPlan.MonthlyRate,
		RepayMethod:   installment.InstallmentRepayMethod(dbPlan.RepayMethod),
		Currency:      dbPlan.Currency,
		FirstDueDate:  dbPlan.FirstDueDate.Format("2006-01-02"),
		Status:        installment.InstallmentPlanStatus(dbPlan.Status),
		CreatedAt:     dbPlan.CreatedAt.Format(time.RFC3339),
	}
}
