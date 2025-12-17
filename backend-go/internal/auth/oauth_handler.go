package auth

import (
	"net/http"

	"neverbroke/common/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OAuthHandler OAuth 处理器
type OAuthHandler struct {
	service *OAuthService
}

// NewOAuthHandler 创建 OAuth 处理器
func NewOAuthHandler(service *OAuthService) *OAuthHandler {
	return &OAuthHandler{
		service: service,
	}
}

// Start OAuth 开始流程
// GET /api/oauth/:provider/start
// 参数：
// - provider: wechat, google, apple, github
// - redirect_uri: 前端回调地址（可选）
func (h *OAuthHandler) Start(c *gin.Context) {
	provider := c.Param("provider")
	redirectURI := c.Query("redirect_uri")

	// 开始 OAuth 流程
	response, err := h.service.StartOAuth(provider, redirectURI)
	if err != nil {
		logger.Error("Failed to start OAuth",
			zap.Error(err),
			zap.String("provider", provider),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("OAuth flow started",
		zap.String("provider", provider),
		zap.String("state", response.State),
	)

	c.JSON(http.StatusOK, response)
}

// Callback OAuth 回调处理
// GET /api/oauth/:provider/callback
// 参数：
// - code: 授权码
// - state: 状态码
func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	state := c.Query("state")

	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "code and state are required",
		})
		return
	}

	// 生成 device_id（如果客户端没有提供）
	deviceID := c.Query("device_id")
	if deviceID == "" {
		// 可以从 Cookie 或 header 获取，或生成新的
		deviceID = "web_" + state[:8] // 简化处理
	}

	// 处理 OAuth 回调
	tokenPair, user, err := h.service.HandleOAuthCallback(
		provider,
		code,
		state,
		deviceID,
		c.Request.UserAgent(),
		c.ClientIP(),
	)
	if err != nil {
		logger.Error("Failed to handle OAuth callback",
			zap.Error(err),
			zap.String("provider", provider),
		)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("OAuth login successful",
		zap.String("provider", provider),
		zap.Int("user_id", user.ID),
	)

	// 判断客户端类型
	if h.isWebClient(c) {
		// Web 端：refresh token 用 HttpOnly Cookie
		h.setRefreshTokenCookie(c, tokenPair.RefreshToken)
		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenPair.AccessToken,
			"expires_in":   tokenPair.ExpiresIn,
			"token_type":   tokenPair.TokenType,
			"user": UserResponse{
				Email:    user.Email,
				Nickname: user.Nickname,
			},
		})
	} else {
		// App/小程序：refresh token 在 body 返回
		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    tokenPair.TokenType,
			"user": UserResponse{
				Email:    user.Email,
				Nickname: user.Nickname,
			},
		})
	}
}

// setRefreshTokenCookie 设置 Refresh Token Cookie
func (h *OAuthHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetCookie(
		"refresh_token",
		refreshToken,
		int(RefreshTokenDuration.Seconds()),
		"/api/auth/refresh",
		"",
		false, // 开发环境 false，生产环境必须 true
		true,
	)
	c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie")+"; SameSite=Lax")
}

// isWebClient 判断是否是 Web 客户端
func (h *OAuthHandler) isWebClient(c *gin.Context) bool {
	clientType := c.GetHeader("X-Client-Type")
	return clientType != "app" && clientType != "miniapp"
}
