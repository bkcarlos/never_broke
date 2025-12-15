package reports

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/repository"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// MonthlySummaryHandler 月度汇总报表处理器
func MonthlySummaryHandler(store repository.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 当前使用固定用户 ID
		userID := "demo"

		// 从 query 读取 year, month（默认当前年月）
		yearStr := c.Query("year")
		monthStr := c.Query("month")

		now := time.Now()
		year := now.Year()
		month := int(now.Month())

		if yearStr != "" {
			if y, err := strconv.Atoi(yearStr); err == nil {
				year = y
			}
		}
		if monthStr != "" {
			if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
				month = m
			}
		}

		logger.Debug("Building monthly summary",
			zap.String("userID", userID),
			zap.Int("year", year),
			zap.Int("month", month),
		)

		result, err := BuildMonthlySummary(c.Request.Context(), store, userID, year, month)
		if err != nil {
			logger.Error("Failed to build monthly summary",
				zap.String("userID", userID),
				zap.Error(err),
			)
			response.InternalError(c, "Failed to build monthly summary: "+err.Error())
			return
		}

		logger.Info("Monthly summary built successfully",
			zap.String("userID", userID),
			zap.Float64("totalIncome", result.TotalIncome),
			zap.Float64("totalExpense", result.TotalExpense),
		)
		response.OK(c, result)
	}
}
