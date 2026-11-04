package payroll

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// Official notices are reviewed snapshots, served without a dependency on legacy
// seed tables or live scraping. Never extend a snapshot beyond its review window.
type RuleSource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type officialRule struct {
	Code, Name                 string
	Year                       int
	From, Through              string
	SSMin, SSMax, HFMin, HFMax int64
	HFRateMax                  float64
	Sources                    []RuleSource
}

var officialRules = []officialRule{
	{Code: "CN-BJ", Name: "北京", Year: 2025, From: "2025-07-01", Through: "2025-08-31", SSMin: 7162, SSMax: 35811, HFMin: 2420, HFMax: 35811, HFRateMax: .12, Sources: []RuleSource{
		{"北京市人社局 · 2025年度社保基数", "https://www.beijing.gov.cn/fuwu/bmfw/sy/jrts/tzxx/202509/t20250919_4205649.html"},
		{"北京住房公积金管理中心 · 2025年度缴存政策（下限9月调整）", "https://banshi.beijing.gov.cn/zcjd/202509/t20250925_428450.html"},
		{"北京住房公积金管理中心 · 调整前下限", "https://gjj.beijing.gov.cn/web/zwgk61/2024zcwj/436433461/543368452/index.html"},
	}},
	{Code: "CN-BJ", Name: "北京", Year: 2024, From: "2024-07-01", Through: "2025-06-30", SSMin: 6821, SSMax: 35283, HFMin: 2420, HFMax: 35283, HFRateMax: .12, Sources: []RuleSource{
		{"北京市人社局 · 2024年度社保基数", "https://rsj.beijing.gov.cn/xxgk/2024zcjd/202407/t20240731_3763679.html"},
		{"北京住房公积金管理中心 · 2024年度缴存政策", "https://gjj.beijing.gov.cn/web/zwgk61/2024zcwj/436433461/543368452/index.html"},
	}},
	{Code: "CN-SH", Name: "上海", Year: 2025, From: "2025-07-01", Through: "2026-06-30", SSMin: 7460, SSMax: 37302, HFMin: 2690, HFMax: 37302, HFRateMax: .07, Sources: []RuleSource{
		{"上海市人社局 · 2025年度社保基数", "https://rsj.sh.gov.cn/tgsgg_17341/20250918/t0035_1435637.html"},
		{"上海市公积金中心 · 2025年度缴存政策", "https://www.shzfgjj.cn/html/newxxgk/zcwj/gfxwj/228051.html"},
	}},
	{Code: "CN-BJ", Name: "北京", Year: 2025, From: "2025-09-01", Through: "2026-06-30", SSMin: 7162, SSMax: 35811, HFMin: 2540, HFMax: 35811, HFRateMax: .12, Sources: []RuleSource{
		{"北京市人社局 · 2025年度社保基数", "https://www.beijing.gov.cn/fuwu/bmfw/sy/jrts/tzxx/202509/t20250919_4205649.html"},
		{"北京住房公积金管理中心 · 2025年度缴存政策", "https://gjj.beijing.gov.cn/web/zwgk61/2024zcjd/743765478/index.html"},
	}},
	{Code: "CN-BJ", Name: "北京", Year: 2026, From: "2026-07-01", Through: "2027-06-30", SSMin: 7270, SSMax: 36348, HFMin: 2540, HFMax: 36348, HFRateMax: .12, Sources: []RuleSource{
		{"北京市人社局 · 2026年度社保基数", "https://www.beijing.gov.cn/zhengce/zhengcefagui/202608/t20260821_4831679.html"},
		{"北京住房公积金管理中心 · 2026年度缴存政策", "https://gjj.beijing.gov.cn/web/zwgk61/2024zcjd/744103386/index.html"},
	}},
	{Code: "CN-SH", Name: "上海", Year: 2026, From: "2026-07-01", Through: "2027-06-30", SSMin: 7546, SSMax: 37731, HFMin: 2740, HFMax: 37731, HFRateMax: .07, Sources: []RuleSource{
		{"上海市人社局 · 2026年度社保基数", "https://rsj.sh.gov.cn/tdjjf_17554/20260824/t0035_1443297.html"},
		{"上海市住房公积金管理委员会 · 2026年度缴存政策", "https://www.shanghai.gov.cn/gwk/affairs/content/0064%40f3d788ec72e9404ebcbf3d27fda98435"},
	}},
}

func findOfficialRule(code, date string) (*RegionRuleResponse, error) {
	parsed, err := time.Parse(time.DateOnly, date)
	if err != nil || parsed.Format(time.DateOnly) != date {
		return nil, fmt.Errorf("invalid rule date")
	}
	for _, snapshot := range officialRules {
		if snapshot.Code != code || date < snapshot.From || date > snapshot.Through {
			continue
		}
		return &RegionRuleResponse{
			RegionCode: code, Year: snapshot.Year, EffectiveFrom: snapshot.From, EffectiveThrough: snapshot.Through, AsOf: date, VerifiedAt: "2026-11-04", Sources: snapshot.Sources,
			SSBaseMin: decimal.NewFromInt(snapshot.SSMin), SSBaseMax: decimal.NewFromInt(snapshot.SSMax), HFBaseMin: decimal.NewFromInt(snapshot.HFMin), HFBaseMax: decimal.NewFromInt(snapshot.HFMax),
			// These remain editable calculation assumptions; the linked notices certify bases.
			SSPersonRateSum: decimal.NewFromFloat(.105), SSCompanyRateSum: decimal.NewFromFloat(.28),
			HFRateMin: decimal.NewFromFloat(.05), HFRateMax: decimal.NewFromFloat(snapshot.HFRateMax), DefaultHFRatePerson: decimal.NewFromFloat(snapshot.HFRateMax), DefaultHFRateCompany: decimal.NewFromFloat(snapshot.HFRateMax),
		}, nil
	}
	return nil, ErrRuleNotFound
}

func currentRuleDate() string {
	return time.Now().In(time.FixedZone("Asia/Shanghai", 8*60*60)).Format(time.DateOnly)
}
