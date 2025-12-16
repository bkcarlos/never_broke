package citysocialrate

// CitySocialRate 城市社保比例配置
type CitySocialRate struct {
	ID               string  `json:"id"`
	CityCode         string  `json:"cityCode"`
	CityName         string  `json:"cityName"`
	EffectiveDate    string  `json:"effectiveDate"` // YYYY-MM-DD
	PensionRate      float64 `json:"pensionRate"`
	MedicalRate      float64 `json:"medicalRate"`
	UnemploymentRate float64 `json:"unemploymentRate"`
	InjuryRate       float64 `json:"injuryRate"`
	MaternityRate    float64 `json:"maternityRate"`
	HousingFundRate  float64 `json:"housingFundRate"`
	TotalRate        float64 `json:"totalRate"`
	IsActive         bool    `json:"isActive"`
	Remark           string  `json:"remark,omitempty"`
	CreatedAt        string  `json:"createdAt"`
}

// CitySocialRateDetail 社保比例明细(用于前端展示)
type CitySocialRateDetail struct {
	CityCode      string                `json:"cityCode"`
	CityName      string                `json:"cityName"`
	TotalRate     float64               `json:"totalRate"`
	Components    []SocialRateComponent `json:"components"`
	EffectiveDate string                `json:"effectiveDate"`
	Remark        string                `json:"remark,omitempty"`
}

// SocialRateComponent 社保组成部分
type SocialRateComponent struct {
	Name string  `json:"name"` // 养老/医疗/失业/公积金
	Rate float64 `json:"rate"`
}
