package user

import (
	"neverbroke/common/logger"
	"neverbroke/internal/auth"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 用户处理器
type Handler struct {
	userRepo auth.Repository
}

// NewHandler 创建用户处理器
func NewHandler(userRepo auth.Repository) *Handler {
	return &Handler{
		userRepo: userRepo,
	}
}

// GetSettings 获取用户设置
func (h *Handler) GetSettings(c *gin.Context) {
	// 从认证中间件获取当前用户
	userVal, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, ok := userVal.(*auth.User)
	if !ok {
		response.InternalError(c, "Invalid user context")
		return
	}

	logger.Debug("Fetching user settings", zap.Int("userID", user.ID))

	// 直接从用户对象返回设置
	settings := &auth.UserSettings{
		Locale:     user.Locale,
		ShowAmount: user.ShowAmount,
	}

	logger.Info("User settings fetched successfully", zap.Int("userID", user.ID))
	response.OK(c, settings)
}

// UpdateSettings 更新用户设置
func (h *Handler) UpdateSettings(c *gin.Context) {
	// 从认证中间件获取当前用户
	userVal, exists := c.Get("user")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	user, ok := userVal.(*auth.User)
	if !ok {
		response.InternalError(c, "Invalid user context")
		return
	}

	logger.Debug("Updating user settings", zap.Int("userID", user.ID))

	// 解析请求体（部分更新）
	var updates struct {
		Locale     *string `json:"locale"`
		ShowAmount *bool   `json:"showAmount"`
	}

	if err := c.ShouldBindJSON(&updates); err != nil {
		logger.Warn("Invalid request body for settings update",
			zap.Int("userID", user.ID),
			zap.Error(err),
		)
		response.BadRequest(c, "Invalid request body")
		return
	}

	// 构建新的设置（基于当前用户设置）
	newSettings := &auth.UserSettings{
		Locale:     user.Locale,
		ShowAmount: user.ShowAmount,
	}

	// 应用更新
	if updates.Locale != nil {
		newSettings.Locale = *updates.Locale
		logger.Debug("Locale updated", zap.String("locale", *updates.Locale))
	}
	if updates.ShowAmount != nil {
		newSettings.ShowAmount = *updates.ShowAmount
		logger.Debug("ShowAmount updated", zap.Bool("showAmount", *updates.ShowAmount))
	}

	// 保存设置到数据库
	if err := h.userRepo.UpdateUserSettings(user.ID, newSettings); err != nil {
		logger.Error("Failed to save user settings",
			zap.Int("userID", user.ID),
			zap.Error(err),
		)
		response.InternalError(c, "Failed to update user settings")
		return
	}

	logger.Info("User settings updated successfully",
		zap.Int("userID", user.ID),
		zap.String("locale", newSettings.Locale),
		zap.Bool("showAmount", newSettings.ShowAmount),
	)
	response.OK(c, newSettings)
}
