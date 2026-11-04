package payroll

import "github.com/shopspring/decimal"

// BaseMode 基数模式
type BaseMode string

const (
	BaseModeByWage BaseMode = "BY_WAGE"
	BaseModeByMin  BaseMode = "BY_MIN"
	BaseModeCustom BaseMode = "CUSTOM"
)

// RegionListItem 地区列表项
type RegionListItem struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// RegionRuleResponse 地区规则响应
type RegionRuleResponse struct {
	EffectiveFrom        string          `json:"effectiveFrom"`
	EffectiveThrough     string          `json:"effectiveThrough"`
	AsOf                 string          `json:"asOf"`
	VerifiedAt           string          `json:"verifiedAt"`
	Sources              []RuleSource    `json:"sources"`
	RegionCode           string          `json:"regionCode"`
	Year                 int             `json:"year"`
	SSBaseMin            decimal.Decimal `json:"ssBaseMin"`
	SSBaseMax            decimal.Decimal `json:"ssBaseMax"`
	HFBaseMin            decimal.Decimal `json:"hfBaseMin"`
	HFBaseMax            decimal.Decimal `json:"hfBaseMax"`
	SSPersonRateSum      decimal.Decimal `json:"ssPersonRateSum"`
	SSCompanyRateSum     decimal.Decimal `json:"ssCompanyRateSum"`
	HFRateMin            decimal.Decimal `json:"hfRateMin"`
	HFRateMax            decimal.Decimal `json:"hfRateMax"`
	DefaultHFRatePerson  decimal.Decimal `json:"defaultHfRatePerson"`
	DefaultHFRateCompany decimal.Decimal `json:"defaultHfRateCompany"`
}

// SpecialDeduction 专项附加扣除
type SpecialDeduction struct {
	ChildEdu       decimal.Decimal `json:"childEdu"`       // 子女教育
	ContEdu        decimal.Decimal `json:"contEdu"`        // 继续教育
	ElderlyCare    decimal.Decimal `json:"elderlyCare"`    // 赡养老人
	HousingLoan    decimal.Decimal `json:"housingLoan"`    // 住房贷款利息
	HousingRent    decimal.Decimal `json:"housingRent"`    // 住房租金
	InfantCare     decimal.Decimal `json:"infantCare"`     // 3岁以下婴幼儿照护
	PensionMonthly decimal.Decimal `json:"pensionMonthly"` // 个人养老金（每月）
}

// SSRateDetail 社保各险种比例明细
type SSRateDetail struct {
	Pension      decimal.Decimal `json:"pension"`      // 养老保险
	Medical      decimal.Decimal `json:"medical"`      // 医疗保险
	Unemployment decimal.Decimal `json:"unemployment"` // 失业保险
	Injury       decimal.Decimal `json:"injury"`       // 工伤保险（仅公司缴纳）
	Maternity    decimal.Decimal `json:"maternity"`    // 生育保险（仅公司缴纳）
}

// CalcSalaryRequest 薪资计算请求
type CalcSalaryRequest struct {
	RuleDate     string          `json:"ruleDate,omitempty"`
	RegionCode   string          `json:"regionCode" binding:"required"`
	Year         int             `json:"year" binding:"required,min=2020,max=2030"`
	GrossMonthly decimal.Decimal `json:"grossMonthly" binding:"required"`

	// 社保配置
	SSEnabled           bool             `json:"ssEnabled"`
	SSBaseMode          BaseMode         `json:"ssBaseMode"`
	SSBaseCustom        *decimal.Decimal `json:"ssBaseCustom"`
	SSRatePersonSum     *decimal.Decimal `json:"ssRatePersonSum"`     // 自定义个人社保合计比例
	SSRateCompanySum    *decimal.Decimal `json:"ssRateCompanySum"`    // 自定义公司社保合计比例
	SSRatePersonDetail  *SSRateDetail    `json:"ssRatePersonDetail"`  // 个人社保明细（可选）
	SSRateCompanyDetail *SSRateDetail    `json:"ssRateCompanyDetail"` // 公司社保明细（可选）

	// 公积金配置
	HFEnabled     bool             `json:"hfEnabled"`
	HFBaseMode    BaseMode         `json:"hfBaseMode"`
	HFBaseCustom  *decimal.Decimal `json:"hfBaseCustom"`
	HFRatePerson  decimal.Decimal  `json:"hfRatePerson"`
	HFRateCompany decimal.Decimal  `json:"hfRateCompany"`

	// 专项附加扣除
	SpecialEnabled bool             `json:"specialEnabled"`
	Special        SpecialDeduction `json:"special"`
}

// CalcSalaryResponse 薪资计算响应
type CalcSalaryResponse struct {
	Summary CalcSalarySummary `json:"summary"`
	Months  []CalcSalaryMonth `json:"months"`
	Totals  CalcSalaryTotals  `json:"totals"`
}

// CalcSalarySummary 薪资概览
type CalcSalarySummary struct {
	GrossMonthly       decimal.Decimal `json:"grossMonthly"`
	SSPersonMonthly    decimal.Decimal `json:"ssPersonMonthly"`
	SSCompanyMonthly   decimal.Decimal `json:"ssCompanyMonthly"`
	HFPersonMonthly    decimal.Decimal `json:"hfPersonMonthly"`
	HFCompanyMonthly   decimal.Decimal `json:"hfCompanyMonthly"`
	SpecialMonthly     decimal.Decimal `json:"specialMonthly"`
	CompanyCostMonthly decimal.Decimal `json:"companyCostMonthly"`
}

// CalcSalaryMonth 每月明细
type CalcSalaryMonth struct {
	Month    int             `json:"month"`
	Gross    decimal.Decimal `json:"gross"`
	SSPerson decimal.Decimal `json:"ssPerson"`
	HFPerson decimal.Decimal `json:"hfPerson"`
	Tax      decimal.Decimal `json:"tax"`
	Net      decimal.Decimal `json:"net"`
}

// CalcSalaryTotals 年度合计
type CalcSalaryTotals struct {
	NetYear decimal.Decimal `json:"netYear"`
	TaxYear decimal.Decimal `json:"taxYear"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
