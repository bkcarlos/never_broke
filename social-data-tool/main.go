package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"social-data-tool/crawler"
	"social-data-tool/db"
	"social-data-tool/model"
)

func main() {
	// 命令行参数
	dsn := flag.String("dsn", "", "PostgreSQL DSN (或设置 NEVERBROKE_PG_DSN 环境变量)")
	outputJSON := flag.String("output", "", "输出到 JSON 文件")
	syncDB := flag.Bool("sync", false, "同步到数据库")
	listSources := flag.Bool("list-sources", false, "显示数据来源说明")
	dryRun := flag.Bool("dry-run", false, "仅显示将要执行的操作")
	useFallback := flag.Bool("fallback", false, "直接使用后备数据（跳过爬虫）")
	useMock := flag.Bool("mock", false, "使用 mock 模式（模拟爬取 m12333.cn/cat/fee）")
	analyzeURL := flag.String("analyze", "", "分析指定 URL 的请求和数据结构")
	flag.Parse()

	// 分析模式
	if *analyzeURL != "" {
		runAnalyzer(*analyzeURL)
		return
	}

	if *listSources {
		printDataSources()
		return
	}

	// 获取 DSN
	dbDSN := *dsn
	if dbDSN == "" {
		dbDSN = os.Getenv("NEVERBROKE_PG_DSN")
	}

	// 打印横幅
	printBanner()

	var data []model.CitySSData

	if *useMock {
		// Mock 模式：模拟从 m12333.cn/cat/fee 爬取
		c := crawler.NewM12333CrawlerMock()
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		data, _ = c.CrawlAllCities(ctx)
	} else if *useFallback {
		// 直接使用后备数据
		fmt.Println("📋 使用后备数据源（基于官方来源手动整理）...")
		data = crawler.GetFallbackData()
		fmt.Printf("✅ 已加载 %d 个城市的后备数据\n\n", len(data))
	} else {
		// 创建爬虫
		c := crawler.NewM12333Crawler()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		// 爬取数据
		var err error
		data, err = c.CrawlAllCities(ctx)
		if err != nil {
			log.Printf("⚠️ 爬取失败: %v", err)
		}

		// 如果爬取失败或数据为空，使用后备数据
		if len(data) == 0 {
			fmt.Println("📋 爬取失败，使用后备数据源...")
			data = crawler.GetFallbackData()
			fmt.Printf("✅ 已加载 %d 个城市的后备数据\n\n", len(data))
		}
	}

	fmt.Printf("\n✅ 成功爬取 %d 个城市的社保数据\n\n", len(data))

	// 显示数据摘要
	printDataSummary(data)

	if *dryRun {
		fmt.Println("\n[DRY-RUN] 以上为将要处理的数据，未实际执行")
		return
	}

	// 输出到 JSON 文件
	if *outputJSON != "" {
		if err := saveToJSON(*outputJSON, data); err != nil {
			log.Fatalf("写入文件失败: %v", err)
		}
		fmt.Printf("📁 数据已保存到: %s\n", *outputJSON)
	}

	// 同步到数据库
	if *syncDB {
		if dbDSN == "" {
			log.Fatal("请提供数据库连接: -dsn 或设置 NEVERBROKE_PG_DSN")
		}

		syncCtx, syncCancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer syncCancel()

		if err := db.SyncToDatabase(syncCtx, dbDSN, data); err != nil {
			log.Fatalf("同步数据库失败: %v", err)
		}
		fmt.Println("✅ 数据已同步到数据库")
	}

	// 如果没有指定输出，默认输出到 data 目录
	if *outputJSON == "" && !*syncDB {
		defaultOutput := fmt.Sprintf("data/cities_%d.json", time.Now().Year())
		if err := saveToJSON(defaultOutput, data); err != nil {
			log.Fatalf("写入文件失败: %v", err)
		}
		fmt.Printf("📁 数据已保存到: %s\n", defaultOutput)
	}
}

func printBanner() {
	fmt.Println(`
═══════════════════════════════════════════════════════════════
          m12333.cn 社保数据爬取工具
═══════════════════════════════════════════════════════════════

⚠️  数据来源: https://m12333.cn/
⚠️  仅供参考，请以当地人社局官方数据为准
⚠️  如有侵权问题，请联系我们删除相关数据
`)
}

func printDataSummary(data []model.CitySSData) {
	fmt.Println("┌────────────────────────────────────────────────────────────┐")
	fmt.Println("│                    数据摘要                                │")
	fmt.Println("├────────────────────────────────────────────────────────────┤")
	fmt.Printf("│ 城市数量: %-48d │\n", len(data))
	fmt.Println("├────────────────────────────────────────────────────────────┤")
	fmt.Println("│ 城市名称          社保基数下限      社保基数上限           │")
	fmt.Println("├────────────────────────────────────────────────────────────┤")

	for i, city := range data {
		if i >= 10 {
			fmt.Printf("│ ... 还有 %d 个城市                                         │\n", len(data)-10)
			break
		}
		fmt.Printf("│ %-16s  %-16.0f  %-16.0f       │\n",
			city.CityName, city.SSBaseMin, city.SSBaseMax)
	}
	fmt.Println("└────────────────────────────────────────────────────────────┘")
}

func saveToJSON(filepath string, data []model.CitySSData) error {
	output := model.NewDataFile(data)
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, jsonData, 0644)
}

func runAnalyzer(url string) {
	fmt.Println(`
═══════════════════════════════════════════════════════════════
          m12333.cn 页面分析工具
═══════════════════════════════════════════════════════════════
`)
	fmt.Printf("🔍 正在分析: %s\n", url)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	analyzer := crawler.NewPageAnalyzer(url)
	html, requests, err := analyzer.AnalyzePage(ctx)
	if err != nil {
		log.Printf("❌ 分析失败: %v", err)
		return
	}

	analyzer.PrintAnalysis(html, requests)

	// 尝试提取 JSON 数据
	jsonData := crawler.ExtractJSONData(html)
	if len(jsonData) > 0 {
		fmt.Printf("\n📊 发现 %d 条 JSON 数据\n", len(jsonData))
	}

	// 提取结构化数据
	extractedData := crawler.ExtractAllDataFromHTML(html)
	crawler.PrintExtractedData(extractedData)
}

func printDataSources() {
	fmt.Println(`
═══════════════════════════════════════════════════════════════════════════════
                         社保数据来源声明
═══════════════════════════════════════════════════════════════════════════════

数据来源
────────
本工具从 https://m12333.cn/ 爬取社保基数数据。
m12333.cn 汇集了全国各城市的社保政策信息。

免责声明
────────
1. 数据仅供参考，不构成任何官方建议
2. 请以各地人力资源和社会保障局的官方公布为准
3. 社保基数通常每年7月更新，数据可能存在滞后

版权说明
────────
• 如有版权问题，请联系我们，我们将立即删除相关数据
• 本工具仅供学习和个人使用

官方数据查询渠道
─────────────────
• 北京: http://rsj.beijing.gov.cn/
• 上海: http://rsj.sh.gov.cn/
• 广州: http://rsj.gz.gov.cn/
• 深圳: http://hrss.sz.gov.cn/
• 杭州: http://hrss.hangzhou.gov.cn/
• 南京: http://rsj.nanjing.gov.cn/
• 成都: http://cdhrss.chengdu.gov.cn/
• 武汉: http://rsj.wuhan.gov.cn/
• 天津: http://hrss.tj.gov.cn/
• 重庆: http://rlsbj.cq.gov.cn/

更多城市请访问当地人社局官网查询。

═══════════════════════════════════════════════════════════════════════════════
`)
}
