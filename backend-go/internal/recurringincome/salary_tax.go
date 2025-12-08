package recurringincome

import (
	"math"
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

// 城市社保个人部分比例（简化配置）
var citySocialRate = map[string]float64{
	"beijing":   0.12,
	"shanghai":  0.10,
	"guangzhou": 0.11,
	"shenzhen":  0.10,
	"default":   0.11,
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

// GetSocialSecurityRate 获取城市社保个人部分比例（导出供其他包使用）
func GetSocialSecurityRate(cityCode string) float64 {
	if rate, ok := citySocialRate[cityCode]; ok {
		return rate
	}
	return citySocialRate["default"]
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
func PreviewSalaryToYearEnd(req SalaryPreviewRequest) (SalaryPreviewResponse, error) {
	// 解析开始日期
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return SalaryPreviewResponse{}, err
	}

	year := startDate.Year()
	yearEnd := time.Date(year, 12, 31, 0, 0, 0, 0, startDate.Location())

	// 获取社保比例
	socialRate := GetSocialSecurityRate(req.CityCode)

	// 计算发薪日期列表
	var payDates []time.Time
	currentDate := startDate

	// 根据频率生成发薪日期
	for !currentDate.After(yearEnd) {
		// 设置发薪日
		payDate := time.Date(currentDate.Year(), currentDate.Month(), req.PayDay, 0, 0, 0, 0, currentDate.Location())
		// 如果发薪日小于当前日期，则使用当前日期所在月份的发薪日
		if payDate.Before(currentDate) {
			payDate = AddMonths(payDate, 1)
		}

		if payDate.After(yearEnd) {
			break
		}

		payDates = append(payDates, payDate)

		// 计算下一个发薪日期
		switch req.Frequency {
		case FrequencyMonthly:
			currentDate = AddMonths(currentDate, 1)
		case FrequencyBimonthly:
			currentDate = AddMonths(currentDate, 2)
		case FrequencyQuarterly:
			currentDate = AddMonths(currentDate, 3)
		case FrequencySemiannual:
			currentDate = AddMonths(currentDate, 6)
		case FrequencyWeekly:
			currentDate = currentDate.AddDate(0, 0, 7)
		case FrequencyYearly:
			currentDate = currentDate.AddDate(1, 0, 0)
		default:
			currentDate = AddMonths(currentDate, 1)
		}
	}

	// 计算每次发薪的明细
	var items []SalaryPreviewItem
	var cumulativePreTax float64
	var cumulativeTax float64
	var cumulativeTaxableIncome float64

	for _, payDate := range payDates {
		preTax := req.PreTaxAmount
		socialSecurity := preTax * socialRate
		taxableIncome := preTax - socialSecurity - MonthlyDeduction

		// 累计计算
		cumulativePreTax += preTax
		cumulativeTaxableIncome += taxableIncome

		// 计算累计个税
		cumulativeTaxTotal := CalculateTax(cumulativeTaxableIncome)
		// 本次应缴个税 = 累计个税 - 已缴个税
		currentTax := cumulativeTaxTotal - cumulativeTax
		if currentTax < 0 {
			currentTax = 0
		}
		cumulativeTax = cumulativeTaxTotal

		// 实际到手
		postTax := preTax - socialSecurity - currentTax

		items = append(items, SalaryPreviewItem{
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
