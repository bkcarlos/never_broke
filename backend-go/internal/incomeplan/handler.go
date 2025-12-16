package incomeplan

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler 收入计划处理器
type Handler struct{}

// NewHandler 创建收入计划处理器
func NewHandler() *Handler {
	return &Handler{}
}

// ListIncomePlansHandler 列出所有收入计划
func (h *Handler) ListIncomePlansHandler(c *gin.Context) {
	list := ListPlans()
	response.OK(c, list)
}

// CreateIncomePlanHandler 创建收入计划
func (h *Handler) CreateIncomePlanHandler(c *gin.Context) {
	var req CreateIncomePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	// 校验必填字段
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	if req.Amount <= 0 {
		response.BadRequest(c, "amount must be > 0")
		return
	}

	validFrequencies := map[IncomeFrequency]bool{
		FrequencyMonthly:  true,
		FrequencyWeekly:   true,
		FrequencyBiweekly: true,
		FrequencyYearly:   true,
	}
	if !validFrequencies[req.Frequency] {
		response.BadRequest(c, "invalid frequency")
		return
	}

	if req.AccountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}

	if req.FirstPayDate == "" {
		response.BadRequest(c, "firstPayDate is required")
		return
	}

	// 生成 ID
	id := uuid.NewString()

	// 创建收入计划
	plan := IncomePlan{
		ID:           id,
		Name:         req.Name,
		Amount:       req.Amount,
		Frequency:    req.Frequency,
		AccountID:    req.AccountID,
		FirstPayDate: req.FirstPayDate,
		EndDate:      req.EndDate,
		Active:       true,
	}

	// 添加到存储
	result := AddPlan(plan)

	logger.Info("Income plan created", zap.String("id", result.ID), zap.String("name", result.Name))
	response.OK(c, result)
}




