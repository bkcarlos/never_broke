package cashflow

import (
	"gorm.io/gorm"
	"math"
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/repository"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ForecastHandler 现金流预测处理器
func ForecastHandler(store repository.Store, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDVal, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			return
		}
		userID := userIDVal.(string)

		// 从 query 参数读取
		startDate := c.Query("startDate")
		if startDate == "" {
			response.BadRequest(c, "startDate is required")
			return
		}

		daysStr := c.Query("days")
		if daysStr == "" {
			daysStr = "30" // 默认 30 天
		}
		days, err := strconv.Atoi(daysStr)
		if err != nil || days <= 0 || days > 366 {
			response.BadRequest(c, "invalid days parameter")
			return
		}

		baseStr := c.Query("base")
		if baseStr == "" {
			baseStr = string(ForecastBaseTotalAssets) // 默认总资产
		}
		base := ForecastBase(baseStr)
		if base != ForecastBaseTotalAssets && base != ForecastBaseAccount {
			response.BadRequest(c, "invalid base parameter")
			return
		}

		accountID := c.Query("accountId")
		if base == ForecastBaseAccount && accountID == "" {
			response.BadRequest(c, "accountId is required when base is account")
			return
		}

		safeAmountStr := c.Query("safeAmount")
		if safeAmountStr == "" {
			safeAmountStr = "2000" // 默认安全垫 2000
		}
		safeAmount, err := strconv.ParseFloat(safeAmountStr, 64)
		if err != nil || safeAmount < 0 || math.IsNaN(safeAmount) || math.IsInf(safeAmount, 0) {
			response.BadRequest(c, "invalid safeAmount parameter")
			return
		}

		req := CashflowForecastRequest{
			StartDate:  startDate,
			Days:       days,
			Base:       base,
			AccountID:  accountID,
			SafeAmount: safeAmount,
		}

		if _, err := validateRequest(req); err != nil {
			response.BadRequest(c, err.Error())
			return
		}

		logger.Debug("Building cashflow forecast",
			zap.String("userID", userID),
			zap.String("startDate", startDate),
			zap.Int("days", days),
			zap.String("base", string(base)),
		)

		var result CashflowForecastResponse
		if db != nil {
			result, err = BuildForecastWithDB(c.Request.Context(), store, db, userID, req)
		} else {
			result, err = BuildForecast(c.Request.Context(), store, userID, req)
		}
		if err != nil {
			logger.Error("Failed to build forecast",
				zap.String("userID", userID),
				zap.Error(err),
			)
			response.InternalError(c, "Failed to build forecast: "+err.Error())
			return
		}

		logger.Info("Cashflow forecast built successfully",
			zap.String("userID", userID),
			zap.String("status", string(result.Status)),
			zap.Int("pointsCount", len(result.Points)),
		)
		response.OK(c, result)
	}
}
