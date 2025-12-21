package payroll

import (
	"time"

	"github.com/shopspring/decimal"
)

// PayrollRegion 地区
type PayrollRegion struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Code      string    `gorm:"type:varchar(50);unique;not null" json:"code"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	CreatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:timestamptz;not null;default:now()" json:"updatedAt"`
}

func (PayrollRegion) TableName() string {
	return "payroll_regions"
}

// PayrollRegionRule 地区规则（按年度）
type PayrollRegionRule struct {
	ID                   string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RegionCode           string          `gorm:"type:varchar(50);not null;index:idx_region_code" json:"regionCode"`
	Year                 int             `gorm:"type:int;not null;index:idx_year" json:"year"`
	SSBaseMin            decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0" json:"ssBaseMin"`
	SSBaseMax            decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0" json:"ssBaseMax"`
	HFBaseMin            decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0" json:"hfBaseMin"`
	HFBaseMax            decimal.Decimal `gorm:"type:numeric(15,2);not null;default:0" json:"hfBaseMax"`
	SSPersonRateSum      decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0" json:"ssPersonRateSum"`
	SSCompanyRateSum     decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0" json:"ssCompanyRateSum"`
	HFRateMin            decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0.05" json:"hfRateMin"`
	HFRateMax            decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0.12" json:"hfRateMax"`
	DefaultHFRatePerson  decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0.12" json:"defaultHfRatePerson"`
	DefaultHFRateCompany decimal.Decimal `gorm:"type:numeric(10,6);not null;default:0.12" json:"defaultHfRateCompany"`
	CreatedAt            time.Time       `gorm:"type:timestamptz;not null;default:now()" json:"createdAt"`
	UpdatedAt            time.Time       `gorm:"type:timestamptz;not null;default:now()" json:"updatedAt"`
}

func (PayrollRegionRule) TableName() string {
	return "payroll_region_rules"
}

// TaxBracketCN 中国综合所得税率表
type TaxBracketCN struct {
	ID             string           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Year           int              `gorm:"type:int;not null;index:idx_year" json:"year"`
	LowerBound     decimal.Decimal  `gorm:"type:numeric(15,2);not null" json:"lowerBound"`
	UpperBound     *decimal.Decimal `gorm:"type:numeric(15,2)" json:"upperBound"` // NULL 表示无上限
	Rate           decimal.Decimal  `gorm:"type:numeric(10,6);not null" json:"rate"`
	QuickDeduction decimal.Decimal  `gorm:"type:numeric(15,2);not null;default:0" json:"quickDeduction"`
	CreatedAt      time.Time        `gorm:"type:timestamptz;not null;default:now()" json:"createdAt"`
	UpdatedAt      time.Time        `gorm:"type:timestamptz;not null;default:now()" json:"updatedAt"`
}

func (TaxBracketCN) TableName() string {
	return "tax_brackets_cn"
}
