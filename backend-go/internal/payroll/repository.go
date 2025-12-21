package payroll

import (
	"gorm.io/gorm"
)

// Repository 薪资计算仓储接口
type Repository interface {
	// ListRegions 获取所有地区
	ListRegions() ([]PayrollRegion, error)
	// GetRegionByCode 根据代码获取地区
	GetRegionByCode(code string) (*PayrollRegion, error)
	// GetRegionRule 获取地区规则
	GetRegionRule(regionCode string, year int) (*PayrollRegionRule, error)
	// GetTaxBrackets 获取税率表
	GetTaxBrackets(year int) ([]TaxBracketCN, error)
}

// GormRepository GORM实现
type GormRepository struct {
	db *gorm.DB
}

// NewRepository 创建仓储
func NewRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// ListRegions 获取所有地区
func (r *GormRepository) ListRegions() ([]PayrollRegion, error) {
	var regions []PayrollRegion
	if err := r.db.Order("code").Find(&regions).Error; err != nil {
		return nil, err
	}
	return regions, nil
}

// GetRegionByCode 根据代码获取地区
func (r *GormRepository) GetRegionByCode(code string) (*PayrollRegion, error) {
	var region PayrollRegion
	if err := r.db.Where("code = ?", code).First(&region).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRegionNotFound
		}
		return nil, err
	}
	return &region, nil
}

// GetRegionRule 获取地区规则
func (r *GormRepository) GetRegionRule(regionCode string, year int) (*PayrollRegionRule, error) {
	var rule PayrollRegionRule
	if err := r.db.Where("region_code = ? AND year = ?", regionCode, year).First(&rule).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrRuleNotFound
		}
		return nil, err
	}
	return &rule, nil
}

// GetTaxBrackets 获取税率表
func (r *GormRepository) GetTaxBrackets(year int) ([]TaxBracketCN, error) {
	var brackets []TaxBracketCN
	if err := r.db.Where("year = ?", year).Order("lower_bound").Find(&brackets).Error; err != nil {
		return nil, err
	}
	if len(brackets) == 0 {
		return nil, ErrTaxBracketsNotFound
	}
	return brackets, nil
}
