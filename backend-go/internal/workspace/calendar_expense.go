package workspace

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// CalendarFixedExpense is a planned occurrence, never an actual ledger entry.
// It is separate from repayments so hiding repayments does not hide rent, etc.
type CalendarFixedExpense struct {
	Key        string `json:"key"`
	SettingKey string `json:"settingKey"`
	Title      string `json:"title"`
	Date       string `json:"date"`
	Cents      int64  `json:"cents"`
	Currency   string `json:"currency"`
	AccountID  string `json:"accountId"`
	Period     string `json:"period"`
}

func loadCalendarFixedExpenses(db *gorm.DB, userID string, q CalendarQuery) ([]CalendarFixedExpense, error) {
	var settings []Setting
	if err := db.Where("user_id = ? AND kind IN ? AND enabled = ? AND currency = ?", userID, []string{"fixed_expense", "subscription"}, true, q.Currency).Order("key").Find(&settings).Error; err != nil {
		return nil, err
	}
	return buildCalendarFixedExpenses(settings, q)
}

func buildCalendarFixedExpenses(settings []Setting, q CalendarQuery) ([]CalendarFixedExpense, error) {
	first, err := q.validate()
	if err != nil {
		return nil, err
	}
	result := []CalendarFixedExpense{}
	for _, s := range settings {
		if !s.Enabled || (s.Kind != "fixed_expense" && s.Kind != "subscription") || s.Currency != q.Currency || (q.Account != "" && s.AccountID != q.Account) {
			continue
		}
		if err := ValidateRecurringExpense(s); err != nil {
			return nil, fmt.Errorf("固定支出「%s」：%w", s.Name, err)
		}
		anchor, _ := time.Parse(time.DateOnly, s.Date)
		for _, date := range RecurrenceDates(anchor, s.Period, s.EndDate, first, first.AddDate(0, 1, 0)) {
			result = append(result, CalendarFixedExpense{Key: "fixed:" + s.Key + ":" + date, SettingKey: s.Key, Title: s.Name, Date: date, Cents: s.Cents, Currency: s.Currency, AccountID: s.AccountID, Period: s.Period})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Date == result[j].Date {
			return result[i].Key < result[j].Key
		}
		return result[i].Date < result[j].Date
	})
	return result, nil
}
