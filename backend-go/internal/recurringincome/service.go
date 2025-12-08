package recurringincome

import "time"

// FindDuePlans 给定一个日期字符串 YYYY-MM-DD，找出 nextDueDate <= date 且 Active 的计划
func FindDuePlans(date string) []RecurringIncomePlan {
	mu.RLock()
	defer mu.RUnlock()

	target, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil
	}

	var result []RecurringIncomePlan
	for _, p := range plans {
		if !p.Active || p.NextDueDate == "" {
			continue
		}
		due, err := time.Parse("2006-01-02", p.NextDueDate)
		if err != nil {
			continue
		}
		if !due.After(target) { // due <= target
			result = append(result, p)
		}
	}
	return result
}

// CalcNextDueDate 计算下一次应收日期
func CalcNextDueDate(p RecurringIncomePlan) string {
	if p.NextDueDate == "" {
		return ""
	}
	d, err := time.Parse("2006-01-02", p.NextDueDate)
	if err != nil {
		return ""
	}

	switch p.Frequency {
	case FrequencyMonthly:
		return d.AddDate(0, 1, 0).Format("2006-01-02")
	case FrequencyBimonthly:
		return d.AddDate(0, 2, 0).Format("2006-01-02")
	case FrequencyQuarterly:
		return d.AddDate(0, 3, 0).Format("2006-01-02")
	case FrequencySemiannual:
		return d.AddDate(0, 6, 0).Format("2006-01-02")
	case FrequencyWeekly:
		return d.AddDate(0, 0, 7).Format("2006-01-02")
	case FrequencyYearly:
		return d.AddDate(1, 0, 0).Format("2006-01-02")
	default:
		return d.Format("2006-01-02")
	}
}
