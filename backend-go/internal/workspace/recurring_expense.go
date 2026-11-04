package workspace

import (
	"errors"
	"strings"
	"time"
)

func ValidateRecurringExpense(s Setting) error {
	if strings.TrimSpace(s.Name) == "" || s.Cents <= 0 {
		return errors.New("请填写固定支出名称和大于零的金额")
	}
	switch s.Period {
	case "monthly", "quarterly", "yearly", "weekly":
	default:
		return errors.New("固定支出周期无效")
	}
	first, err := time.Parse(time.DateOnly, s.Date)
	if err != nil || first.Year() < 1900 || first.Year() > 2200 {
		return errors.New("首次扣款日期无效")
	}
	if s.EndDate != "" {
		end, err := time.Parse(time.DateOnly, s.EndDate)
		if err != nil || end.Before(first) || end.Year() > 2200 {
			return errors.New("结束日期不能早于首次扣款日期")
		}
	}
	return nil
}

// RecurrenceDates keeps the original day anchor (Jan 31 -> Feb 28 -> Mar 31).
// The end of the query is exclusive; the plan's last date is inclusive.
func RecurrenceDates(first time.Time, period, last string, from, to time.Time) []string {
	result := []string{}
	if first.IsZero() || first.Year() < 1900 {
		return result
	}
	step := map[string]int{"monthly": 1, "bimonthly": 2, "quarterly": 3, "semiannual": 6, "yearly": 12}[period]
	if step == 0 && period != "weekly" {
		return result
	}
	anchor := time.Date(first.Year(), first.Month(), 1, 0, 0, 0, 0, time.UTC)
	for n := 0; ; n++ {
		var date time.Time
		if period == "weekly" {
			date = time.Date(first.Year(), first.Month(), first.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, n*7)
		} else {
			month := anchor.AddDate(0, n*step, 0)
			day := min(first.Day(), month.AddDate(0, 1, -1).Day())
			date = month.AddDate(0, 0, day-1)
		}
		key := date.Format(time.DateOnly)
		if !date.Before(to) || (last != "" && key > last) {
			break
		}
		if !date.Before(from) {
			result = append(result, key)
		}
	}
	return result
}
