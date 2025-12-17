package postgres

import (
	"context"
	"neverbroke/internal/models"
	"neverbroke/internal/recurringincome"
	"time"

	"gorm.io/gorm"
)

type pgRecurringIncomeRepository struct {
	db *gorm.DB
}

func NewPgRecurringIncomeRepository(db *gorm.DB) *pgRecurringIncomeRepository {
	return &pgRecurringIncomeRepository{db: db}
}

func (r *pgRecurringIncomeRepository) GetByID(ctx context.Context, userID, id string) (*recurringincome.RecurringIncomePlan, error) {
	var dbPlan models.RecurringIncomePlan

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, id).
		First(&dbPlan).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return toRecurringIncomeModel(&dbPlan), nil
}

func (r *pgRecurringIncomeRepository) ListByUser(ctx context.Context, userID string) ([]*recurringincome.RecurringIncomePlan, error) {
	var dbPlans []models.RecurringIncomePlan

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("active DESC, created_at DESC").
		Find(&dbPlans).Error

	if err != nil {
		return nil, err
	}

	result := make([]*recurringincome.RecurringIncomePlan, len(dbPlans))
	for i, dbPlan := range dbPlans {
		result[i] = toRecurringIncomeModel(&dbPlan)
	}

	return result, nil
}

func (r *pgRecurringIncomeRepository) Create(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	dbPlan, err := toDBRecurringIncome(userID, plan)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(dbPlan).Error
}

func (r *pgRecurringIncomeRepository) Update(ctx context.Context, userID string, plan *recurringincome.RecurringIncomePlan) error {
	dbPlan, err := toDBRecurringIncome(userID, plan)
	if err != nil {
		return err
	}

	return r.db.WithContext(ctx).
		Where("user_id = ? AND id = ?", userID, plan.ID).
		Updates(dbPlan).Error
}

// 辅助函数：将业务模型转换为数据库模型
func toDBRecurringIncome(userID string, plan *recurringincome.RecurringIncomePlan) (*models.RecurringIncomePlan, error) {
	nextDueDate, err := time.Parse("2006-01-02", plan.NextDueDate)
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

	dbPlan := &models.RecurringIncomePlan{
		ID:          plan.ID,
		UserID:      userID,
		Type:        string(plan.Type),
		Label:       plan.Label,
		AccountID:   plan.AccountID,
		Amount:      plan.Amount,
		Currency:    plan.Currency,
		Frequency:   string(plan.Frequency),
		NextDueDate: nextDueDate,
		Active:      plan.Active,
		CreatedAt:   createdAt,
	}

	if plan.DayOfMonth > 0 {
		dbPlan.DayOfMonth = &plan.DayOfMonth
	}
	if plan.Weekday > 0 {
		dbPlan.Weekday = &plan.Weekday
	}
	if plan.LastFiredDate != "" {
		t, err := time.Parse("2006-01-02", plan.LastFiredDate)
		if err != nil {
			return nil, err
		}
		dbPlan.LastFiredDate = &t
	}
	if plan.CityCode != "" {
		dbPlan.CityCode = &plan.CityCode
	}
	if plan.SalaryPreTaxMonthly > 0 {
		dbPlan.SalaryPreTaxMonthly = &plan.SalaryPreTaxMonthly
	}
	if plan.SalarySocialRate > 0 {
		dbPlan.SalarySocialRate = &plan.SalarySocialRate
	}
	if plan.SalaryExtraDeductions > 0 {
		dbPlan.SalaryExtraDeductions = &plan.SalaryExtraDeductions
	}
	if plan.SalaryPayDay > 0 {
		dbPlan.SalaryPayDay = &plan.SalaryPayDay
	}
	if plan.SalaryStartDate != "" {
		t, err := time.Parse("2006-01-02", plan.SalaryStartDate)
		if err != nil {
			return nil, err
		}
		dbPlan.SalaryStartDate = &t
	}

	return dbPlan, nil
}

// 辅助函数：将数据库模型转换为业务模型
func toRecurringIncomeModel(dbPlan *models.RecurringIncomePlan) *recurringincome.RecurringIncomePlan {
	plan := &recurringincome.RecurringIncomePlan{
		ID:          dbPlan.ID,
		Type:        recurringincome.RecurringIncomeType(dbPlan.Type),
		Label:       dbPlan.Label,
		AccountID:   dbPlan.AccountID,
		Amount:      dbPlan.Amount,
		Currency:    dbPlan.Currency,
		Frequency:   recurringincome.RecurringIncomeFrequency(dbPlan.Frequency),
		NextDueDate: dbPlan.NextDueDate.Format("2006-01-02"),
		Active:      dbPlan.Active,
		CreatedAt:   dbPlan.CreatedAt.Format(time.RFC3339),
	}

	if dbPlan.DayOfMonth != nil {
		plan.DayOfMonth = *dbPlan.DayOfMonth
	}
	if dbPlan.Weekday != nil {
		plan.Weekday = *dbPlan.Weekday
	}
	if dbPlan.LastFiredDate != nil && !dbPlan.LastFiredDate.IsZero() {
		plan.LastFiredDate = dbPlan.LastFiredDate.Format("2006-01-02")
	}
	if dbPlan.CityCode != nil {
		plan.CityCode = *dbPlan.CityCode
	}
	if dbPlan.SalaryPreTaxMonthly != nil {
		plan.SalaryPreTaxMonthly = *dbPlan.SalaryPreTaxMonthly
	}
	if dbPlan.SalarySocialRate != nil {
		plan.SalarySocialRate = *dbPlan.SalarySocialRate
	}
	if dbPlan.SalaryExtraDeductions != nil {
		plan.SalaryExtraDeductions = *dbPlan.SalaryExtraDeductions
	}
	if dbPlan.SalaryPayDay != nil {
		plan.SalaryPayDay = *dbPlan.SalaryPayDay
	}
	if dbPlan.SalaryStartDate != nil && !dbPlan.SalaryStartDate.IsZero() {
		plan.SalaryStartDate = dbPlan.SalaryStartDate.Format("2006-01-02")
	}

	return plan
}
