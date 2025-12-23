package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// NetworkRequest 网络请求记录
type NetworkRequest struct {
	URL          string            `json:"url"`
	Method       string            `json:"method"`
	Headers      map[string]string `json:"headers,omitempty"`
	PostData     string            `json:"postData,omitempty"`
	ResponseBody string            `json:"responseBody,omitempty"`
	StatusCode   int               `json:"statusCode,omitempty"`
}

// PageAnalyzer 页面分析器
type PageAnalyzer struct {
	targetURL string
	requests  []NetworkRequest
}

// NewPageAnalyzer 创建页面分析器
func NewPageAnalyzer(url string) *PageAnalyzer {
	return &PageAnalyzer{
		targetURL: url,
		requests:  make([]NetworkRequest, 0),
	}
}

// AnalyzePage 分析页面，捕获所有网络请求
func (a *PageAnalyzer) AnalyzePage(ctx context.Context) (string, []NetworkRequest, error) {
	log.Printf("🔍 开始分析页面: %s", a.targetURL)

	// 创建 chromedp 上下文
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer browserCancel()

	// 设置超时
	browserCtx, cancel := context.WithTimeout(browserCtx, 60*time.Second)
	defer cancel()

	// 存储请求和响应
	requests := make(map[network.RequestID]*NetworkRequest)

	// 监听网络事件
	chromedp.ListenTarget(browserCtx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			req := &NetworkRequest{
				URL:    e.Request.URL,
				Method: e.Request.Method,
			}
			if e.Request.PostData != "" {
				req.PostData = e.Request.PostData
			}
			requests[e.RequestID] = req

		case *network.EventResponseReceived:
			if req, ok := requests[e.RequestID]; ok {
				req.StatusCode = int(e.Response.Status)
			}
		}
	})

	var pageHTML string

	// 执行浏览器操作
	err := chromedp.Run(browserCtx,
		network.Enable(),
		chromedp.Navigate(a.targetURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.Sleep(8*time.Second), // 等待页面完全加载
		chromedp.OuterHTML("html", &pageHTML, chromedp.ByQuery),
	)

	if err != nil {
		return "", nil, fmt.Errorf("页面加载失败: %w", err)
	}

	// 整理请求列表
	var result []NetworkRequest
	for _, req := range requests {
		// 过滤掉静态资源
		if isAPIRequest(req.URL) {
			result = append(result, *req)
		}
	}

	log.Printf("✅ 捕获到 %d 个 API 请求", len(result))
	a.requests = result

	return pageHTML, result, nil
}

// isAPIRequest 判断是否为 API 请求
func isAPIRequest(url string) bool {
	// 排除静态资源
	excludeExts := []string{".js", ".css", ".png", ".jpg", ".gif", ".ico", ".woff", ".ttf", ".svg"}
	for _, ext := range excludeExts {
		if strings.HasSuffix(strings.ToLower(url), ext) {
			return false
		}
	}

	// 排除 CDN 和第三方
	excludeDomains := []string{"google", "facebook", "baidu", "cnzz", "51.la", "hm.baidu"}
	for _, domain := range excludeDomains {
		if strings.Contains(strings.ToLower(url), domain) {
			return false
		}
	}

	// 包含可能的 API 路径
	apiPatterns := []string{"/api/", "/data/", "/json", ".json", "/cat/", "/city/", "fee"}
	for _, pattern := range apiPatterns {
		if strings.Contains(strings.ToLower(url), pattern) {
			return true
		}
	}

	return false
}

// ExtractHTMLSamples 提取 HTML 中可能包含数据的片段
func ExtractHTMLSamples(html string) []string {
	var samples []string

	// 查找包含关键词的上下文
	keywords := []string{"社保基数", "缴费基数", "养老", "医疗", "失业", "公积金", "下限", "上限"}

	for _, keyword := range keywords {
		idx := strings.Index(html, keyword)
		if idx != -1 {
			// 提取关键词前后 200 个字符
			start := idx - 100
			if start < 0 {
				start = 0
			}
			end := idx + len(keyword) + 200
			if end > len(html) {
				end = len(html)
			}
			sample := html[start:end]
			// 清理 HTML 标签，只保留文本
			sample = cleanHTML(sample)
			if len(sample) > 20 {
				samples = append(samples, fmt.Sprintf("[%s] %s", keyword, sample))
			}
		}
	}

	return samples
}

// cleanHTML 简单清理 HTML 标签
func cleanHTML(html string) string {
	// 移除 script 和 style 标签
	result := html

	// 简单的标签替换
	result = strings.ReplaceAll(result, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n")
	result = strings.ReplaceAll(result, "</div>", " ")
	result = strings.ReplaceAll(result, "</span>", " ")
	result = strings.ReplaceAll(result, "</li>", "\n")
	result = strings.ReplaceAll(result, "</td>", " | ")
	result = strings.ReplaceAll(result, "</tr>", "\n")

	// 移除所有 HTML 标签
	inTag := false
	var cleaned strings.Builder
	for _, r := range result {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			cleaned.WriteRune(r)
		}
	}

	// 清理多余空白
	result = cleaned.String()
	result = strings.Join(strings.Fields(result), " ")

	return result
}

// PrintAnalysis 打印分析结果
func (a *PageAnalyzer) PrintAnalysis(html string, requests []NetworkRequest) {
	fmt.Println("\n═══════════════════════════════════════════════════════════════")
	fmt.Println("                    页面分析结果")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Printf("\n📍 目标 URL: %s\n", a.targetURL)
	fmt.Printf("📄 页面大小: %d 字节\n", len(html))
	fmt.Printf("🌐 捕获请求: %d 个\n", len(requests))

	if len(requests) > 0 {
		fmt.Println("\n─────────────────────────────────────────────────────────────────")
		fmt.Println("                    API 请求列表")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for i, req := range requests {
			fmt.Printf("\n[%d] %s %s\n", i+1, req.Method, req.URL)
			if req.StatusCode > 0 {
				fmt.Printf("    状态码: %d\n", req.StatusCode)
			}
			if req.PostData != "" {
				fmt.Printf("    POST数据: %s\n", truncate(req.PostData, 200))
			}
		}
	}

	// 分析页面内容中的数据
	fmt.Println("\n─────────────────────────────────────────────────────────────────")
	fmt.Println("                    页面内容分析")
	fmt.Println("─────────────────────────────────────────────────────────────────")

	// 搜索可能的数据模式
	patterns := []string{
		"社保基数",
		"缴费基数",
		"养老保险",
		"医疗保险",
		"失业保险",
		"公积金",
		"下限",
		"上限",
	}

	for _, pattern := range patterns {
		count := strings.Count(html, pattern)
		if count > 0 {
			fmt.Printf("  ✓ 包含 \"%s\": %d 次\n", pattern, count)
		}
	}

	// 检查 JSON 数据
	if strings.Contains(html, `"cityCode"`) || strings.Contains(html, `"cityName"`) {
		fmt.Println("\n  ⚠️ 发现可能的城市数据结构")
	}

	// 提取 HTML 样本
	samples := ExtractHTMLSamples(html)
	if len(samples) > 0 {
		fmt.Println("\n─────────────────────────────────────────────────────────────────")
		fmt.Println("                    数据样本 (清理后)")
		fmt.Println("─────────────────────────────────────────────────────────────────")
		for i, sample := range samples {
			if i >= 8 {
				fmt.Printf("\n  ... 还有 %d 个样本\n", len(samples)-8)
				break
			}
			fmt.Printf("\n  %s\n", truncate(sample, 300))
		}
	}

	fmt.Println("\n═══════════════════════════════════════════════════════════════")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ExtractJSONData 尝试从页面提取 JSON 数据
func ExtractJSONData(html string) []map[string]interface{} {
	var results []map[string]interface{}

	// 简单的 JSON 提取逻辑 - 查找数组形式的 JSON
	start := strings.Index(html, "[{")
	if start != -1 {
		end := strings.Index(html[start:], "}]")
		if end != -1 {
			jsonStr := html[start : start+end+2]
			var data []map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &data); err == nil {
				results = append(results, data...)
			}
		}
	}

	return results
}
