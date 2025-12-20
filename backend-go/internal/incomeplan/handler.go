package incomeplan

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 收入计划处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建收入计划处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// ListIncomePlansHandler 列出所有收入计划
func (h *Handler) ListIncomePlansHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var dbPlans []models.IncomePlan
	err := h.db.Where("user_id = ?", userID.(string)).
		Order("active DESC, created_at DESC").
		Find(&dbPlans).Error

	if err != nil {
		logger.Error("Failed to list income plans", zap.Error(err))
		response.InternalError(c, "failed to list income plans")
		return
	}

	// 转换为响应格式
	result := make([]IncomePlan, len(dbPlans))
	for i, p := range dbPlans {
		result[i] = IncomePlan{
			ID:           p.ID,
			Name:         p.Name,
			Amount:       p.Amount,
			Frequency:    IncomeFrequency(p.Frequency),
			AccountID:    p.AccountID,
			FirstPayDate: p.FirstPayDate.Format("2006-01-02"),
			Active:       p.Active,
		}
		if p.EndDate != nil && !p.EndDate.IsZero() {
			result[i].EndDate = p.EndDate.Format("2006-01-02")
		}
	}

	response.OK(c, result)
}

// CreateIncomePlanHandler 创建收入计划
func (h *Handler) CreateIncomePlanHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

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

	// 解析日期
	firstPayDate, err := time.Parse("2006-01-02", req.FirstPayDate)
	if err != nil {
		response.BadRequest(c, "invalid firstPayDate format, expected YYYY-MM-DD")
		return
	}

	// 创建数据库记录
	dbPlan := &models.IncomePlan{
		ID:           uuid.NewString(),
		UserID:       userID.(string),
		Name:         req.Name,
		Amount:       req.Amount,
		Frequency:    string(req.Frequency),
		AccountID:    req.AccountID,
		FirstPayDate: firstPayDate,
		Active:       true,
	}

	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			response.BadRequest(c, "invalid endDate format, expected YYYY-MM-DD")
			return
		}
		dbPlan.EndDate = &endDate
	}

	if err := h.db.Create(dbPlan).Error; err != nil {
		logger.Error("Failed to create income plan", zap.Error(err))
		response.InternalError(c, "failed to create income plan")
		return
	}

	// 返回响应
	plan := IncomePlan{
		ID:           dbPlan.ID,
		Name:         dbPlan.Name,
		Amount:       dbPlan.Amount,
		Frequency:    IncomeFrequency(dbPlan.Frequency),
		AccountID:    dbPlan.AccountID,
		FirstPayDate: dbPlan.FirstPayDate.Format("2006-01-02"),
		EndDate:      req.EndDate,
		Active:       dbPlan.Active,
	}

	logger.Info("Income plan created",
		zap.String("userID", userID.(string)),
		zap.String("id", plan.ID),
		zap.String("name", plan.Name),
	)
	response.OK(c, plan)
}
