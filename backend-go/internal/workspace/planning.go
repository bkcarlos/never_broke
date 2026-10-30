package workspace

import (
	"fmt"
	"gorm.io/gorm"
	"math"
	"neverbroke/internal/installment"
	"neverbroke/internal/models"
	"time"
)

func nativePlans(db *gorm.DB, user string, accounts map[string]models.Account) ([]Entry, error) {
	var plans []models.InstallmentPlan
	if !db.Migrator().HasTable(&models.InstallmentPlan{}) {
		return []Entry{}, nil
	}
	if e := db.Where("user_id = ? AND status = ?", user, "active").Find(&plans).Error; e != nil {
		return nil, e
	}
	result := []Entry{}
	for _, p := range plans {
		schedule, e := installment.BuildSchedule(installment.PreviewInstallmentRequest{Principal: p.Principal, TermMonths: p.TermMonths, MonthlyRatePercent: p.MonthlyRate * 100, RepayMethod: installment.InstallmentRepayMethod(p.RepayMethod), FirstDueDate: p.FirstDueDate.Format("2006-01-02"), Currency: p.Currency})
		if e != nil {
			return nil, e
		}
		for _, s := range schedule.Schedule {
			cents := int64(math.Round(s.Total * 100))
			a := accounts[p.AccountID]
			result = append(result, Entry{Key: fmt.Sprintf("native:%s:%d", p.ID, s.Period), Kind: "schedule", Date: s.DueDate, Currency: p.Currency, Cents: &cents, Description: fmt.Sprintf("原生分期 · 第 %d/%d 期", s.Period, p.TermMonths), AccountID: p.AccountID, AccountName: accountName(a), Confirmation: "manual", Metadata: map[string]any{"分期计划": p.ID, "本期期数": s.Period, "总期数": p.TermMonths, "本金": s.Principal, "利息": s.Interest}, Edit: Edit{Status: "estimated"}})
		}
	}
	return result, nil
}

type ProjectionPoint struct {
	Date      string `json:"date"`
	Billed    int64  `json:"billed"`
	Estimated int64  `json:"estimated"`
	Remaining *int64 `json:"remaining"`
}
type Projection struct {
	Starting       *int64            `json:"starting"`
	Unverified     int               `json:"unverified"`
	UnknownAmounts int               `json:"unknownAmounts"`
	Points         []ProjectionPoint `json:"points"`
}

func Project(items []Entry, accounts []models.Account, settings []Setting, currency, from string, days int) Projection {
	out := Projection{Points: []ProjectionPoint{}}
	verified := map[string]bool{}
	for _, s := range settings {
		if s.Kind == "balance" {
			verified[s.Match] = true
		}
	}
	var start int64
	cashAccounts := 0
	for _, a := range accounts {
		if a.Currency != currency || a.Archived || a.Type == "credit" {
			continue
		}
		cashAccounts++
		if !verified[a.ID] {
			out.Unverified++
		}
		start += int64(math.Round(a.Balance * 100))
	}
	if cashAccounts > 0 && out.Unverified == 0 {
		v := start
		out.Starting = &v
	}
	date, e := time.Parse("2006-01-02", from)
	if e != nil {
		return out
	}
	byDate := map[string]*ProjectionPoint{}
	for i := 0; i < days; i++ {
		d := date.AddDate(0, 0, i).Format("2006-01-02")
		byDate[d] = &ProjectionPoint{Date: d}
	}
	for _, x := range items {
		if x.Kind != "schedule" || x.Currency != currency || x.Edit.Status == "paid" || x.Edit.Status == "cancelled" || x.Edit.Status == "replaced" {
			continue
		}
		p, ok := byDate[x.Date]
		if !ok {
			continue
		}
		if x.Cents == nil {
			out.UnknownAmounts++
			continue
		}
		if x.Edit.Status == "billed" {
			p.Billed += *x.Cents
		} else {
			p.Estimated += *x.Cents
		}
	}
	for i := 0; i < days; i++ {
		d := date.AddDate(0, 0, i).Format("2006-01-02")
		p := byDate[d]
		start -= p.Billed + p.Estimated
		if out.Starting != nil && out.UnknownAmounts == 0 {
			v := start
			p.Remaining = &v
		}
		out.Points = append(out.Points, *p)
	}
	return out
}
