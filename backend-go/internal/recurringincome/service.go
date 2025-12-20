package recurringincome

import "time"

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
