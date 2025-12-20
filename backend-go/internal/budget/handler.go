package budget

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Handler 预算处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建预算处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// GetMonthlyBudgetHandler 获取月度预算
func (h *Handler) GetMonthlyBudgetHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	month := c.Query("month")
	if month == "" {
		response.BadRequest(c, "month parameter is required")
		return
	}

	var dbBudget models.Budget
	err := h.db.Where("user_id = ? AND month = ?", userID.(string), month).First(&dbBudget).Error

	if err == gorm.ErrRecordNotFound {
		// 无记录时返回 totalBudget=0
		response.OK(c, MonthlyBudget{
			Month:       month,
			TotalBudget: 0,
		})
		return
	}
	if err != nil {
		logger.Error("Failed to get budget", zap.Error(err))
		response.InternalError(c, "failed to get budget")
		return
	}

	response.OK(c, MonthlyBudget{
		Month:       dbBudget.Month,
		TotalBudget: dbBudget.TotalBudget,
	})
}

// PutMonthlyBudgetHandler 更新月度预算
func (h *Handler) PutMonthlyBudgetHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

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

	// 使用 upsert 写入数据库
	dbBudget := &models.Budget{
		ID:          uuid.NewString(),
		UserID:      userID.(string),
		Month:       req.Month,
		TotalBudget: req.TotalBudget,
	}

	err := h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "month"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_budget", "updated_at"}),
	}).Create(dbBudget).Error

	if err != nil {
		logger.Error("Failed to upsert budget", zap.Error(err))
		response.InternalError(c, "failed to save budget")
		return
	}

	logger.Info("Monthly budget updated",
		zap.String("userID", userID.(string)),
		zap.String("month", req.Month),
		zap.Float64("totalBudget", req.TotalBudget),
	)
	response.OK(c, req)
}
