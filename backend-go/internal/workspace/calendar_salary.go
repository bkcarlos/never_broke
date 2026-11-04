package workspace

import (
	"gorm.io/gorm"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"neverbroke/internal/recurringincome"
)

// Salary estimates are separate from ledger entries: viewing a forecast never posts income.
type CalendarSalary struct {
	recurringincome.SalaryPreviewItem
	Key             string                               `json:"key"`
	Title           string                               `json:"title"`
	AccountID       string                               `json:"accountId"`
	Currency        string                               `json:"currency"`
	CityCode        string                               `json:"cityCode"`
	BasicDeduction  float64                              `json:"basicDeduction"`
	ExtraDeductions float64                              `json:"extraDeductions"`
	Configuration   *recurringincome.SalaryContributions `json:"configuration,omitempty"`
}

func loadCalendarSalaries(db *gorm.DB, userID string, q CalendarQuery) ([]CalendarSalary, error) {
	var plans []models.RecurringIncomePlan
	query := db.Where("user_id = ? AND type = ? AND active = ?", userID, "salary", true)
	if err := query.Order("created_at, id").Find(&plans).Error; err != nil {
		return nil, err
	}
	return buildCalendarSalaryHistory(plans, q, payroll.NewService(payroll.NewRepository(db)).GetRegionRuleAt)
}

func buildCalendarSalaries(plans []models.RecurringIncomePlan, q CalendarQuery) ([]CalendarSalary, error) {
	return buildCalendarSalaryHistory(plans, q, payroll.NewService(nil).GetRegionRuleAt)
}

func buildCalendarSalaryHistory(plans []models.RecurringIncomePlan, q CalendarQuery, lookup recurringincome.PolicyLookup) ([]CalendarSalary, error) {
	result := []CalendarSalary{}
	first, err := q.validate()
	if err != nil {
		return nil, err
	}
	for _, p := range plans {
		if !p.Active || p.Type != "salary" {
			continue
		}
		versions, err := recurringincome.SalaryVersions(p)
		if err != nil {
			return nil, err
		}
		items, err := recurringincome.PreviewSalaryHistory(versions, first.Year(), lookup)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			v := item.Config
			if item.Date[:7] != q.Month || v.Currency != q.Currency || (q.Account != "" && v.AccountID != q.Account) {
				continue
			}
			title := p.Label
			if title == "" {
				title = "工资"
			}
			result = append(result, CalendarSalary{SalaryPreviewItem: item.SalaryPreviewItem, Key: "salary:" + p.ID + ":" + item.Date, Title: title, AccountID: v.AccountID, Currency: v.Currency, CityCode: v.CityCode, BasicDeduction: recurringincome.MonthlyDeduction, ExtraDeductions: v.ExtraDeductions, Configuration: item.Configuration})
		}
	}
	return result, nil
}
