package crawler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"social-data-tool/model"
)

// ParseM12333HTML 从 m12333.cn 页面解析社保数据
func ParseM12333HTML(html string) []model.CitySSData {
	var results []model.CitySSData

	// 清理 HTML
	text := cleanHTML(html)

	// 提取基数上限
	upperLimit := extractNumber(text, `缴费基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`)
	// 提取基数下限
	lowerLimit := extractNumber(text, `缴费基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`)

	// 提取城市名称
	cityName := extractCityName(text)

	if upperLimit > 0 || lowerLimit > 0 {
		results = append(results, model.CitySSData{
			CityName:  cityName,
			SSBaseMin: lowerLimit,
			SSBaseMax: upperLimit,
		})
	}

	return results
}

// extractNumber 使用正则提取数字
func extractNumber(text, pattern string) float64 {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(text)
	if len(match) >= 2 {
		val, err := strconv.ParseFloat(match[1], 64)
		if err == nil {
			return val
		}
	}
	return 0
}

// extractCityName 提取城市名称
func extractCityName(text string) string {
	// 常见城市名模式
	cityPattern := regexp.MustCompile(`(北京|上海|广州|深圳|杭州|南京|成都|武汉|天津|重庆|西安|苏州|东莞|青岛|郑州|长沙|沈阳|宁波|佛山|无锡)[市]?`)
	match := cityPattern.FindString(text)
	if match != "" {
		return strings.TrimSuffix(match, "市")
	}
	return ""
}

// ExtractAllDataFromHTML 从 HTML 提取所有可能的社保数据
func ExtractAllDataFromHTML(html string) map[string]interface{} {
	text := cleanHTML(html)

	data := map[string]interface{}{
		"rawTextLength": len(text),
	}

	// 提取所有数字和标签
	patterns := map[string]string{
		"养老保险基数上限": `养老保险[^0-9]*基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"养老保险基数下限": `养老保险[^0-9]*基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"医疗保险基数上限": `医疗保险[^0-9]*基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"医疗保险基数下限": `医疗保险[^0-9]*基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"失业保险基数上限": `失业保险[^0-9]*基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"失业保险基数下限": `失业保险[^0-9]*基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"公积金基数上限":  `公积金[^0-9]*基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"公积金基数下限":  `公积金[^0-9]*基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"缴费基数上限":   `缴费基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"缴费基数下限":   `缴费基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"社保基数上限":   `社保基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		"社保基数下限":   `社保基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
	}

	for name, pattern := range patterns {
		val := extractNumber(text, pattern)
		if val > 0 {
			data[name] = val
		}
	}

	// 提取城市
	cityName := extractCityName(text)
	if cityName != "" {
		data["城市"] = cityName
	}

	// 提取百分比
	percentPatterns := map[string]string{
		"养老保险比例": `养老保险[^0-9]*(\d+(?:\.\d+)?)[%％]`,
		"医疗保险比例": `医疗保险[^0-9]*(\d+(?:\.\d+)?)[%％]`,
		"失业保险比例": `失业保险[^0-9]*(\d+(?:\.\d+)?)[%％]`,
		"公积金比例":  `公积金[^0-9]*(\d+(?:\.\d+)?)[%％]`,
	}

	for name, pattern := range percentPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		if len(matches) > 0 {
			var percents []float64
			for _, m := range matches {
				if len(m) >= 2 {
					if val, err := strconv.ParseFloat(m[1], 64); err == nil {
						percents = append(percents, val)
					}
				}
			}
			if len(percents) > 0 {
				data[name] = percents
			}
		}
	}

	return data
}

// PrintExtractedData 打印提取的数据
func PrintExtractedData(data map[string]interface{}) {
	fmt.Println("\n─────────────────────────────────────────────────────────────────")
	fmt.Println("                    提取的结构化数据")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	// 按分类打印
	categories := []struct {
		name string
		keys []string
	}{
		{"城市信息", []string{"城市"}},
		{"缴费基数", []string{"缴费基数上限", "缴费基数下限", "社保基数上限", "社保基数下限"}},
		{"养老保险", []string{"养老保险基数上限", "养老保险基数下限", "养老保险比例"}},
		{"医疗保险", []string{"医疗保险基数上限", "医疗保险基数下限", "医疗保险比例"}},
		{"失业保险", []string{"失业保险基数上限", "失业保险基数下限", "失业保险比例"}},
		{"公积金", []string{"公积金基数上限", "公积金基数下限", "公积金比例"}},
	}

	for _, cat := range categories {
		hasData := false
		for _, key := range cat.keys {
			if _, ok := data[key]; ok {
				hasData = true
				break
			}
		}
		if hasData {
			fmt.Printf("\n  📊 %s:\n", cat.name)
			for _, key := range cat.keys {
				if val, ok := data[key]; ok {
					fmt.Printf("     • %s: %v\n", key, val)
				}
			}
		}
	}
}
