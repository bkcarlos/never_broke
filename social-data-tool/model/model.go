package model

import "time"

// CitySSData 城市社保数据
type CitySSData struct {
	CityCode  string  `json:"cityCode"`
	CityName  string  `json:"cityName"`
	Province  string  `json:"province,omitempty"`
	Year      int     `json:"year"`
	SSBaseMin float64 `json:"ssBaseMin"` // 社保基数下限
	SSBaseMax float64 `json:"ssBaseMax"` // 社保基数上限
	HFBaseMin float64 `json:"hfBaseMin"` // 公积金基数下限
	HFBaseMax float64 `json:"hfBaseMax"` // 公积金基数上限
	// 社保缴费比例（个人）
	SSPersonRateSum float64 `json:"ssPersonRateSum,omitempty"`
	// 社保缴费比例（公司）
	SSCompanyRateSum float64 `json:"ssCompanyRateSum,omitempty"`
	// 公积金比例范围
	HFRateMin float64 `json:"hfRateMin,omitempty"`
	HFRateMax float64 `json:"hfRateMax,omitempty"`
	// 默认公积金比例
	DefaultHFRatePerson  float64 `json:"defaultHfRatePerson,omitempty"`
	DefaultHFRateCompany float64 `json:"defaultHfRateCompany,omitempty"`
	// 数据来源
	SourceURL string `json:"sourceUrl,omitempty"`
	CrawledAt string `json:"crawledAt,omitempty"`
}

// DataFile 数据文件结构
type DataFile struct {
	Version     string       `json:"version"`
	LastUpdated string       `json:"lastUpdated"`
	DataSource  string       `json:"dataSource"`
	Disclaimer  string       `json:"disclaimer"`
	Cities      []CitySSData `json:"cities"`
}

// NewDataFile 创建数据文件
func NewDataFile(cities []CitySSData) *DataFile {
	// 自动检测数据来源
	dataSource := "https://m12333.cn/"
	if len(cities) > 0 && cities[0].SourceURL != "" {
		dataSource = cities[0].SourceURL
	}

	return &DataFile{
		Version:     "2025.1",
		LastUpdated: time.Now().Format("2006-01-02"),
		DataSource:  dataSource,
		Disclaimer:  "数据仅供参考，请以当地人社局官方数据为准。如有侵权请联系删除。",
		Cities:      cities,
	}
}
