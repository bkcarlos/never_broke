package payroll

import (
	"errors"
	"net/http"
	"strconv"

	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler 薪资计算处理器
type Handler struct {
	service *Service
}

// NewHandler 创建处理器
func NewHandler(db *gorm.DB) *Handler {
	repo := NewRepository(db)
	service := NewService(repo)
	return &Handler{service: service}
}

// ListRegions 获取地区列表
// GET /api/v1/payroll/regions
func (h *Handler) ListRegions(c *gin.Context) {
	regions, err := h.service.ListRegions()
	if err != nil {
		response.InternalError(c, "Failed to fetch regions")
		return
	}
	response.OK(c, regions)
}

// GetRegionRule 获取地区规则
// GET /api/v1/payroll/regions/:code/rule
func (h *Handler) GetRegionRule(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, "region code is required")
		return
	}

	yearStr := c.DefaultQuery("year", "2025")
	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 2020 || year > 2030 {
		h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, "invalid year")
		return
	}

	rule, err := h.service.GetRegionRule(code, year)
	if err != nil {
		if errors.Is(err, ErrRuleNotFound) {
			h.sendError(c, http.StatusNotFound, ErrorCodeNotFound, "region rule not found")
			return
		}
		response.InternalError(c, "Failed to fetch region rule")
		return
	}

	response.OK(c, rule)
}

// CalcSalary 计算薪资
// POST /api/v1/payroll/calc/salary
func (h *Handler) CalcSalary(c *gin.Context) {
	var req CalcSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		return
	}

	result, err := h.service.CalculateSalary(&req)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidGrossMonthly):
			h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		case errors.Is(err, ErrCustomBaseRequired):
			h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		case errors.Is(err, ErrCustomBaseMustBePositive):
			h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		case errors.Is(err, ErrHFRateOutOfRange):
			h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		case errors.Is(err, ErrSSRateOutOfRange):
			h.sendError(c, http.StatusBadRequest, ErrorCodeInvalidArgument, err.Error())
		case errors.Is(err, ErrRuleNotFound):
			h.sendError(c, http.StatusNotFound, ErrorCodeNotFound, "region rule not found")
		case errors.Is(err, ErrTaxBracketsNotFound):
			h.sendError(c, http.StatusNotFound, ErrorCodeNotFound, "tax brackets not found for year")
		default:
			response.InternalError(c, "Failed to calculate salary")
		}
		return
	}

	response.OK(c, result)
}

// sendError 发送错误响应
func (h *Handler) sendError(c *gin.Context, status int, code ErrorCode, message string) {
	c.JSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": message,
		},
	})
}
