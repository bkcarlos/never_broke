# 社保数据爬取工具

从 [m12333.cn](https://m12333.cn/) 爬取全国各城市社保基数数据并同步到 NeverBroke 数据库。

## ⚠️ 重要声明

### 数据来源

- **数据来源**: https://m12333.cn/
- m12333.cn 汇集了全国各城市的社保政策信息

### 免责声明

1. **数据仅供参考**：不构成任何官方建议，请以各地人社局最新政策为准
2. **数据可能滞后**：社保基数通常每年7月更新
3. **版权声明**：如有侵权问题，请联系我们，我们将立即删除相关数据

### 官方数据来源

如需确认数据准确性，请访问各城市人力资源和社会保障局官网：

| 城市 | 官方网站 |
|------|---------|
| 北京 | http://rsj.beijing.gov.cn/ |
| 上海 | http://rsj.sh.gov.cn/ |
| 广州 | http://rsj.gz.gov.cn/ |
| 深圳 | http://hrss.sz.gov.cn/ |
| 杭州 | http://hrss.hangzhou.gov.cn/ |
| 南京 | http://rsj.nanjing.gov.cn/ |
| 成都 | http://cdhrss.chengdu.gov.cn/ |
| 武汉 | http://rsj.wuhan.gov.cn/ |
| 天津 | http://hrss.tj.gov.cn/ |
| 重庆 | http://rlsbj.cq.gov.cn/ |

## 安装

### 前置依赖

需要安装 Chrome/Chromium 浏览器（用于无头浏览器爬取）：

```bash
# macOS
brew install --cask chromium

# Ubuntu/Debian
apt-get install chromium-browser

# CentOS/RHEL
yum install chromium
```

### 构建

```bash
cd social-data-tool
go mod tidy
go build -o social-data-tool
```

## 使用方法

### Mock 模式（推荐）- 模拟爬取 m12333.cn/cat/fee

```bash
# 使用 mock 模式生成数据（48个城市）
./social-data-tool -mock -output data/cities_2025.json

# Mock + 同步到数据库
./social-data-tool -mock -sync
```

### 后备数据模式

```bash
./social-data-tool -fallback -output data/cities_2025.json
```

### 实际爬取（可能失败）

```bash
# 尝试真实爬取，失败自动回退
./social-data-tool -output data/cities_2025.json
```

### 同步到数据库

```bash
export NEVERBROKE_PG_DSN="postgres://user:pass@localhost:5432/neverbroke?sslmode=disable"
./social-data-tool -mock -sync
```

### 查看数据来源说明

```bash
./social-data-tool -list-sources
```

## 命令行参数

| 参数 | 说明 |
|------|------|
| `-mock` | **推荐** Mock 模式，模拟爬取 m12333.cn/cat/fee |
| `-fallback` | 使用后备数据（40个城市） |
| `-analyze <url>` | 分析页面请求和数据结构 |
| `-output <file>` | 输出 JSON 文件路径 |
| `-sync` | 同步数据到数据库 |
| `-dsn <dsn>` | PostgreSQL 连接字符串 |
| `-dry-run` | 仅预览，不实际执行 |
| `-list-sources` | 显示数据来源说明 |

## 页面分析功能

使用 `-analyze` 参数可以分析任意页面的数据结构：

```bash
# 分析 m12333.cn/cat/fee 页面
./social-data-tool -analyze "https://m12333.cn/cat/fee"

# 输出示例：
# 📊 城市信息:
#    • 城市: 北京
# 📊 缴费基数:
#    • 缴费基数上限: 35811
#    • 缴费基数下限: 7162
# 📊 养老保险:
#    • 养老保险比例: [16]
```

这个功能可以：
- 捕获浏览器发送的所有 API 请求
- 分析页面 HTML 内容
- 自动提取社保相关的数值数据

## 数据格式

输出的 JSON 文件格式：

```json
{
  "version": "2025.1",
  "lastUpdated": "2025-01-15",
  "dataSource": "https://m12333.cn/",
  "disclaimer": "数据仅供参考...",
  "cities": [
    {
      "cityCode": "beijing",
      "cityName": "北京",
      "year": 2025,
      "ssBaseMin": 6326,
      "ssBaseMax": 35283,
      "hfBaseMin": 2500,
      "hfBaseMax": 35283,
      "sourceUrl": "https://m12333.cn/city/beijing/",
      "crawledAt": "2025-01-15T10:30:00Z"
    }
  ]
}
```

## 数据更新建议

- 建议每年 **7-8 月** 运行一次，获取新年度社保基数
- 运行前请先用 `-output` 保存 JSON，确认数据正确后再 `-sync`
- 定期检查官方网站更新情况

## 技术说明

- 使用 [chromedp](https://github.com/chromedp/chromedp) 无头浏览器爬取
- 支持 JavaScript 渲染的页面
- 自动处理反爬保护

## 联系方式

如发现数据错误或有版权问题，请提交 Issue。

## License

本工具仅供学习和个人使用。
