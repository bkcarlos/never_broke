package timeline

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/repository"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DailySummaryHandler 每日汇总处理器
func DailySummaryHandler(store repository.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDVal, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			return
		}
		userID := userIDVal.(string)

		fromDate := c.Query("from")
		if fromDate == "" {
			response.BadRequest(c, "from parameter is required")
			return
		}

		toDate := c.Query("to")
		if toDate == "" {
			response.BadRequest(c, "to parameter is required")
			return
		}

		modeStr := c.Query("mode")
		if modeStr == "" {
			modeStr = string(TimelineModeHistory) // 默认历史模式
		}
		mode := TimelineMode(modeStr)
		if mode != TimelineModeHistory && mode != TimelineModeWithFuture {
			response.BadRequest(c, "invalid mode parameter")
			return
		}

		logger.Debug("Building daily summary",
			zap.String("userID", userID),
			zap.String("fromDate", fromDate),
			zap.String("toDate", toDate),
			zap.String("mode", string(mode)),
		)

		result, err := BuildDailySummary(c.Request.Context(), store, userID, fromDate, toDate, mode)
		if err != nil {
			logger.Error("Failed to build daily summary",
				zap.String("userID", userID),
				zap.Error(err),
			)
			response.InternalError(c, "Failed to build daily summary: "+err.Error())
			return
		}

		logger.Info("Daily summary built successfully",
			zap.String("userID", userID),
			zap.Int("daysCount", len(result.Days)),
		)
		response.OK(c, result)
	}
}

// DailyDetailHandler 每日明细处理器
func DailyDetailHandler(store repository.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userIDVal, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "user not authenticated")
			return
		}
		userID := userIDVal.(string)

		date := c.Query("date")
		if date == "" {
			response.BadRequest(c, "date parameter is required")
			return
		}

		modeStr := c.Query("mode")
		if modeStr == "" {
			modeStr = string(TimelineModeHistory) // 默认历史模式
		}
		mode := TimelineMode(modeStr)
		if mode != TimelineModeHistory && mode != TimelineModeWithFuture {
			response.BadRequest(c, "invalid mode parameter")
			return
		}

		logger.Debug("Building daily detail",
			zap.String("userID", userID),
			zap.String("date", date),
			zap.String("mode", string(mode)),
		)

		result, err := BuildDailyDetail(c.Request.Context(), store, userID, date, mode)
		if err != nil {
			logger.Error("Failed to build daily detail",
				zap.String("userID", userID),
				zap.Error(err),
			)
			response.InternalError(c, "Failed to build daily detail: "+err.Error())
			return
		}

		logger.Info("Daily detail built successfully",
			zap.String("userID", userID),
			zap.Int("transactionsCount", len(result.Transactions)),
		)
		response.OK(c, result)
	}
}
