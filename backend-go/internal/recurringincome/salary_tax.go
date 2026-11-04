package recurringincome

import (
	"errors"
	"math"
	"neverbroke/internal/citysocialrate"
	"time"
)

// taxBracket 个税税率档次
type taxBracket struct {
	Threshold   float64 // 累计应纳税所得额阈值
	Rate        float64 // 税率
	QuickDeduct float64 // 速算扣除数
}

// 中国个税年度税率表（2023年现行规则）
var annualTaxBrackets = []taxBracket{
	{Threshold: 36000, Rate: 0.03, QuickDeduct: 0},
	{Threshold: 144000, Rate: 0.10, QuickDeduct: 2520},
	{Threshold: 300000, Rate: 0.20, QuickDeduct: 16920},
	{Threshold: 420000, Rate: 0.25, QuickDeduct: 31920},
	{Threshold: 660000, Rate: 0.30, QuickDeduct: 52920},
	{Threshold: 960000, Rate: 0.35, QuickDeduct: 85920},
	{Threshold: math.MaxFloat64, Rate: 0.45, QuickDeduct: 181920},
}

// MonthlyDeduction 基本免税额（每月，导出供其他包使用）
const MonthlyDeduction = 5000.0

// CalculateTax 计算累计应纳税所得额对应的个税（导出供其他包使用）
func CalculateTax(cumulativeTaxableIncome float64) float64 {
	for _, bracket := range annualTaxBrackets {
		if cumulativeTaxableIncome <= bracket.Threshold {
			return cumulativeTaxableIncome*bracket.Rate - bracket.QuickDeduct
		}
	}
	// 理论上不会到这里，但为了安全
	return cumulativeTaxableIncome*0.45 - 181920
}

// GetSocialSecurityRate 获取城市社保个人部分比例（从 citysocialrate 模块获取）
func GetSocialSecurityRate(cityCode string) float64 {
	rate := citysocialrate.GetActiveByCityCode(cityCode)
	if rate != nil {
		return rate.TotalRate
	}
	// 如果没找到，返回默认值
	defaultRate := citysocialrate.GetActiveByCityCode("default")
	if defaultRate != nil {
		return defaultRate.TotalRate
	}
	return 0.11 // 兜底默认值
}

// AddMonths 添加月份（处理月末边界情况，导出供其他包使用）
func AddMonths(t time.Time, months int) time.Time {
	result := t.AddDate(0, months, 0)
	// 如果原日期是月末，保持月末
	if t.Day() != result.Day() && t.Day() > 28 {
		// 原日期可能是月末，调整到新月份的月末
		lastDay := time.Date(result.Year(), result.Month()+1, 0, 0, 0, 0, 0, result.Location()).Day()
		if t.Day() > lastDay {
			result = time.Date(result.Year(), result.Month(), lastDay, 0, 0, 0, 0, result.Location())
		}
	}
	return result
}

// PreviewSalaryToYearEnd 预览工资到年底的明细（按中国个税规则）
// 支持自定义社保比例和专项附加扣除
func PreviewSalaryToYearEnd(req SalaryPreviewRequest) (SalaryPreviewResponse, error) {
	// 解析开始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return SalaryPreviewResponse{}, err
	}

	result, err := PreviewSalaryYear(req, startDate.Year())
	if err == nil && len(result.Items) == 0 {
		return PreviewSalaryYear(req, startDate.Year()+1)
	}
	return result, err
}

func roundSalary(value float64) float64 { return math.Round(value*100) / 100 }

// PreviewSalaryYear preserves the plan's original cadence and resets cumulative tax each year.
func PreviewSalaryYear(req SalaryPreviewRequest, year int) (SalaryPreviewResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return SalaryPreviewResponse{}, err
	}
	if req.PayDay < 1 || req.PayDay > 31 || req.PreTaxAmount <= 0 || req.ExtraDeductions < 0 {
		return SalaryPreviewResponse{}, errors.New("invalid salary configuration")
	}
	socialRate := GetSocialSecurityRate(req.CityCode)
	if req.SocialRate != nil {
		socialRate = *req.SocialRate
	}
	if socialRate < 0 || socialRate > 1 {
		return SalaryPreviewResponse{}, errors.New("invalid social rate")
	}
	var contributions *SalaryContributionAmounts
	socialSecurity := roundSalary(req.PreTaxAmount * socialRate)
	if c := req.Contributions; c != nil {
		if c.SocialBase < 0 || c.HousingBase < 0 {
			return SalaryPreviewResponse{}, errors.New("invalid contribution base")
		}
		for _, rate := range []float64{c.PensionRate, c.MedicalRate, c.UnemploymentRate, c.HousingRate} {
			if rate < 0 || rate > 1 {
				return SalaryPreviewResponse{}, errors.New("invalid contribution rate")
			}
		}
		contributions = &SalaryContributionAmounts{
			Pension: roundSalary(c.SocialBase * c.PensionRate), Medical: roundSalary(c.SocialBase * c.MedicalRate),
			Unemployment: roundSalary(c.SocialBase * c.UnemploymentRate), Housing: roundSalary(c.HousingBase * c.HousingRate),
		}
		socialSecurity = roundSalary(contributions.Pension + contributions.Medical + contributions.Unemployment + contributions.Housing)
	}
	if socialSecurity > req.PreTaxAmount {
		return SalaryPreviewResponse{}, errors.New("contributions exceed salary")
	}
	step := 1
	switch req.Frequency {
	case FrequencyMonthly, "", FrequencyWeekly:
	case FrequencyBimonthly:
		step = 2
	case FrequencyQuarterly:
		step = 3
	case FrequencySemiannual:
		step = 6
	case FrequencyYearly:
		step = 12
	default:
		return SalaryPreviewResponse{}, errors.New("invalid salary frequency")
	}
	payDates := []time.Time{}
	yearEnd := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	if req.Frequency == FrequencyWeekly {
		for date := startDate; date.Before(yearEnd); date = date.AddDate(0, 0, 7) {
			if date.Year() == year {
				payDates = append(payDates, date)
			}
		}
	} else {
		month := time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		// Clamp before advancing so February and 30-day months never overflow.
		payDate := func(month time.Time) time.Time {
			last := month.AddDate(0, 1, -1).Day()
			return month.AddDate(0, 0, min(req.PayDay, last)-1)
		}
		if payDate(month).Before(startDate) {
			month = month.AddDate(0, 1, 0)
		}
		for ; month.Before(yearEnd); month = month.AddDate(0, step, 0) {
			if month.Year() == year {
				payDates = append(payDates, payDate(month))
			}
		}
	}

	// 计算每次发薪的明细
	items := []SalaryPreviewItem{}
	var cumulativePreTax float64
	var cumulativeTax float64
	var cumulativeTaxableIncome float64

	for _, payDate := range payDates {
		preTax := roundSalary(req.PreTaxAmount)
		// 应纳税所得额 = 税前 - 社保 - 基本免税额 - 专项附加扣除
		taxableIncome := roundSalary(preTax - socialSecurity - MonthlyDeduction - req.ExtraDeductions)
		if taxableIncome < 0 {
			taxableIncome = 0
		}

		// 累计计算
		cumulativePreTax += preTax
		cumulativeTaxableIncome += taxableIncome

		// 计算累计个税
		cumulativeTaxTotal := roundSalary(CalculateTax(cumulativeTaxableIncome))
		// 本次应缴个税 = 累计个税 - 已缴个税
		currentTax := roundSalary(cumulativeTaxTotal - cumulativeTax)
		if currentTax < 0 {
			currentTax = 0
		}
		cumulativeTax = cumulativeTaxTotal

		// 实际到手
		postTax := roundSalary(preTax - socialSecurity - currentTax)

		items = append(items, SalaryPreviewItem{
			Contributions:    contributions,
			Date:             payDate.Format("2006-01-02"),
			PreTax:           preTax,
			SocialSecurity:   socialSecurity,
			TaxableIncome:    taxableIncome,
			TaxAmount:        currentTax,
			PostTax:          postTax,
			CumulativePreTax: cumulativePreTax,
			CumulativeTax:    cumulativeTax,
		})
	}

	// 计算汇总
	var totalPreTax, totalTax, totalPostTax float64
	for _, item := range items {
		totalPreTax += item.PreTax
		totalTax += item.TaxAmount
		totalPostTax += item.PostTax
	}

	return SalaryPreviewResponse{
		CityCode:     req.CityCode,
		Year:         year,
		Items:        items,
		TotalPreTax:  totalPreTax,
		TotalTax:     totalTax,
		TotalPostTax: totalPostTax,
	}, nil
}
