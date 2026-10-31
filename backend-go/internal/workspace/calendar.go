package workspace

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"neverbroke/internal/categories"
	"neverbroke/internal/http/response"
)

// Calendar amounts are integer cents. Repayment obligations never become spending.
type CalendarMetric struct {
	Metric
	Billed            int64 `json:"billed"`
	Estimated         int64 `json:"estimated"`
	BilledCount       int   `json:"billedCount"`
	EstimatedCount    int   `json:"estimatedCount"`
	UnknownRepayments int   `json:"unknownRepayments"`
}

type CalendarDay struct {
	Date string `json:"date"`
	CalendarMetric
}

type CalendarQuery struct {
	Month, Day, View string
	Filter
}

func (q CalendarQuery) validate() (time.Time, error) {
	first, err := time.Parse("2006-01", q.Month)
	if err != nil || first.Year() < 1900 || first.Year() > 2200 {
		return time.Time{}, errors.New("月份无效")
	}
	day, err := time.Parse("2006-01-02", q.Day)
	if err != nil || day.Format("2006-01") != q.Month {
		return time.Time{}, errors.New("所选日期必须在当前月份内")
	}
	if q.Currency != "CNY" && q.Currency != "USD" {
		return time.Time{}, errors.New("币种无效")
	}
	if q.View != "all" && q.View != "spending" && q.View != "repayment" {
		return time.Time{}, errors.New("日历视图无效")
	}
	return first, nil
}

func (m *CalendarMetric) add(x Entry) {
	if x.Kind != "schedule" {
		m.Metric.Add(x)
		return
	}
	if x.Edit.Status == "billed" {
		m.BilledCount++
		if x.Cents != nil {
			m.Billed += *x.Cents
		}
	} else {
		m.EstimatedCount++
		if x.Cents != nil {
			m.Estimated += *x.Cents
		}
	}
	if x.Cents == nil {
		m.UnknownRepayments++
	}
}

func BuildCalendar(items []Entry, q CalendarQuery) ([]CalendarDay, CalendarMetric, []Entry, error) {
	first, err := q.validate()
	if err != nil {
		return nil, CalendarMetric{}, nil, err
	}
	next := first.AddDate(0, 1, 0)
	q.From, q.To = first.Format("2006-01-02"), next.AddDate(0, 0, -1).Format("2006-01-02")
	days := []CalendarDay{}
	index := map[string]int{}
	for d := first; d.Before(next); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		index[date] = len(days)
		days = append(days, CalendarDay{Date: date})
	}
	total, selected := CalendarMetric{}, []Entry{}
	for _, x := range items {
		if !Matches(x, q.Filter) {
			continue
		}
		switch x.Kind {
		case "expense", "refund":
			if q.View == "repayment" {
				continue
			}
		case "schedule":
			if q.View == "spending" || x.Edit.Status == "paid" || x.Edit.Status == "cancelled" || x.Edit.Status == "replaced" {
				continue
			}
		default:
			continue
		}
		i, exists := index[x.Date]
		if !exists {
			continue
		}
		days[i].add(x)
		total.add(x)
		if x.Date == q.Day {
			selected = append(selected, x)
		}
	}
	return days, total, selected, nil
}

func calendar(c *gin.Context, db *gorm.DB) {
	q := CalendarQuery{Month: c.Query("month"), Day: c.Query("day"), View: c.DefaultQuery("view", "all"), Filter: Filter{Currency: c.DefaultQuery("currency", "CNY"), Account: c.Query("account")}}
	if _, err := q.validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	items, accounts, err := Load(db.WithContext(c.Request.Context()), c.GetString("user_id"))
	if err != nil {
		response.InternalError(c, "读取日历失败")
		return
	}
	days, totals, rows, err := BuildCalendar(items, q)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	offset = max(0, min(len(rows), offset))
	end := min(offset+50, len(rows))
	response.OK(c, gin.H{"month": q.Month, "day": q.Day, "currency": q.Currency, "days": days, "totals": totals, "items": rows[offset:end], "total": len(rows), "offset": offset, "hasMore": end < len(rows), "accounts": accounts, "categories": categories.All})
}
