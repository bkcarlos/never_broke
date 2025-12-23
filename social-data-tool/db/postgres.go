package db

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"social-data-tool/model"

	"github.com/jackc/pgx/v5"
)

// SyncToDatabase 同步数据到数据库
func SyncToDatabase(ctx context.Context, dsn string, data []model.CitySSData) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	defer conn.Close(ctx)

	log.Println("📊 开始同步数据到数据库...")

	year := time.Now().Year()
	successCount := 0

	for _, city := range data {
		// 生成地区代码
		regionCode := generateRegionCode(city.CityCode, city.CityName)

		// Upsert region
		_, err = conn.Exec(ctx, `
			INSERT INTO payroll_regions (code, name, updated_at)
			VALUES ($1, $2, NOW())
			ON CONFLICT (code) DO UPDATE SET name = $2, updated_at = NOW()
		`, regionCode, city.CityName)
		if err != nil {
			log.Printf("⚠️ 插入地区 %s 失败: %v", city.CityName, err)
			continue
		}

		// 设置默认值
		ssPersonRate := city.SSPersonRateSum
		if ssPersonRate == 0 {
			ssPersonRate = 0.105
		}
		ssCompanyRate := city.SSCompanyRateSum
		if ssCompanyRate == 0 {
			ssCompanyRate = 0.28
		}
		hfRateMin := city.HFRateMin
		if hfRateMin == 0 {
			hfRateMin = 0.05
		}
		hfRateMax := city.HFRateMax
		if hfRateMax == 0 {
			hfRateMax = 0.12
		}
		defaultHFPerson := city.DefaultHFRatePerson
		if defaultHFPerson == 0 {
			defaultHFPerson = 0.12
		}
		defaultHFCompany := city.DefaultHFRateCompany
		if defaultHFCompany == 0 {
			defaultHFCompany = 0.12
		}

		// 公积金基数默认使用社保基数
		hfBaseMin := city.HFBaseMin
		if hfBaseMin == 0 {
			hfBaseMin = city.SSBaseMin
		}
		hfBaseMax := city.HFBaseMax
		if hfBaseMax == 0 {
			hfBaseMax = city.SSBaseMax
		}

		// Upsert region rule
		_, err = conn.Exec(ctx, `
			INSERT INTO payroll_region_rules (
				region_code, year,
				ss_base_min, ss_base_max,
				hf_base_min, hf_base_max,
				ss_person_rate_sum, ss_company_rate_sum,
				hf_rate_min, hf_rate_max,
				default_hf_rate_person, default_hf_rate_company,
				updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
			ON CONFLICT (region_code, year) DO UPDATE SET
				ss_base_min = $3, ss_base_max = $4,
				hf_base_min = $5, hf_base_max = $6,
				ss_person_rate_sum = $7, ss_company_rate_sum = $8,
				hf_rate_min = $9, hf_rate_max = $10,
				default_hf_rate_person = $11, default_hf_rate_company = $12,
				updated_at = NOW()
		`, regionCode, year,
			city.SSBaseMin, city.SSBaseMax,
			hfBaseMin, hfBaseMax,
			ssPersonRate, ssCompanyRate,
			hfRateMin, hfRateMax,
			defaultHFPerson, defaultHFCompany,
		)
		if err != nil {
			log.Printf("⚠️ 插入规则 %s 失败: %v", city.CityName, err)
			continue
		}

		log.Printf("  ✓ %s (%s)", city.CityName, regionCode)
		successCount++
	}

	log.Printf("✅ 同步完成，成功 %d/%d 个城市", successCount, len(data))
	return nil
}

// generateRegionCode 生成地区代码
func generateRegionCode(cityCode, cityName string) string {
	// 特殊城市映射
	specialCodes := map[string]string{
		"beijing":   "CN-BJ",
		"shanghai":  "CN-SH",
		"tianjin":   "CN-TJ",
		"chongqing": "CN-CQ",
		// 广东
		"guangzhou": "CN-GD-GZ",
		"shenzhen":  "CN-GD-SZ",
		"dongguan":  "CN-GD-DG",
		"foshan":    "CN-GD-FS",
		"zhuhai":    "CN-GD-ZH",
		"huizhou":   "CN-GD-HZ",
		"zhongshan": "CN-GD-ZS",
		// 江苏
		"nanjing":   "CN-JS-NJ",
		"suzhou":    "CN-JS-SZ",
		"wuxi":      "CN-JS-WX",
		"changzhou": "CN-JS-CZ",
		// 浙江
		"hangzhou": "CN-ZJ-HZ",
		"ningbo":   "CN-ZJ-NB",
		"wenzhou":  "CN-ZJ-WZ",
		// 四川
		"chengdu": "CN-SC-CD",
		// 湖北
		"wuhan": "CN-HB-WH",
		// 陕西
		"xian": "CN-SX-XA",
		// 湖南
		"changsha": "CN-HN-CS",
		// 河南
		"zhengzhou": "CN-HA-ZZ",
		// 山东
		"jinan":   "CN-SD-JN",
		"qingdao": "CN-SD-QD",
		// 辽宁
		"shenyang": "CN-LN-SY",
		"dalian":   "CN-LN-DL",
		// 福建
		"fuzhou":   "CN-FJ-FZ",
		"xiamen":   "CN-FJ-XM",
		"quanzhou": "CN-FJ-QZ",
		// 黑龙江
		"harbin": "CN-HL-HRB",
		// 吉林
		"changchun": "CN-JL-CC",
		// 江西
		"nanchang": "CN-JX-NC",
		// 安徽
		"hefei": "CN-AH-HF",
		// 河北
		"shijiazhuang": "CN-HE-SJZ",
		// 山西
		"taiyuan": "CN-SX-TY",
		// 云南
		"kunming": "CN-YN-KM",
		// 贵州
		"guiyang": "CN-GZ-GY",
		// 广西
		"nanning": "CN-GX-NN",
		// 甘肃
		"lanzhou": "CN-GS-LZ",
	}

	if code, ok := specialCodes[strings.ToLower(cityCode)]; ok {
		return code
	}

	// 默认使用城市代码
	return fmt.Sprintf("CN-%s", strings.ToUpper(cityCode))
}
