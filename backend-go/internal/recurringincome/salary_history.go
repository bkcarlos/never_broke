package recurringincome

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"sort"
	"time"
)

type PolicyLookup func(string, string) (*payroll.RegionRuleResponse, error)
type SalaryHistoryItem struct {
	SalaryPreviewItem
	Config        UpsertSalaryPlanRequest
	Configuration *SalaryContributions
}

func SalaryVersions(p models.RecurringIncomePlan) ([]UpsertSalaryPlanRequest, error) {
	if p.SalaryVersions != nil {
		var versions []UpsertSalaryPlanRequest
		if err := json.Unmarshal([]byte(*p.SalaryVersions), &versions); err != nil {
			return nil, err
		}
		return versions, nil
	}
	if p.SalaryStartDate == nil || p.SalaryPreTaxMonthly == nil || p.SalaryPayDay == nil {
		return nil, nil
	}
	v := UpsertSalaryPlanRequest{StartDate: p.SalaryStartDate.Format(time.DateOnly), PreTaxMonthly: *p.SalaryPreTaxMonthly, PayDay: *p.SalaryPayDay, AccountID: p.AccountID, Currency: p.Currency, BaseMode: "fixed"}
	if p.CityCode != nil {
		v.CityCode = *p.CityCode
	}
	if p.SalarySocialRate != nil {
		v.SocialRate = *p.SalarySocialRate
	}
	if p.SalaryExtraDeductions != nil {
		v.ExtraDeductions = *p.SalaryExtraDeductions
	}
	if p.SalaryContributions != nil {
		if err := json.Unmarshal([]byte(*p.SalaryContributions), &v.Contributions); err != nil {
			return nil, err
		}
	}
	return []UpsertSalaryPlanRequest{v}, nil
}

func MergeSalaryVersion(versions []UpsertSalaryPlanRequest, v UpsertSalaryPlanRequest) ([]UpsertSalaryPlanRequest, error) {
	if _, err := time.Parse(time.DateOnly, v.StartDate); err != nil {
		return nil, errors.New("生效日期无效")
	}
	if v.BaseMode != "" && v.BaseMode != "fixed" && v.BaseMode != "policy" {
		return nil, errors.New("基数模式无效")
	}
	if v.PayMonthOffset < 0 || v.PayMonthOffset > 1 {
		return nil, errors.New("发薪月份偏移只能为 0 或 1")
	}
	if v.BaseMode == "policy" && v.Contributions == nil {
		return nil, errors.New("自动政策模式需要配置缴费比例")
	}
	out := []UpsertSalaryPlanRequest{}
	for _, old := range versions {
		if old.Currency != v.Currency {
			return nil, errors.New("薪资历史必须使用同一币种")
		}
		if old.PayMonthOffset != v.PayMonthOffset {
			return nil, errors.New("已有历史方案时，请保持相同的当月/次月发薪方式")
		}
		if old.StartDate[:7] != v.StartDate[:7] {
			out = append(out, old)
		}
	}
	out = append(out, v)
	sort.Slice(out, func(i, j int) bool { return out[i].StartDate < out[j].StartDate })
	return out, nil
}

// PreviewSalaryHistory selects by salary month, then accumulates tax by payment year.
// Account filters must be applied only after this calculation.
func PreviewSalaryHistory(versions []UpsertSalaryPlanRequest, year int, lookup PolicyLookup) ([]SalaryHistoryItem, error) {
	versions = append([]UpsertSalaryPlanRequest(nil), versions...)
	sort.Slice(versions, func(i, j int) bool { return versions[i].StartDate < versions[j].StartDate })
	out := []SalaryHistoryItem{}
	var grossYTD, taxableYTD, taxPaid float64
	estimated := false
	for paymentMonth := 1; paymentMonth <= 12; paymentMonth++ {
		payment := time.Date(year, time.Month(paymentMonth), 1, 0, 0, 0, 0, time.UTC)
		var selected *UpsertSalaryPlanRequest
		var salaryMonth time.Time
		for i := range versions {
			v := &versions[i]
			start, err := time.Parse(time.DateOnly, v.StartDate)
			if err != nil {
				return nil, err
			}
			month := payment.AddDate(0, -v.PayMonthOffset, 0)
			if month.Format("2006-01") < start.Format("2006-01") {
				continue
			}
			selected = v
			salaryMonth = month
		}
		if selected == nil {
			continue
		}
		v := *selected
		last := payment.AddDate(0, 1, -1).Day()
		date := payment.AddDate(0, 0, min(v.PayDay, last)-1)
		// Preserve legacy partial-start dates for the initial month.
		if v.PayMonthOffset == 0 && date.Format(time.DateOnly) < v.StartDate {
			continue
		}
		var c *SalaryContributions
		if v.Contributions != nil {
			copy := *v.Contributions
			c = &copy
		}
		policyFrom, warning := "", ""
		if v.BaseMode == "policy" {
			if c == nil {
				return nil, errors.New("缺少缴费配置")
			}
			code := map[string]string{"beijing": "CN-BJ", "shanghai": "CN-SH"}[v.CityCode]
			if code == "" {
				code = v.CityCode
			}
			rule, err := lookup(code, salaryMonth.Format(time.DateOnly))
			if err != nil {
				if !errors.Is(err, payroll.ErrRuleNotFound) {
					return nil, err
				}
				estimated = true
				warning = "该工资月份暂无已核验政策，暂按此版本保存的基数估算；累计个税也可能受影响。"
			} else {
				c.SocialBase = math.Max(rule.SSBaseMin.InexactFloat64(), math.Min(v.PreTaxMonthly, rule.SSBaseMax.InexactFloat64()))
				c.HousingBase = math.Max(rule.HFBaseMin.InexactFloat64(), math.Min(v.PreTaxMonthly, rule.HFBaseMax.InexactFloat64()))
				policyFrom = rule.EffectiveFrom
			}
		}
		req := SalaryPreviewRequest{PreTaxAmount: v.PreTaxMonthly, CityCode: v.CityCode, StartDate: date.Format(time.DateOnly), PayDay: v.PayDay, Frequency: FrequencyYearly, Currency: v.Currency, SocialRate: &v.SocialRate, ExtraDeductions: v.ExtraDeductions, Contributions: c}
		preview, err := PreviewSalaryYear(req, year)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", salaryMonth.Format("2006-01"), err)
		}
		if len(preview.Items) != 1 {
			return nil, errors.New("无效发薪配置")
		}
		item := preview.Items[0]
		grossYTD = roundSalary(grossYTD + item.PreTax)
		// Accumulate signed deductions; a low-income month still contributes its allowance.
		taxableYTD = roundSalary(taxableYTD + item.PreTax - item.SocialSecurity - MonthlyDeduction - v.ExtraDeductions)
		tax := roundSalary(math.Max(0, CalculateTax(math.Max(0, taxableYTD))-taxPaid))
		taxPaid = roundSalary(taxPaid + tax)
		item.TaxAmount = tax
		item.CumulativeTax = taxPaid
		item.CumulativePreTax = grossYTD
		item.PostTax = roundSalary(item.PreTax - item.SocialSecurity - tax)
		item.SalaryMonth = salaryMonth.Format("2006-01")
		item.PolicyFrom = policyFrom
		if warning == "" && estimated {
			warning = "本年先前月份缺少已核验政策，累计个税为估算。"
		}
		item.Warning = warning
		out = append(out, SalaryHistoryItem{SalaryPreviewItem: item, Config: v, Configuration: c})
	}
	return out, nil
}

func HistoryPreview(items []SalaryHistoryItem, year int) SalaryPreviewResponse {
	r := SalaryPreviewResponse{Year: year, Items: []SalaryPreviewItem{}}
	for _, i := range items {
		r.Items = append(r.Items, i.SalaryPreviewItem)
		r.CityCode = i.Config.CityCode
		r.TotalPreTax += i.PreTax
		r.TotalTax += i.TaxAmount
		r.TotalPostTax += i.PostTax
	}
	r.TotalPreTax = roundSalary(r.TotalPreTax)
	r.TotalTax = roundSalary(r.TotalTax)
	r.TotalPostTax = roundSalary(r.TotalPostTax)
	return r
}
