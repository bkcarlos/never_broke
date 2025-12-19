package auth

import (
	"net/http"
	"strings"
	"time"

	"neverbroke/common/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// TokenHandler Token 处理器
type TokenHandler struct {
	service     *TokenService
	rateLimiter *RateLimiter
}

// NewTokenHandler 创建 Token 处理器
func NewTokenHandler(service *TokenService) *TokenHandler {
	// 限流配置：5 分钟内最多 10 次尝试
	rateLimiter := NewRateLimiter(10, 5*time.Minute)

	return &TokenHandler{
		service:     service,
		rateLimiter: rateLimiter,
	}
}

// Register 注册接口
// POST /api/auth/register
// 安全说明：
// 1. 密码使用 argon2id 哈希
// 2. 注册成功后自动登录
// 3. Web 端：refresh token 用 HttpOnly Cookie
// 4. App/小程序：refresh token 在 body 返回
func (h *TokenHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	// 限流检查
	rateLimitKey := h.getRateLimitKey(c, req.Email)
	if !h.rateLimiter.Check(rateLimitKey) {
		logger.Warn("Rate limit exceeded for register",
			zap.String("ip", c.ClientIP()),
			zap.String("email", req.Email),
		)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many attempts, please try again later",
		})
		return
	}

	// 注册用户
	tokenPair, user, err := h.service.Register(&req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logger.Warn("Failed to register user",
			zap.Error(err),
			zap.String("email", req.Email),
		)

		// 记录失败尝试
		h.rateLimiter.RecordFailure(rateLimitKey)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 注册成功，重置限流
	h.rateLimiter.Reset(rateLimitKey)

	logger.Info("User registered successfully",
		zap.Int("user_id", user.ID),
		zap.String("email", user.Email),
	)

	// 判断客户端类型
	if h.isWebClient(c) {
		// Web 端：refresh token 用 HttpOnly Cookie
		h.setRefreshTokenCookie(c, tokenPair.RefreshToken)
		c.JSON(http.StatusCreated, gin.H{
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
		c.JSON(http.StatusCreated, gin.H{
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

// Login 登录接口
// POST /api/auth/login
// 安全说明：
// 1. 统一错误消息（防止邮箱枚举）
// 2. 限流保护（防止暴力破解）
// 3. Web 端：refresh token 用 HttpOnly Cookie
// 4. App/小程序：refresh token 在 body 返回
func (h *TokenHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	// 限流检查
	rateLimitKey := h.getRateLimitKey(c, req.Email)
	if !h.rateLimiter.Check(rateLimitKey) {
		logger.Warn("Rate limit exceeded for login",
			zap.String("ip", c.ClientIP()),
			zap.String("email", req.Email),
		)
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many attempts, please try again later",
		})
		return
	}

	// 登录
	tokenPair, user, err := h.service.Login(&req, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logger.Warn("Failed to login",
			zap.Error(err),
			zap.String("email", req.Email),
		)

		// 记录失败尝试
		h.rateLimiter.RecordFailure(rateLimitKey)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	// 登录成功，重置限流
	h.rateLimiter.Reset(rateLimitKey)

	logger.Info("User logged in successfully",
		zap.Int("user_id", user.ID),
		zap.String("email", user.Email),
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

// Refresh 刷新 Access Token
// POST /api/auth/refresh
// 安全说明：
// 1. Web 端：从 HttpOnly Cookie 获取 refresh token
// 2. App/小程序：从 body 获取 refresh token
// 3. Token Rotation：每次刷新生成新 refresh token，旧 token 作废
// 4. 重放检测：使用已撤销的 token 会触发安全策略
func (h *TokenHandler) Refresh(c *gin.Context) {
	var refreshToken string

	// 判断客户端类型
	if h.isWebClient(c) {
		// Web 端：从 Cookie 获取
		var err error
		refreshToken, err = c.Cookie("refresh_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "refresh token not found",
			})
			return
		}
	} else {
		// App/小程序：从 body 获取
		var req RefreshRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request",
			})
			return
		}
		refreshToken = req.RefreshToken
	}

	// 刷新 token
	tokenPair, err := h.service.RefreshAccessToken(refreshToken, c.Request.UserAgent(), c.ClientIP())
	if err != nil {
		logger.Warn("Failed to refresh token",
			zap.Error(err),
			zap.String("ip", c.ClientIP()),
		)

		// 清除 Cookie（如果是 Web 端）
		if h.isWebClient(c) {
			h.clearRefreshTokenCookie(c)
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	logger.Info("Token refreshed successfully")

	// 判断客户端类型
	if h.isWebClient(c) {
		// Web 端：新 refresh token 用 HttpOnly Cookie
		h.setRefreshTokenCookie(c, tokenPair.RefreshToken)
		c.JSON(http.StatusOK, gin.H{
			"access_token": tokenPair.AccessToken,
			"expires_in":   tokenPair.ExpiresIn,
			"token_type":   tokenPair.TokenType,
		})
	} else {
		// App/小程序：新 refresh token 在 body 返回
		c.JSON(http.StatusOK, gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"expires_in":    tokenPair.ExpiresIn,
			"token_type":    tokenPair.TokenType,
		})
	}
}

// Logout 登出（撤销当前设备）
// POST /api/auth/logout
func (h *TokenHandler) Logout(c *gin.Context) {
	// 从上下文获取用户（由 AuthMiddleware 设置）
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	u := user.(*User)

	// 获取 device_id
	var req LogoutRequest
	c.ShouldBindJSON(&req)

	deviceID := req.DeviceID
	if deviceID == "" {
		// 如果没有指定设备 ID，默认撤销所有设备
		deviceID = ""
	}

	// 撤销 token
	var err error
	if deviceID != "" {
		err = h.service.Logout(u.ID, deviceID)
	} else {
		// 如果没有指定设备，撤销所有设备
		err = h.service.LogoutAll(u.ID)
	}

	if err != nil {
		logger.Error("Failed to logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// 清除 Cookie（如果是 Web 端）
	if h.isWebClient(c) {
		h.clearRefreshTokenCookie(c)
	}

	logger.Info("User logged out successfully", zap.Int("user_id", u.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "logout successful",
	})
}

// LogoutAll 全部登出（撤销所有设备）
// POST /api/auth/logout_all
func (h *TokenHandler) LogoutAll(c *gin.Context) {
	// 从上下文获取用户
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	u := user.(*User)

	// 撤销所有设备的 token
	if err := h.service.LogoutAll(u.ID); err != nil {
		logger.Error("Failed to logout all", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	// 清除 Cookie（如果是 Web 端）
	if h.isWebClient(c) {
		h.clearRefreshTokenCookie(c)
	}

	logger.Info("User logged out from all devices", zap.Int("user_id", u.ID))

	c.JSON(http.StatusOK, gin.H{
		"message": "logged out from all devices",
	})
}

// Me 获取当前用户信息
// GET /api/me
func (h *TokenHandler) Me(c *gin.Context) {
	// 从上下文获取用户（由 AuthMiddleware 设置）
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "not authenticated",
		})
		return
	}

	u := user.(*User)
	c.JSON(http.StatusOK, UserResponse{
		Email:    u.Email,
		Nickname: u.Nickname,
	})
}

// AuthMiddleware 鉴权中间件
// 安全说明：
// 1. 从 Authorization header 获取 access token
// 2. 验证 token 有效性
// 3. 将用户信息存入上下文
func (h *TokenHandler) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format",
			})
			c.Abort()
			return
		}

		accessToken := parts[1]

		// 验证 token
		user, err := h.service.ValidateAccessToken(accessToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			c.Abort()
			return
		}

		// 将用户存入上下文
		c.Set("user", user)
		c.Next()
	}
}

// setRefreshTokenCookie 设置 Refresh Token Cookie
// 安全说明：
// 1. HttpOnly=true：防止 XSS 攻击
// 2. Secure=true：仅 HTTPS 传输（生产环境必须）
// 3. SameSite=Lax：防止 CSRF 攻击
// 4. Path=/api/auth/refresh：缩小作用域
func (h *TokenHandler) setRefreshTokenCookie(c *gin.Context, refreshToken string) {
	c.SetCookie(
		"refresh_token",                     // name
		refreshToken,                        // value
		int(RefreshTokenDuration.Seconds()), // maxAge
		"/api/auth/refresh",                 // path（缩小作用域）
		"",                                  // domain
		false,                               // secure（开发环境 false，生产环境必须 true）
		true,                                // httpOnly
	)

	// 设置 SameSite 属性
	c.Header("Set-Cookie", c.Writer.Header().Get("Set-Cookie")+"; SameSite=Lax")
}

// clearRefreshTokenCookie 清除 Refresh Token Cookie
func (h *TokenHandler) clearRefreshTokenCookie(c *gin.Context) {
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/api/auth/refresh",
		"",
		false,
		true,
	)
}

// isWebClient 判断是否是 Web 客户端
// 简单实现：检查 User-Agent 或自定义 header
func (h *TokenHandler) isWebClient(c *gin.Context) bool {
	// 方案1：通过自定义 header 判断
	clientType := c.GetHeader("X-Client-Type")
	if clientType == "app" || clientType == "miniapp" {
		return false
	}

	// 方案2：通过 User-Agent 判断（简化版）
	userAgent := c.Request.UserAgent()
	if strings.Contains(userAgent, "MicroMessenger") { // 微信小程序
		return false
	}
	if strings.Contains(userAgent, "Mobile") && !strings.Contains(userAgent, "Mozilla") {
		return false
	}

	// 默认认为是 Web
	return true
}

// getRateLimitKey 生成限流 key
func (h *TokenHandler) getRateLimitKey(c *gin.Context, email string) string {
	ip := c.ClientIP()
	email = strings.ToLower(strings.TrimSpace(email))
	return ip + ":" + email
}
