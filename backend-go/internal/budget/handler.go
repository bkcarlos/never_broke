package budget

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 预算处理器
type Handler struct{}

// NewHandler 创建预算处理器
func NewHandler() *Handler {
	return &Handler{}
}

// GetMonthlyBudgetHandler 获取月度预算
func (h *Handler) GetMonthlyBudgetHandler(c *gin.Context) {
	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "month parameter is required")
		return
	}

	budget, exists := GetMonthlyBudget(month)
	if !exists {
		// 无记录时返回 totalBudget=0
		budget = MonthlyBudget{
			Month:       month,
			TotalBudget: 0,
		}
	}

	response.OK(c, budget)
}

// PutMonthlyBudgetHandler 更新月度预算
func (h *Handler) PutMonthlyBudgetHandler(c *gin.Context) {
	var req MonthlyBudget
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	// 校验 month 格式非空
	if req.Month == "" {
		response.BadRequest(c, "month is required")
		return
	}

	// 校验 totalBudget >= 0
	if req.TotalBudget < 0 {
		response.BadRequest(c, "totalBudget must be >= 0")
		return
	}

	// 写入内存
	result := UpsertMonthlyBudget(req)

	logger.Info("Monthly budget updated", zap.String("month", result.Month), zap.Float64("totalBudget", result.TotalBudget))
	response.OK(c, result)
}




