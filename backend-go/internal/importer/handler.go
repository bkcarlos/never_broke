package importer

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 导入处理器
type Handler struct{}

// NewHandler 创建导入处理器
func NewHandler() *Handler {
	return &Handler{}
}

// PreviewHandler 预览导入
func (h *Handler) PreviewHandler(c *gin.Context) {
	var req ImportPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.Raw == "" {
		response.BadRequest(c, "raw is required")
		return
	}

	// 简单解析每一行
	lines := strings.Split(req.Raw, "\n")
	items := []ImportPreviewItem{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 按逗号切分
		parts := strings.Split(line, ",")
		if len(parts) < 3 {
			continue
		}

		date := strings.TrimSpace(parts[0])
		remark := strings.TrimSpace(parts[1])
		amountStr := strings.TrimSpace(parts[2])

		amount, err := strconv.ParseFloat(amountStr, 64)
		if err != nil {
			logger.Warn("Failed to parse amount", zap.String("amount", amountStr), zap.Error(err))
			continue
		}

		// 简单关键词匹配分类
		categoryGuess := guessCategory(remark)

		items = append(items, ImportPreviewItem{
			Date:          date,
			CategoryGuess: categoryGuess,
			Amount:        amount,
			Remark:        remark,
		})
	}

	result := ImportPreviewResponse{
		Count: len(items),
		Items: items,
	}

	logger.Info("Import preview generated", zap.Int("count", result.Count))
	response.OK(c, result)
}

// guessCategory 根据备注猜测分类
func guessCategory(remark string) string {
	remark = strings.ToLower(remark)

	// 简单关键词匹配
	if strings.Contains(remark, "早餐") || strings.Contains(remark, "午餐") ||
		strings.Contains(remark, "晚餐") || strings.Contains(remark, "吃饭") ||
		strings.Contains(remark, "餐厅") || strings.Contains(remark, "外卖") {
		return "FOOD_DINING"
	}

	if strings.Contains(remark, "地铁") || strings.Contains(remark, "公交") ||
		strings.Contains(remark, "打车") || strings.Contains(remark, "交通") ||
		strings.Contains(remark, "uber") || strings.Contains(remark, "滴滴") {
		return "TRANSPORT"
	}

	if strings.Contains(remark, "超市") || strings.Contains(remark, "购物") ||
		strings.Contains(remark, "商场") {
		return "SHOPPING"
	}

	if strings.Contains(remark, "电影") || strings.Contains(remark, "娱乐") ||
		strings.Contains(remark, "游戏") {
		return "ENTERTAINMENT"
	}

	// 默认返回 OTHER
	return "OTHER"
}







