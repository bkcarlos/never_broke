package payroll

import (
	"github.com/shopspring/decimal"
)

// Calculator 薪资计算器
type Calculator struct {
	rule     *PayrollRegionRule
	brackets []TaxBracketCN
}

// NewCalculator 创建计算器
func NewCalculator(rule *PayrollRegionRule, brackets []TaxBracketCN) *Calculator {
	return &Calculator{
		rule:     rule,
		brackets: brackets,
	}
}

// Calculate 执行薪资计算
func (c *Calculator) Calculate(req *CalcSalaryRequest) (*CalcSalaryResponse, error) {
	// 校验
	if err := c.validate(req); err != nil {
		return nil, err
	}

	// 计算基数
	ssBase := c.calculateSSBase(req)
	hfBase := c.calculateHFBase(req)

	// 获取社保比例（优先使用自定义，否则使用地区默认）
	ssPersonRate := c.rule.SSPersonRateSum
	ssCompanyRate := c.rule.SSCompanyRateSum
	if req.SSRatePersonSum != nil {
		ssPersonRate = *req.SSRatePersonSum
	}
	if req.SSRateCompanySum != nil {
		ssCompanyRate = *req.SSRateCompanySum
	}

	// 计算每月金额
	ssPerson := decimal.Zero
	ssCompany := decimal.Zero
	hfPerson := decimal.Zero
	hfCompany := decimal.Zero

	if req.SSEnabled {
		ssPerson = ssBase.Mul(ssPersonRate).Round(2)
		ssCompany = ssBase.Mul(ssCompanyRate).Round(2)
	}

	if req.HFEnabled {
		hfPerson = hfBase.Mul(req.HFRatePerson).Round(2)
		hfCompany = hfBase.Mul(req.HFRateCompany).Round(2)
	}

	// 专项附加扣除
	specialMonthly := decimal.Zero
	if req.SpecialEnabled {
		specialMonthly = c.calculateSpecialDeduction(&req.Special)
	}

	// 基本减除费用
	basicDeduction := decimal.NewFromInt(5000)

	// 累计预扣法计算12个月
	months := make([]CalcSalaryMonth, 12)
	cumIncome := decimal.Zero
	cumDeduction := decimal.Zero
	taxPaidSoFar := decimal.Zero

	for m := 1; m <= 12; m++ {
		// 累计收入
		cumIncome = cumIncome.Add(req.GrossMonthly)

		// 累计扣除 = 基本减除 + 社保 + 公积金 + 专项附加 + 个人养老金
		monthlyDeduction := basicDeduction.Add(ssPerson).Add(hfPerson).Add(specialMonthly)
		if req.SpecialEnabled {
			monthlyDeduction = monthlyDeduction.Add(req.Special.PensionMonthly)
		}
		cumDeduction = cumDeduction.Add(monthlyDeduction)

		// 累计应纳税所得额
		cumTaxable := cumIncome.Sub(cumDeduction)
		if cumTaxable.LessThan(decimal.Zero) {
			cumTaxable = decimal.Zero
		}

		// 累计应纳税额
		cumTax := c.calculateBracketTax(cumTaxable)

		// 本月应纳税 = 累计应纳税 - 已缴税
		taxThisMonth := cumTax.Sub(taxPaidSoFar)
		if taxThisMonth.LessThan(decimal.Zero) {
			taxThisMonth = decimal.Zero
		}
		taxPaidSoFar = cumTax

		// 本月到手 = 税前 - 社保 - 公积金 - 个税
		net := req.GrossMonthly.Sub(ssPerson).Sub(hfPerson).Sub(taxThisMonth)

		months[m-1] = CalcSalaryMonth{
			Month:    m,
			Gross:    req.GrossMonthly,
			SSPerson: ssPerson,
			HFPerson: hfPerson,
			Tax:      taxThisMonth.Round(2),
			Net:      net.Round(2),
		}
	}

	// 计算年度合计
	netYear := decimal.Zero
	taxYear := decimal.Zero
	for _, month := range months {
		netYear = netYear.Add(month.Net)
		taxYear = taxYear.Add(month.Tax)
	}

	// 公司成本
	companyCost := req.GrossMonthly.Add(ssCompany).Add(hfCompany)

	return &CalcSalaryResponse{
		Summary: CalcSalarySummary{
			GrossMonthly:       req.GrossMonthly,
			SSPersonMonthly:    ssPerson,
			SSCompanyMonthly:   ssCompany,
			HFPersonMonthly:    hfPerson,
			HFCompanyMonthly:   hfCompany,
			SpecialMonthly:     specialMonthly,
			CompanyCostMonthly: companyCost.Round(2),
		},
		Months: months,
		Totals: CalcSalaryTotals{
			NetYear: netYear.Round(2),
			TaxYear: taxYear.Round(2),
		},
	}, nil
}

// validate 校验请求
func (c *Calculator) validate(req *CalcSalaryRequest) error {
	// 校验月薪
	if req.GrossMonthly.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidGrossMonthly
	}

	// 校验社保自定义基数
	if req.SSEnabled && req.SSBaseMode == BaseModeCustom {
		if req.SSBaseCustom == nil {
			return ErrCustomBaseRequired
		}
		if req.SSBaseCustom.LessThanOrEqual(decimal.Zero) {
			return ErrCustomBaseMustBePositive
		}
	}

	// 校验公积金自定义基数
	if req.HFEnabled && req.HFBaseMode == BaseModeCustom {
		if req.HFBaseCustom == nil {
			return ErrCustomBaseRequired
		}
		if req.HFBaseCustom.LessThanOrEqual(decimal.Zero) {
			return ErrCustomBaseMustBePositive
		}
	}

	// 校验公积金比例（放宽限制，允许0-100%）
	if req.HFEnabled {
		if req.HFRatePerson.LessThan(decimal.Zero) || req.HFRatePerson.GreaterThan(decimal.NewFromInt(1)) {
			return ErrHFRateOutOfRange
		}
		if req.HFRateCompany.LessThan(decimal.Zero) || req.HFRateCompany.GreaterThan(decimal.NewFromInt(1)) {
			return ErrHFRateOutOfRange
		}
	}

	// 校验社保比例（如果自定义）
	if req.SSEnabled && req.SSRatePersonSum != nil {
		if req.SSRatePersonSum.LessThan(decimal.Zero) || req.SSRatePersonSum.GreaterThan(decimal.NewFromInt(1)) {
			return ErrSSRateOutOfRange
		}
	}
	if req.SSEnabled && req.SSRateCompanySum != nil {
		if req.SSRateCompanySum.LessThan(decimal.Zero) || req.SSRateCompanySum.GreaterThan(decimal.NewFromInt(1)) {
			return ErrSSRateOutOfRange
		}
	}

	return nil
}

// calculateSSBase 计算社保基数
func (c *Calculator) calculateSSBase(req *CalcSalaryRequest) decimal.Decimal {
	if !req.SSEnabled {
		return decimal.Zero
	}

	switch req.SSBaseMode {
	case BaseModeByMin:
		return c.rule.SSBaseMin
	case BaseModeCustom:
		return c.clamp(*req.SSBaseCustom, c.rule.SSBaseMin, c.rule.SSBaseMax)
	default: // BY_WAGE
		return c.clamp(req.GrossMonthly, c.rule.SSBaseMin, c.rule.SSBaseMax)
	}
}

// calculateHFBase 计算公积金基数
func (c *Calculator) calculateHFBase(req *CalcSalaryRequest) decimal.Decimal {
	if !req.HFEnabled {
		return decimal.Zero
	}

	switch req.HFBaseMode {
	case BaseModeByMin:
		return c.rule.HFBaseMin
	case BaseModeCustom:
		return c.clamp(*req.HFBaseCustom, c.rule.HFBaseMin, c.rule.HFBaseMax)
	default: // BY_WAGE
		return c.clamp(req.GrossMonthly, c.rule.HFBaseMin, c.rule.HFBaseMax)
	}
}

// clamp 限制值在范围内
func (c *Calculator) clamp(value, min, max decimal.Decimal) decimal.Decimal {
	if value.LessThan(min) {
		return min
	}
	if value.GreaterThan(max) {
		return max
	}
	return value
}

// calculateSpecialDeduction 计算专项附加扣除
func (c *Calculator) calculateSpecialDeduction(special *SpecialDeduction) decimal.Decimal {
	total := decimal.Zero
	total = total.Add(special.ChildEdu)
	total = total.Add(special.ContEdu)
	total = total.Add(special.ElderlyCare)
	total = total.Add(special.HousingLoan)
	total = total.Add(special.HousingRent)
	total = total.Add(special.InfantCare)
	// PensionMonthly 单独计算，不在这里累加
	return total
}

// calculateBracketTax 根据税率表计算税额
func (c *Calculator) calculateBracketTax(taxable decimal.Decimal) decimal.Decimal {
	if taxable.LessThanOrEqual(decimal.Zero) {
		return decimal.Zero
	}

	for _, bracket := range c.brackets {
		// 如果 upperBound 为 nil（无上限）或 taxable <= upperBound
		if bracket.UpperBound == nil || taxable.LessThanOrEqual(*bracket.UpperBound) {
			// 税额 = 应纳税所得额 * 税率 - 速算扣除数
			tax := taxable.Mul(bracket.Rate).Sub(bracket.QuickDeduction)
			if tax.LessThan(decimal.Zero) {
				return decimal.Zero
			}
			return tax.Round(2)
		}
	}

	// 兜底：最高档45%
	return taxable.Mul(decimal.NewFromFloat(0.45)).Sub(decimal.NewFromInt(181920)).Round(2)
}
