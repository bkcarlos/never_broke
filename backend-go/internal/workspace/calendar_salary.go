package workspace

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
	"neverbroke/internal/models"
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
	query := db.Where("user_id = ? AND type = ? AND active = ? AND currency = ?", userID, "salary", true, q.Currency)
	if q.Account != "" {
		query = query.Where("account_id = ?", q.Account)
	}
	if err := query.Order("created_at, id").Find(&plans).Error; err != nil {
		return nil, err
	}
	return buildCalendarSalaries(plans, q)
}

func buildCalendarSalaries(plans []models.RecurringIncomePlan, q CalendarQuery) ([]CalendarSalary, error) {
	first, err := q.validate()
	if err != nil {
		return nil, err
	}
	result := []CalendarSalary{}
	for _, p := range plans {
		if !p.Active || p.Type != "salary" || p.Currency != q.Currency || (q.Account != "" && p.AccountID != q.Account) {
			continue
		}
		// Generic fixed-income plans have no gross salary configuration to preview.
		if p.SalaryPreTaxMonthly == nil || p.SalaryStartDate == nil || p.SalaryPayDay == nil {
			continue
		}
		req := recurringincome.SalaryPreviewRequest{PreTaxAmount: *p.SalaryPreTaxMonthly, StartDate: p.SalaryStartDate.Format(time.DateOnly), PayDay: *p.SalaryPayDay, Frequency: recurringincome.RecurringIncomeFrequency(p.Frequency), Currency: p.Currency, SocialRate: p.SalarySocialRate}
		if p.CityCode != nil {
			req.CityCode = *p.CityCode
		}
		if p.SalaryExtraDeductions != nil {
			req.ExtraDeductions = *p.SalaryExtraDeductions
		}
		if p.SalaryContributions != nil {
			if err := json.Unmarshal([]byte(*p.SalaryContributions), &req.Contributions); err != nil {
				return nil, fmt.Errorf("salary contributions: %w", err)
			}
		}
		preview, err := recurringincome.PreviewSalaryYear(req, first.Year())
		if err != nil {
			return nil, err
		}
		for _, item := range preview.Items {
			if item.Date[:7] != q.Month {
				continue
			}
			title := p.Label
			if title == "" {
				title = "工资"
			}
			result = append(result, CalendarSalary{SalaryPreviewItem: item, Key: "salary:" + p.ID + ":" + item.Date, Title: title, AccountID: p.AccountID, Currency: p.Currency, CityCode: req.CityCode, BasicDeduction: recurringincome.MonthlyDeduction, ExtraDeductions: req.ExtraDeductions, Configuration: req.Contributions})
		}
	}
	return result, nil
}
