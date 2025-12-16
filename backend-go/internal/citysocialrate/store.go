package citysocialrate

import "sync"

var (
	mu    sync.RWMutex
	rates = []CitySocialRate{}
)

// InitDefaultRates 初始化默认社保比例数据(用于内存存储)
func InitDefaultRates() {
	mu.Lock()
	defer mu.Unlock()

	rates = []CitySocialRate{
		{
			ID:               "beijing-2024",
			CityCode:         "beijing",
			CityName:         "北京",
			EffectiveDate:    "2024-01-01",
			PensionRate:      0.08,
			MedicalRate:      0.02,
			UnemploymentRate: 0.002,
			InjuryRate:       0,
			MaternityRate:    0,
			HousingFundRate:  0.12,
			TotalRate:        0.202,
			IsActive:         true,
			Remark:           "养老8% + 医疗2% + 失业0.2% + 公积金12%",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "shanghai-2024",
			CityCode:         "shanghai",
			CityName:         "上海",
			EffectiveDate:    "2024-01-01",
			PensionRate:      0.08,
			MedicalRate:      0.02,
			UnemploymentRate: 0.005,
			InjuryRate:       0,
			MaternityRate:    0,
			HousingFundRate:  0.07,
			TotalRate:        0.175,
			IsActive:         true,
			Remark:           "养老8% + 医疗2% + 失业0.5% + 公积金7%",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "guangzhou-2024",
			CityCode:         "guangzhou",
			CityName:         "广州",
			EffectiveDate:    "2024-01-01",
			PensionRate:      0.08,
			MedicalRate:      0.02,
			UnemploymentRate: 0.002,
			InjuryRate:       0,
			MaternityRate:    0,
			HousingFundRate:  0.05,
			TotalRate:        0.152,
			IsActive:         true,
			Remark:           "养老8% + 医疗2% + 失业0.2% + 公积金5%",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "shenzhen-2024",
			CityCode:         "shenzhen",
			CityName:         "深圳",
			EffectiveDate:    "2024-01-01",
			PensionRate:      0.08,
			MedicalRate:      0.02,
			UnemploymentRate: 0.005,
			InjuryRate:       0,
			MaternityRate:    0,
			HousingFundRate:  0.05,
			TotalRate:        0.155,
			IsActive:         true,
			Remark:           "养老8% + 医疗2% + 失业0.5% + 公积金5%",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
		{
			ID:               "default-2024",
			CityCode:         "default",
			CityName:         "其他城市",
			EffectiveDate:    "2024-01-01",
			PensionRate:      0.08,
			MedicalRate:      0.02,
			UnemploymentRate: 0.002,
			InjuryRate:       0,
			MaternityRate:    0,
			HousingFundRate:  0.10,
			TotalRate:        0.182,
			IsActive:         true,
			Remark:           "养老8% + 医疗2% + 失业0.2% + 公积金10%(参考值)",
			CreatedAt:        "2024-01-01T00:00:00Z",
		},
	}
}

// ListRates 列出所有社保比例配置
func ListRates() []CitySocialRate {
	mu.RLock()
	defer mu.RUnlock()
	result := make([]CitySocialRate, len(rates))
	copy(result, rates)
	return result
}

// GetActiveByCityCode 获取指定城市当前生效的社保比例
func GetActiveByCityCode(cityCode string) *CitySocialRate {
	mu.RLock()
	defer mu.RUnlock()

	for i := range rates {
		if rates[i].CityCode == cityCode && rates[i].IsActive {
			result := rates[i]
			return &result
		}
	}

	// 如果没找到,返回默认城市
	for i := range rates {
		if rates[i].CityCode == "default" && rates[i].IsActive {
			result := rates[i]
			return &result
		}
	}

	return nil
}

// GetRateDetail 获取社保比例详情(含各项拆分)
func GetRateDetail(cityCode string) *CitySocialRateDetail {
	rate := GetActiveByCityCode(cityCode)
	if rate == nil {
		return nil
	}

	components := []SocialRateComponent{}

	if rate.PensionRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "养老保险",
			Rate: rate.PensionRate,
		})
	}

	if rate.MedicalRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "医疗保险",
			Rate: rate.MedicalRate,
		})
	}

	if rate.UnemploymentRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "失业保险",
			Rate: rate.UnemploymentRate,
		})
	}

	if rate.InjuryRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "工伤保险",
			Rate: rate.InjuryRate,
		})
	}

	if rate.MaternityRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "生育保险",
			Rate: rate.MaternityRate,
		})
	}

	if rate.HousingFundRate > 0 {
		components = append(components, SocialRateComponent{
			Name: "住房公积金",
			Rate: rate.HousingFundRate,
		})
	}

	return &CitySocialRateDetail{
		CityCode:      rate.CityCode,
		CityName:      rate.CityName,
		TotalRate:     rate.TotalRate,
		Components:    components,
		EffectiveDate: rate.EffectiveDate,
		Remark:        rate.Remark,
	}
}
