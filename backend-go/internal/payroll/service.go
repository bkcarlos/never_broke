package payroll

import (
	"errors"
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
	regions, err := s.repo.ListRegions()
	if err != nil {
		return nil, err
	}

	items := make([]RegionListItem, len(regions))
	for i, region := range regions {
		items[i] = RegionListItem{
			Code: region.Code,
			Name: region.Name,
		}
	}
	return items, nil
}

// GetRegionRule 获取地区规则
func (s *Service) GetRegionRule(regionCode string, year int) (*RegionRuleResponse, error) {
	rule, err := s.repo.GetRegionRule(regionCode, year)
	if err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			// 尝试获取默认地区规则
			rule, err = s.repo.GetRegionRule("CN-DEFAULT", year)
			if err != nil {
				return nil, ErrRuleNotFound
			}
		} else {
			return nil, err
		}
	}

	return &RegionRuleResponse{
		RegionCode:           rule.RegionCode,
		Year:                 rule.Year,
		SSBaseMin:            rule.SSBaseMin,
		SSBaseMax:            rule.SSBaseMax,
		HFBaseMin:            rule.HFBaseMin,
		HFBaseMax:            rule.HFBaseMax,
		SSPersonRateSum:      rule.SSPersonRateSum,
		SSCompanyRateSum:     rule.SSCompanyRateSum,
		HFRateMin:            rule.HFRateMin,
		HFRateMax:            rule.HFRateMax,
		DefaultHFRatePerson:  rule.DefaultHFRatePerson,
		DefaultHFRateCompany: rule.DefaultHFRateCompany,
	}, nil
}

// CalculateSalary 计算薪资
func (s *Service) CalculateSalary(req *CalcSalaryRequest) (*CalcSalaryResponse, error) {
	// 获取地区规则
	rule, err := s.repo.GetRegionRule(req.RegionCode, req.Year)
	if err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			// 尝试获取默认地区规则
			rule, err = s.repo.GetRegionRule("CN-DEFAULT", req.Year)
			if err != nil {
				return nil, ErrRuleNotFound
			}
		} else {
			return nil, err
		}
	}

	// 获取税率表
	brackets, err := s.repo.GetTaxBrackets(req.Year)
	if err != nil {
		return nil, err
	}

	// 创建计算器并执行计算
	calculator := NewCalculator(rule, brackets)
	return calculator.Calculate(req)
}
