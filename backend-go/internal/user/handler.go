package user

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/memory"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 用户处理器
type Handler struct {
	store *memory.Store
}

// NewHandler 创建用户处理器
func NewHandler(store *memory.Store) *Handler {
	return &Handler{
		store: store,
	}
}

// GetSettings 获取用户设置
func (h *Handler) GetSettings(c *gin.Context) {
	// 当前使用固定用户 ID
	userID := "demo"

	logger.Debug("Fetching user settings", zap.String("userID", userID))

	settings, err := h.store.GetUserSettings(userID)
	if err != nil {
		logger.Error("Failed to get user settings",
			zap.String("userID", userID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to get user settings")
		return
	}

	logger.Info("User settings fetched successfully", zap.String("userID", userID))
	response.OK(c, settings)
}

// UpdateSettings 更新用户设置
func (h *Handler) UpdateSettings(c *gin.Context) {
	// 当前使用固定用户 ID
	userID := "demo"

	logger.Debug("Updating user settings", zap.String("userID", userID))

	// 获取当前设置
	currentSettings, err := h.store.GetUserSettings(userID)
	if err != nil {
		logger.Error("Failed to get user settings for update",
			zap.String("userID", userID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to get user settings")
		return
	}

	// 解析请求体（部分更新）
	var updates struct {
		Locale     *string `json:"locale"`
		ShowAmount *bool   `json:"showAmount"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		logger.Warn("Invalid request body for settings update",
			zap.String("userID", userID),
			zap.Error(err),
		)
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 应用更新
	if updates.Locale != nil {
		currentSettings.Locale = *updates.Locale
		logger.Debug("Locale updated", zap.String("locale", *updates.Locale))
	}
	if updates.ShowAmount != nil {
		currentSettings.ShowAmount = *updates.ShowAmount
		logger.Debug("ShowAmount updated", zap.Bool("showAmount", *updates.ShowAmount))
	}

	// 保存设置
	if err := h.store.UpdateUserSettings(userID, currentSettings); err != nil {
		logger.Error("Failed to save user settings",
			zap.String("userID", userID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to update user settings")
		return
	}

	logger.Info("User settings updated successfully",
		zap.String("userID", userID),
		zap.String("locale", currentSettings.Locale),
		zap.Bool("showAmount", currentSettings.ShowAmount),
	)
	response.OK(c, currentSettings)
}

