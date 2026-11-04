package payroll

import (
	"fmt"
	"github.com/shopspring/decimal"
)

// Service 薪资计算服务
type Service struct {
	repo *GormRepository
}

// NewService 创建服务
func NewService(repo *GormRepository) *Service {
	return &Service{repo: repo}
}

// ListRegions 获取地区列表
func (s *Service) ListRegions() ([]RegionListItem, error) {
	return []RegionListItem{{Code: "CN-BJ", Name: "北京"}, {Code: "CN-SH", Name: "上海"}}, nil
}

// GetRegionRule preserves the legacy year query, with the exact selected date in the response.
func (s *Service) GetRegionRule(regionCode string, year int) (*RegionRuleResponse, error) {
	return s.GetRegionRuleAt(regionCode, fmt.Sprintf("%04d-12-01", year))
}

func (s *Service) GetRegionRuleAt(regionCode, date string) (*RegionRuleResponse, error) {
	return findOfficialRule(regionCode, date)
}

// CalculateSalary 计算薪资
func (s *Service) CalculateSalary(req *CalcSalaryRequest) (*CalcSalaryResponse, error) {
	date := req.RuleDate
	if date == "" {
		date = fmt.Sprintf("%04d-12-01", req.Year)
	}
	verified, err := s.GetRegionRuleAt(req.RegionCode, date)
	if err != nil {
		return nil, err
	}
	rule := &PayrollRegionRule{
		RegionCode: verified.RegionCode, Year: verified.Year,
		SSBaseMin: verified.SSBaseMin, SSBaseMax: verified.SSBaseMax, HFBaseMin: verified.HFBaseMin, HFBaseMax: verified.HFBaseMax,
		SSPersonRateSum: verified.SSPersonRateSum, SSCompanyRateSum: verified.SSCompanyRateSum, HFRateMin: verified.HFRateMin, HFRateMax: verified.HFRateMax,
	}
	// Keep the existing comprehensive-income bracket schedule available when
	// the optional legacy payroll seed tables were never installed.
	brackets := standardTaxBrackets(req.Year)

	// 创建计算器并执行计算
	calculator := NewCalculator(rule, brackets)
	return calculator.Calculate(req)
}

// Same seven brackets used by the existing salary preview and migration 0008.
func standardTaxBrackets(year int) []TaxBracketCN {
	limits := []int64{36000, 144000, 300000, 420000, 660000, 960000}
	rates := []float64{.03, .10, .20, .25, .30, .35, .45}
	quick := []int64{0, 2520, 16920, 31920, 52920, 85920, 181920}
	result := make([]TaxBracketCN, 7)
	for i := range result {
		result[i] = TaxBracketCN{Year: year, Rate: decimal.NewFromFloat(rates[i]), QuickDeduction: decimal.NewFromInt(quick[i])}
		if i < len(limits) {
			upper := decimal.NewFromInt(limits[i])
			result[i].UpperBound = &upper
		}
	}
	return result
}
