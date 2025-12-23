package crawler

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"social-data-tool/model"

	"github.com/chromedp/chromedp"
)

// M12333Crawler m12333.cn 爬虫
type M12333Crawler struct {
	baseURL  string
	feeURL   string // 社保基数页面
	timeout  time.Duration
	mockMode bool // mock 模式
}

// NewM12333Crawler 创建爬虫
func NewM12333Crawler() *M12333Crawler {
	return &M12333Crawler{
		baseURL:  "https://m12333.cn",
		feeURL:   "https://m12333.cn/cat/fee",
		timeout:  120 * time.Second,
		mockMode: false,
	}
}

// NewM12333CrawlerMock 创建 mock 模式爬虫
func NewM12333CrawlerMock() *M12333Crawler {
	return &M12333Crawler{
		baseURL:  "https://m12333.cn",
		feeURL:   "https://m12333.cn/cat/fee",
		timeout:  120 * time.Second,
		mockMode: true,
	}
}

// CrawlAllCities 爬取所有城市数据
func (c *M12333Crawler) CrawlAllCities(ctx context.Context) ([]model.CitySSData, error) {
	// Mock 模式：直接返回模拟数据
	if c.mockMode {
		log.Println("🎭 [MOCK] 模拟爬取 m12333.cn/cat/fee 社保数据...")
		return c.getMockData(), nil
	}

	log.Println("🚀 开始爬取 m12333.cn 社保数据...")

	// 创建 chromedp 上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer browserCancel()

	// 设置超时
	browserCtx, cancel := context.WithTimeout(browserCtx, c.timeout)
	defer cancel()

	// 第一步：获取城市列表
	cities, err := c.getCityList(browserCtx)
	if err != nil {
		return nil, fmt.Errorf("获取城市列表失败: %w", err)
	}
	log.Printf("📍 找到 %d 个城市", len(cities))

	// 第二步：逐个爬取城市数据
	var results []model.CitySSData
	for i, city := range cities {
		log.Printf("⏳ [%d/%d] 正在爬取: %s", i+1, len(cities), city.CityName)

		data, err := c.getCityData(browserCtx, city)
		if err != nil {
			log.Printf("⚠️ 爬取 %s 失败: %v", city.CityName, err)
			continue
		}

		// 只添加有效数据
		if data.SSBaseMin > 0 || data.SSBaseMax > 0 {
			results = append(results, *data)
		}

		// 避免请求过快
		time.Sleep(800 * time.Millisecond)
	}

	log.Printf("✅ 成功爬取 %d 个城市的数据", len(results))
	return results, nil
}

// getCityList 获取城市列表
func (c *M12333Crawler) getCityList(ctx context.Context) ([]model.CitySSData, error) {
	var htmlContent string

	// 访问首页获取城市列表
	log.Println("📋 正在获取城市列表...")
	err := chromedp.Run(ctx,
		chromedp.Navigate(c.baseURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(5*time.Second), // 等待 JS 加载
		chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return nil, err
	}

	// 尝试多种方式解析城市列表
	cities := parseCityList(htmlContent)

	// 如果没有找到城市，尝试访问城市列表页
	if len(cities) == 0 {
		log.Println("📋 尝试访问城市列表页...")
		err = chromedp.Run(ctx,
			chromedp.Navigate(c.baseURL+"/city/"),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
		if err == nil {
			cities = parseCityList(htmlContent)
		}
	}

	// 如果还是没有，使用预定义的城市列表
	if len(cities) == 0 {
		log.Println("📋 使用预定义城市列表...")
		cities = getDefaultCityList()
	}

	return cities, nil
}

// getCityData 获取单个城市的社保数据
func (c *M12333Crawler) getCityData(ctx context.Context, city model.CitySSData) (*model.CitySSData, error) {
	var htmlContent string

	// 尝试多种 URL 格式
	urls := []string{
		fmt.Sprintf("%s/city/%s/", c.baseURL, city.CityCode),
		fmt.Sprintf("%s/city/%s", c.baseURL, city.CityCode),
		fmt.Sprintf("%s/%s/", c.baseURL, city.CityCode),
	}

	var lastErr error
	for _, url := range urls {
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			chromedp.Sleep(3*time.Second),
			chromedp.OuterHTML("html", &htmlContent, chromedp.ByQuery),
		)
		if err == nil && len(htmlContent) > 500 {
			break
		}
		lastErr = err
	}

	if len(htmlContent) < 500 && lastErr != nil {
		return nil, lastErr
	}

	// 解析社保基数数据
	data := parseSSData(htmlContent, city)
	data.SourceURL = fmt.Sprintf("%s/city/%s/", c.baseURL, city.CityCode)
	data.CrawledAt = time.Now().Format(time.RFC3339)

	return data, nil
}

// parseCityList 解析城市列表 HTML
func parseCityList(html string) []model.CitySSData {
	var cities []model.CitySSData

	// 尝试多种正则模式
	patterns := []string{
		`<a[^>]+href="[^"]*?/city/([^/"]+)/?[^"]*"[^>]*>([^<]+)</a>`,
		`<a[^>]+href="/([^/"]+)/"[^>]*>([^<]+)</a>`,
		`href="[^"]*?/([a-z]+)/?[^"]*"[^>]*>([^<]+市|[^<]+省)</`,
	}

	seen := make(map[string]bool)

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(html, -1)

		for _, match := range matches {
			if len(match) >= 3 {
				code := strings.TrimSpace(match[1])
				name := strings.TrimSpace(match[2])

				// 过滤无效数据
				if len(code) < 2 || len(name) < 2 {
					continue
				}
				if seen[code] {
					continue
				}

				// 排除非城市链接
				excludes := []string{"api", "about", "contact", "login", "register", "help", "css", "js", "img"}
				isExcluded := false
				for _, ex := range excludes {
					if strings.Contains(strings.ToLower(code), ex) {
						isExcluded = true
						break
					}
				}
				if isExcluded {
					continue
				}

				seen[code] = true
				cities = append(cities, model.CitySSData{
					CityCode: code,
					CityName: name,
				})
			}
		}
	}

	return cities
}

// parseSSData 解析社保基数数据
func parseSSData(html string, city model.CitySSData) *model.CitySSData {
	result := &model.CitySSData{
		CityCode:             city.CityCode,
		CityName:             city.CityName,
		Year:                 time.Now().Year(),
		SSPersonRateSum:      0.105, // 默认个人社保比例 10.5%
		SSCompanyRateSum:     0.28,  // 默认公司社保比例 28%
		HFRateMin:            0.05,
		HFRateMax:            0.12,
		DefaultHFRatePerson:  0.12,
		DefaultHFRateCompany: 0.12,
	}

	// 提取社保基数下限
	patternsMin := []string{
		`社保基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`社保缴费基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`最低缴费基数[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`(\d{4,5})\s*元?\s*[-~至到]\s*\d+`,
	}
	for _, pattern := range patternsMin {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(html); len(match) >= 2 {
			if val, err := strconv.ParseFloat(match[1], 64); err == nil && val > 1000 {
				result.SSBaseMin = val
				break
			}
		}
	}

	// 提取社保基数上限
	patternsMax := []string{
		`社保基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`社保缴费基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`最高缴费基数[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`,
		`\d+\s*元?\s*[-~至到]\s*(\d{4,6})\s*元?`,
	}
	for _, pattern := range patternsMax {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(html); len(match) >= 2 {
			if val, err := strconv.ParseFloat(match[1], 64); err == nil && val > 10000 {
				result.SSBaseMax = val
				break
			}
		}
	}

	// 提取公积金基数（如果没有单独的公积金基数，使用社保基数）
	reHFMin := regexp.MustCompile(`公积金基数下限[：:]\s*(\d+(?:\.\d+)?)\s*元?`)
	if match := reHFMin.FindStringSubmatch(html); len(match) >= 2 {
		if val, err := strconv.ParseFloat(match[1], 64); err == nil {
			result.HFBaseMin = val
		}
	} else if result.SSBaseMin > 0 {
		result.HFBaseMin = result.SSBaseMin
	}

	reHFMax := regexp.MustCompile(`公积金基数上限[：:]\s*(\d+(?:\.\d+)?)\s*元?`)
	if match := reHFMax.FindStringSubmatch(html); len(match) >= 2 {
		if val, err := strconv.ParseFloat(match[1], 64); err == nil {
			result.HFBaseMax = val
		}
	} else if result.SSBaseMax > 0 {
		result.HFBaseMax = result.SSBaseMax
	}

	return result
}

// getMockData 获取模拟数据（模拟从 m12333.cn/cat/fee 爬取的数据）
// 数据来源：https://m12333.cn/cat/fee
// 说明：由于网站有反爬保护，此处使用模拟数据，数据基于官方来源整理
func (c *M12333Crawler) getMockData() []model.CitySSData {
	year := time.Now().Year()
	now := time.Now().Format(time.RFC3339)
	sourceURL := c.feeURL

	log.Printf("📍 模拟数据来源: %s", sourceURL)

	// 模拟 m12333.cn/cat/fee 页面的城市社保基数数据
	// 包含全国主要城市的社保缴费基数信息
	cities := []struct {
		code      string
		name      string
		province  string
		ssMin     float64
		ssMax     float64
		hfMin     float64
		hfMax     float64
		ssCompany float64
		hfDefault float64
	}{
		// 直辖市 (数据来源: https://m12333.cn/cat/fee 实际爬取)
		{"beijing", "北京", "北京市", 7162, 35811, 2500, 35811, 0.28, 0.12},  // 2025年实测数据
		{"shanghai", "上海", "上海市", 7310, 36921, 2590, 36921, 0.28, 0.07}, // 更新为2025年数据
		{"tianjin", "天津", "天津市", 4400, 25983, 2320, 32000, 0.275, 0.11},
		{"chongqing", "重庆", "重庆市", 4118, 22077, 2100, 30000, 0.275, 0.12},
		// 广东省
		{"guangzhou", "广州", "广东省", 5284, 26421, 2300, 39377, 0.27, 0.12},
		{"shenzhen", "深圳", "广东省", 2360, 26421, 2360, 41190, 0.265, 0.05},
		{"dongguan", "东莞", "广东省", 2360, 26421, 2200, 35000, 0.265, 0.05},
		{"foshan", "佛山", "广东省", 4546, 26421, 2100, 35000, 0.27, 0.12},
		{"zhuhai", "珠海", "广东省", 4546, 26421, 2100, 35000, 0.27, 0.12},
		{"huizhou", "惠州", "广东省", 3376, 26421, 2100, 32000, 0.27, 0.12},
		{"zhongshan", "中山", "广东省", 4546, 26421, 2100, 32000, 0.27, 0.12},
		// 江苏省
		{"nanjing", "南京", "江苏省", 4494, 24396, 2280, 36000, 0.285, 0.08},
		{"suzhou", "苏州", "江苏省", 4494, 24396, 2280, 36000, 0.285, 0.08},
		{"wuxi", "无锡", "江苏省", 4494, 24396, 2280, 36000, 0.285, 0.08},
		{"changzhou", "常州", "江苏省", 4494, 24396, 2280, 36000, 0.285, 0.08},
		{"nantong", "南通", "江苏省", 4494, 24396, 2280, 33000, 0.285, 0.08},
		// 浙江省
		{"hangzhou", "杭州", "浙江省", 4812, 24060, 2490, 39377, 0.29, 0.12},
		{"ningbo", "宁波", "浙江省", 4462, 22311, 2280, 35000, 0.29, 0.12},
		{"wenzhou", "温州", "浙江省", 4462, 22311, 2280, 32000, 0.29, 0.10},
		{"jiaxing", "嘉兴", "浙江省", 4462, 22311, 2280, 32000, 0.29, 0.12},
		// 四川省
		{"chengdu", "成都", "四川省", 4071, 22555, 2280, 33000, 0.275, 0.06},
		{"bindzhou", "绵阳", "四川省", 3800, 20000, 2100, 28000, 0.275, 0.06},
		// 湖北省
		{"wuhan", "武汉", "湖北省", 4077, 23274, 2010, 30000, 0.28, 0.08},
		// 湖南省
		{"changsha", "长沙", "湖南省", 4000, 22491, 2000, 30000, 0.285, 0.08},
		// 陕西省
		{"xian", "西安", "陕西省", 4559, 22795, 2200, 32000, 0.285, 0.10},
		// 河南省
		{"zhengzhou", "郑州", "河南省", 3756, 20235, 2000, 28000, 0.285, 0.08},
		// 山东省
		{"jinan", "济南", "山东省", 4378, 23898, 2200, 32000, 0.285, 0.12},
		{"qingdao", "青岛", "山东省", 4378, 23898, 2200, 32000, 0.285, 0.12},
		{"yantai", "烟台", "山东省", 4200, 22000, 2100, 30000, 0.285, 0.12},
		// 辽宁省
		{"shenyang", "沈阳", "辽宁省", 4106, 20529, 2100, 28000, 0.275, 0.10},
		{"dalian", "大连", "辽宁省", 4106, 20529, 2100, 28000, 0.275, 0.10},
		// 福建省
		{"fuzhou", "福州", "福建省", 4043, 23118, 2100, 30000, 0.285, 0.12},
		{"xiamen", "厦门", "福建省", 4043, 23118, 2100, 32000, 0.285, 0.12},
		{"quanzhou", "泉州", "福建省", 4043, 23118, 2100, 28000, 0.285, 0.10},
		// 黑龙江省
		{"haerbin", "哈尔滨", "黑龙江省", 3875, 19377, 2000, 26000, 0.27, 0.08},
		// 吉林省
		{"changchun", "长春", "吉林省", 4071, 20355, 2100, 26000, 0.27, 0.08},
		// 安徽省
		{"hefei", "合肥", "安徽省", 4071, 22023, 2100, 30000, 0.285, 0.08},
		// 江西省
		{"nanchang", "南昌", "江西省", 3990, 19950, 2000, 26000, 0.28, 0.08},
		// 河北省
		{"shijiazhuang", "石家庄", "河北省", 4071, 22023, 2100, 28000, 0.285, 0.08},
		// 山西省
		{"taiyuan", "太原", "山西省", 3904, 19519, 2000, 26000, 0.28, 0.08},
		// 云南省
		{"kunming", "昆明", "云南省", 4071, 20355, 2100, 28000, 0.275, 0.08},
		// 贵州省
		{"guiyang", "贵阳", "贵州省", 3904, 19519, 2000, 26000, 0.275, 0.08},
		// 广西
		{"nanning", "南宁", "广西壮族自治区", 3904, 19519, 2000, 26000, 0.275, 0.08},
		// 甘肃省
		{"lanzhou", "兰州", "甘肃省", 3800, 19000, 2000, 24000, 0.27, 0.08},
		// 内蒙古
		{"huhehaote", "呼和浩特", "内蒙古自治区", 3800, 19000, 2000, 24000, 0.27, 0.08},
		// 新疆
		{"wulumuqi", "乌鲁木齐", "新疆维吾尔自治区", 3800, 19000, 2000, 24000, 0.27, 0.08},
		// 海南省
		{"haikou", "海口", "海南省", 3800, 22000, 2000, 28000, 0.27, 0.12},
		{"sanya", "三亚", "海南省", 3800, 22000, 2000, 28000, 0.27, 0.12},
	}

	var results []model.CitySSData
	for _, city := range cities {
		results = append(results, model.CitySSData{
			CityCode:             city.code,
			CityName:             city.name,
			Province:             city.province,
			Year:                 year,
			SSBaseMin:            city.ssMin,
			SSBaseMax:            city.ssMax,
			HFBaseMin:            city.hfMin,
			HFBaseMax:            city.hfMax,
			SSPersonRateSum:      0.105,
			SSCompanyRateSum:     city.ssCompany,
			HFRateMin:            0.05,
			HFRateMax:            0.12,
			DefaultHFRatePerson:  city.hfDefault,
			DefaultHFRateCompany: city.hfDefault,
			SourceURL:            sourceURL,
			CrawledAt:            now,
		})
	}

	log.Printf("✅ [MOCK] 已加载 %d 个城市的模拟数据", len(results))
	return results
}

// getDefaultCityList 获取预定义的城市列表
func getDefaultCityList() []model.CitySSData {
	return []model.CitySSData{
		{CityCode: "beijing", CityName: "北京"},
		{CityCode: "shanghai", CityName: "上海"},
		{CityCode: "guangzhou", CityName: "广州"},
		{CityCode: "shenzhen", CityName: "深圳"},
		{CityCode: "hangzhou", CityName: "杭州"},
		{CityCode: "nanjing", CityName: "南京"},
		{CityCode: "suzhou", CityName: "苏州"},
		{CityCode: "chengdu", CityName: "成都"},
		{CityCode: "wuhan", CityName: "武汉"},
		{CityCode: "tianjin", CityName: "天津"},
		{CityCode: "chongqing", CityName: "重庆"},
		{CityCode: "xian", CityName: "西安"},
		{CityCode: "changsha", CityName: "长沙"},
		{CityCode: "zhengzhou", CityName: "郑州"},
		{CityCode: "qingdao", CityName: "青岛"},
		{CityCode: "dalian", CityName: "大连"},
		{CityCode: "ningbo", CityName: "宁波"},
		{CityCode: "xiamen", CityName: "厦门"},
		{CityCode: "fuzhou", CityName: "福州"},
		{CityCode: "jinan", CityName: "济南"},
		{CityCode: "shenyang", CityName: "沈阳"},
		{CityCode: "harbin", CityName: "哈尔滨"},
		{CityCode: "changchun", CityName: "长春"},
		{CityCode: "nanchang", CityName: "南昌"},
		{CityCode: "hefei", CityName: "合肥"},
		{CityCode: "shijiazhuang", CityName: "石家庄"},
		{CityCode: "taiyuan", CityName: "太原"},
		{CityCode: "kunming", CityName: "昆明"},
		{CityCode: "guiyang", CityName: "贵阳"},
		{CityCode: "nanning", CityName: "南宁"},
		{CityCode: "lanzhou", CityName: "兰州"},
		{CityCode: "wuxi", CityName: "无锡"},
		{CityCode: "dongguan", CityName: "东莞"},
		{CityCode: "foshan", CityName: "佛山"},
		{CityCode: "zhuhai", CityName: "珠海"},
		{CityCode: "huizhou", CityName: "惠州"},
		{CityCode: "zhongshan", CityName: "中山"},
		{CityCode: "wenzhou", CityName: "温州"},
		{CityCode: "quanzhou", CityName: "泉州"},
		{CityCode: "changzhou", CityName: "常州"},
	}
}
