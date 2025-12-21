package payroll

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCalculator() *Calculator {
	// 设置测试用的地区规则
	rule := &PayrollRegionRule{
		RegionCode:           "CN-DEFAULT",
		Year:                 2025,
		SSBaseMin:            decimal.NewFromInt(4000),
		SSBaseMax:            decimal.NewFromInt(30000),
		HFBaseMin:            decimal.NewFromInt(2500),
		HFBaseMax:            decimal.NewFromInt(30000),
		SSPersonRateSum:      decimal.NewFromFloat(0.105), // 10.5%
		SSCompanyRateSum:     decimal.NewFromFloat(0.295), // 29.5%
		HFRateMin:            decimal.NewFromFloat(0.05),
		HFRateMax:            decimal.NewFromFloat(0.12),
		DefaultHFRatePerson:  decimal.NewFromFloat(0.12),
		DefaultHFRateCompany: decimal.NewFromFloat(0.12),
	}

	// 设置2025年税率表
	brackets := []TaxBracketCN{
		{Year: 2025, LowerBound: decimal.NewFromInt(0), UpperBound: decPtr(36000), Rate: decimal.NewFromFloat(0.03), QuickDeduction: decimal.Zero},
		{Year: 2025, LowerBound: decimal.NewFromInt(36000), UpperBound: decPtr(144000), Rate: decimal.NewFromFloat(0.10), QuickDeduction: decimal.NewFromInt(2520)},
		{Year: 2025, LowerBound: decimal.NewFromInt(144000), UpperBound: decPtr(300000), Rate: decimal.NewFromFloat(0.20), QuickDeduction: decimal.NewFromInt(16920)},
		{Year: 2025, LowerBound: decimal.NewFromInt(300000), UpperBound: decPtr(420000), Rate: decimal.NewFromFloat(0.25), QuickDeduction: decimal.NewFromInt(31920)},
		{Year: 2025, LowerBound: decimal.NewFromInt(420000), UpperBound: decPtr(660000), Rate: decimal.NewFromFloat(0.30), QuickDeduction: decimal.NewFromInt(52920)},
		{Year: 2025, LowerBound: decimal.NewFromInt(660000), UpperBound: decPtr(960000), Rate: decimal.NewFromFloat(0.35), QuickDeduction: decimal.NewFromInt(85920)},
		{Year: 2025, LowerBound: decimal.NewFromInt(960000), UpperBound: nil, Rate: decimal.NewFromFloat(0.45), QuickDeduction: decimal.NewFromInt(181920)},
	}

	return NewCalculator(rule, brackets)
}

func decPtr(v int64) *decimal.Decimal {
	d := decimal.NewFromInt(v)
	return &d
}

// TestCalculator_NoSocialNoHousingNoSpecial 无社保无公积金无专项扣除
func TestCalculator_NoSocialNoHousingNoSpecial(t *testing.T) {
	calc := setupTestCalculator()

	req := &CalcSalaryRequest{
		RegionCode:     "CN-DEFAULT",
		Year:           2025,
		GrossMonthly:   decimal.NewFromInt(20000),
		SSEnabled:      false,
		HFEnabled:      false,
		SpecialEnabled: false,
	}

	result, err := calc.Calculate(req)
	require.NoError(t, err)

	// 验证基本结构
	assert.Equal(t, 12, len(result.Months))
	assert.True(t, result.Totals.TaxYear.GreaterThan(decimal.Zero), "年税合计应大于0")
	assert.True(t, result.Totals.NetYear.GreaterThan(decimal.Zero), "年到手应大于0")

	// 验证累计预扣：税应该逐月变化（不是每月固定）
	// 月应纳税所得额 = 20000 - 5000 = 15000
	// 第1月累计应纳税所得额 = 15000，税 = 15000 * 3% = 450
	// 第2月累计应纳税所得额 = 30000，累计税 = 30000 * 3% = 900，本月税 = 900 - 450 = 450
	// 第3月累计应纳税所得额 = 45000，累计税 = 45000 * 10% - 2520 = 1980，本月税 = 1980 - 900 = 1080
	assert.Equal(t, "450", result.Months[0].Tax.String(), "第1月税")
	assert.Equal(t, "450", result.Months[1].Tax.String(), "第2月税")
	assert.Equal(t, "1080", result.Months[2].Tax.String(), "第3月税应跳档")

	// 验证社保公积金为0
	for _, month := range result.Months {
		assert.True(t, month.SSPerson.IsZero(), "社保应为0")
		assert.True(t, month.HFPerson.IsZero(), "公积金应为0")
	}

	// 验证 summary
	assert.True(t, result.Summary.SSPersonMonthly.IsZero())
	assert.True(t, result.Summary.HFPersonMonthly.IsZero())
	assert.True(t, result.Summary.SpecialMonthly.IsZero())
}

// TestCalculator_WithSocialAndHousingAndSpecial 有社保公积金专项扣除
func TestCalculator_WithSocialAndHousingAndSpecial(t *testing.T) {
	calc := setupTestCalculator()

	req := &CalcSalaryRequest{
		RegionCode:     "CN-DEFAULT",
		Year:           2025,
		GrossMonthly:   decimal.NewFromInt(20000),
		SSEnabled:      true,
		SSBaseMode:     BaseModeByWage,
		HFEnabled:      true,
		HFBaseMode:     BaseModeByWage,
		HFRatePerson:   decimal.NewFromFloat(0.12),
		HFRateCompany:  decimal.NewFromFloat(0.12),
		SpecialEnabled: true,
		Special: SpecialDeduction{
			ChildEdu:       decimal.NewFromInt(2000),
			ElderlyCare:    decimal.NewFromInt(2000),
			PensionMonthly: decimal.NewFromInt(1000),
		},
	}

	result, err := calc.Calculate(req)
	require.NoError(t, err)

	// 社保 = 20000 * 10.5% = 2100
	assert.Equal(t, "2100", result.Summary.SSPersonMonthly.String())
	// 公积金 = 20000 * 12% = 2400
	assert.Equal(t, "2400", result.Summary.HFPersonMonthly.String())
	// 专项扣除 = 2000 + 2000 = 4000
	assert.Equal(t, "4000", result.Summary.SpecialMonthly.String())

	// 月应纳税所得额 = 20000 - 5000 - 2100 - 2400 - 4000 - 1000 = 5500
	// 比无社保公积金时的15000低很多，税应该更低
	assert.True(t, result.Months[0].Tax.LessThan(decimal.NewFromInt(450)), "有扣除时税应该更低")

	// 到手应该合理
	assert.True(t, result.Months[0].Net.GreaterThan(decimal.NewFromInt(10000)), "到手应大于10000")
}

// TestCalculator_CustomBaseClamp 自定义基数clamp行为
func TestCalculator_CustomBaseClamp(t *testing.T) {
	calc := setupTestCalculator()

	// 测试基数低于最小值时clamp到最小值
	customBase := decimal.NewFromInt(3000) // 低于 SSBaseMin 4000
	req := &CalcSalaryRequest{
		RegionCode:   "CN-DEFAULT",
		Year:         2025,
		GrossMonthly: decimal.NewFromInt(20000),
		SSEnabled:    true,
		SSBaseMode:   BaseModeCustom,
		SSBaseCustom: &customBase,
		HFEnabled:    false,
	}

	result, err := calc.Calculate(req)
	require.NoError(t, err)

	// 社保基数应该被clamp到4000，社保 = 4000 * 10.5% = 420
	assert.Equal(t, "420", result.Summary.SSPersonMonthly.String())

	// 测试基数高于最大值时clamp到最大值
	highBase := decimal.NewFromInt(50000) // 高于 SSBaseMax 30000
	req.SSBaseCustom = &highBase

	result, err = calc.Calculate(req)
	require.NoError(t, err)

	// 社保基数应该被clamp到30000，社保 = 30000 * 10.5% = 3150
	assert.Equal(t, "3150", result.Summary.SSPersonMonthly.String())
}

// TestCalculator_TaxBracketBoundary 税率表边界测试
func TestCalculator_TaxBracketBoundary(t *testing.T) {
	calc := setupTestCalculator()

	// 测试累计应纳税所得额刚好在边界
	// 设置月薪使得12个月累计应纳税所得额刚好在某个边界
	// 如累计 36000 边界：36000 / 12 = 3000 每月应纳税所得额
	// 月薪 = 3000 + 5000（基本减除）= 8000

	req := &CalcSalaryRequest{
		RegionCode:     "CN-DEFAULT",
		Year:           2025,
		GrossMonthly:   decimal.NewFromInt(8000),
		SSEnabled:      false,
		HFEnabled:      false,
		SpecialEnabled: false,
	}

	result, err := calc.Calculate(req)
	require.NoError(t, err)

	// 月应纳税所得额 = 8000 - 5000 = 3000
	// 第12月累计应纳税所得额 = 3000 * 12 = 36000，刚好在边界
	// 累计税 = 36000 * 3% = 1080
	assert.Equal(t, "1080", result.Totals.TaxYear.String(), "年税应为1080")

	// 测试超过边界一点点
	req.GrossMonthly = decimal.NewFromInt(8100)
	result, err = calc.Calculate(req)
	require.NoError(t, err)

	// 月应纳税所得额 = 8100 - 5000 = 3100
	// 第12月累计应纳税所得额 = 3100 * 12 = 37200，超过36000边界
	// 累计税 = 37200 * 10% - 2520 = 1200
	assert.Equal(t, "1200", result.Totals.TaxYear.String(), "年税应为1200")
}

// TestCalculator_ValidationErrors 校验错误
func TestCalculator_ValidationErrors(t *testing.T) {
	calc := setupTestCalculator()

	// 测试月薪为0
	req := &CalcSalaryRequest{
		RegionCode:   "CN-DEFAULT",
		Year:         2025,
		GrossMonthly: decimal.Zero,
	}
	_, err := calc.Calculate(req)
	assert.ErrorIs(t, err, ErrInvalidGrossMonthly)

	// 测试CUSTOM模式缺少自定义基数
	req = &CalcSalaryRequest{
		RegionCode:   "CN-DEFAULT",
		Year:         2025,
		GrossMonthly: decimal.NewFromInt(20000),
		SSEnabled:    true,
		SSBaseMode:   BaseModeCustom,
		SSBaseCustom: nil,
	}
	_, err = calc.Calculate(req)
	assert.ErrorIs(t, err, ErrCustomBaseRequired)

	// 测试公积金比例超出范围（>100%）
	req = &CalcSalaryRequest{
		RegionCode:    "CN-DEFAULT",
		Year:          2025,
		GrossMonthly:  decimal.NewFromInt(20000),
		HFEnabled:     true,
		HFBaseMode:    BaseModeByWage,
		HFRatePerson:  decimal.NewFromFloat(1.5), // 150% 超出0-100%范围
		HFRateCompany: decimal.NewFromFloat(0.12),
	}
	_, err = calc.Calculate(req)
	assert.ErrorIs(t, err, ErrHFRateOutOfRange)

	// 测试社保比例超出范围
	customSSRate := decimal.NewFromFloat(1.5) // 150%
	req = &CalcSalaryRequest{
		RegionCode:      "CN-DEFAULT",
		Year:            2025,
		GrossMonthly:    decimal.NewFromInt(20000),
		SSEnabled:       true,
		SSBaseMode:      BaseModeByWage,
		SSRatePersonSum: &customSSRate,
	}
	_, err = calc.Calculate(req)
	assert.ErrorIs(t, err, ErrSSRateOutOfRange)
}

// TestCalculator_ByMinMode 按最低基数模式
func TestCalculator_ByMinMode(t *testing.T) {
	calc := setupTestCalculator()

	req := &CalcSalaryRequest{
		RegionCode:    "CN-DEFAULT",
		Year:          2025,
		GrossMonthly:  decimal.NewFromInt(50000), // 高薪
		SSEnabled:     true,
		SSBaseMode:    BaseModeByMin, // 按最低基数
		HFEnabled:     true,
		HFBaseMode:    BaseModeByMin, // 按最低基数
		HFRatePerson:  decimal.NewFromFloat(0.12),
		HFRateCompany: decimal.NewFromFloat(0.12),
	}

	result, err := calc.Calculate(req)
	require.NoError(t, err)

	// 社保按最低基数 4000，社保 = 4000 * 10.5% = 420
	assert.Equal(t, "420", result.Summary.SSPersonMonthly.String())
	// 公积金按最低基数 2500，公积金 = 2500 * 12% = 300
	assert.Equal(t, "300", result.Summary.HFPersonMonthly.String())
}
