package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// TokenService Token 服务
type TokenService struct {
	userRepo  Repository      // 用户仓储
	tokenRepo TokenRepository // Token 仓储
	tokenMgr  *TokenManager   // Token 管理器
}

// NewTokenService 创建 Token 服务
func NewTokenService(userRepo Repository, tokenRepo TokenRepository, tokenMgr *TokenManager) *TokenService {
	return &TokenService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokenMgr:  tokenMgr,
	}
}

// Register 用户注册
// 安全说明：
// 1. 邮箱转小写存储
// 2. 密码使用 argon2id 哈希
// 3. 注册成功后自动登录（签发 token）
func (s *TokenService) Register(req *RegisterRequest, userAgent, ipAddress string) (*TokenPair, *User, error) {
	// 1. 验证密码强度
	if len(req.Password) < 8 {
		return nil, nil, errors.New("password must be at least 8 characters")
	}

	// 2. 哈希密码
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 3. 创建用户
	user := &User{
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: passwordHash,
		Nickname:     strings.TrimSpace(req.Nickname),
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		if errors.Is(err, ErrEmailExists) {
			return nil, nil, errors.New("email already exists")
		}
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 4. 签发 token（自动登录）
	tokenPair, err := s.issueTokenPair(user, req.DeviceID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue tokens: %w", err)
	}

	return tokenPair, user, nil
}

// Login 用户登录
// 安全说明：
// 1. 统一错误消息（防止邮箱枚举）
// 2. 验证密码（argon2id）
// 3. 签发 token
func (s *TokenService) Login(req *LoginRequest, userAgent, ipAddress string) (*TokenPair, *User, error) {
	// 1. 查询用户
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		// 统一错误消息（防止邮箱枚举）
		return nil, nil, errors.New("invalid credentials")
	}

	// 2. 验证密码
	valid, err := VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !valid {
		// 统一错误消息（防止邮箱枚举）
		return nil, nil, errors.New("invalid credentials")
	}

	// 3. 签发 token
	tokenPair, err := s.issueTokenPair(user, req.DeviceID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue tokens: %w", err)
	}

	return tokenPair, user, nil
}

// RefreshAccessToken 刷新 Access Token
// 安全说明：
// 1. 验证 Refresh Token
// 2. 实现 Token Rotation（生成新 refresh，旧 refresh 作废）
// 3. 重放检测（如果使用已撤销的 token，触发安全策略）
func (s *TokenService) RefreshAccessToken(refreshToken string, userAgent, ipAddress string) (*TokenPair, error) {
	// 1. 验证 Refresh Token JWT
	claims, err := s.tokenMgr.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// 2. 计算 token hash
	tokenHash := HashRefreshToken(refreshToken)

	// 3. 从数据库获取 refresh token 记录
	tokenRecord, err := s.tokenRepo.GetRefreshTokenByHash(tokenHash)
	if err != nil {
		// 重放检测：如果 token 已被替换，说明可能是重放攻击
		if errors.Is(err, ErrRefreshTokenReplaced) {
			// 安全策略：撤销该用户该设备的所有 token
			s.tokenRepo.RevokeDeviceTokens(claims.UserID, claims.DeviceID)
			return nil, errors.New("token reuse detected, all device tokens revoked")
		}
		return nil, errors.New("invalid refresh token")
	}

	// 4. 并发安全：先撤销旧token（原子操作）
	// 生成新token的JTI（用于replaced_by字段）
	newJti := uuid.New().String()

	// 尝试撤销旧token（并发安全：只有一个请求能成功）
	err = s.tokenRepo.RevokeRefreshToken(tokenRecord.ID, &newJti)
	if err != nil {
		// Token已被其他请求使用（并发冲突）
		return nil, errors.New("token has been used, please try again")
	}

	// 5. 获取用户信息
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// 6. 生成新的 Access Token
	accessToken, expiresIn, err := s.tokenMgr.GenerateAccessToken(user, claims.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 7. Token Rotation：生成新的 Refresh Token
	newRefreshToken, _, err := s.tokenMgr.GenerateRefreshTokenWithJTI(user.ID, claims.DeviceID, newJti)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 8. 保存新的 Refresh Token
	newTokenHash := HashRefreshToken(newRefreshToken)
	newTokenRecord := &RefreshToken{
		ID:        newJti,
		UserID:    user.ID,
		DeviceID:  claims.DeviceID,
		TokenHash: newTokenHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.tokenRepo.CreateRefreshToken(newTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// Logout 登出（撤销当前设备的 Refresh Token）
func (s *TokenService) Logout(userID int, deviceID string) error {
	return s.tokenRepo.RevokeDeviceTokens(userID, deviceID)
}

// LogoutAll 全部登出（撤销用户所有设备的 Refresh Token）
func (s *TokenService) LogoutAll(userID int) error {
	return s.tokenRepo.RevokeAllUserTokens(userID)
}

// ValidateAccessToken 验证 Access Token
func (s *TokenService) ValidateAccessToken(token string) (*User, error) {
	// 1. 验证 JWT
	claims, err := s.tokenMgr.ValidateAccessToken(token)
	if err != nil {
		return nil, err
	}

	// 2. 检查该设备是否还有活跃的 refresh token
	// 如果用户已经logout该设备，refresh token会被撤销，access token也应该失效
	if claims.DeviceID != "" {
		activeTokenCount, err := s.tokenRepo.CountActiveTokensForDevice(claims.UserID, claims.DeviceID)
		if err != nil {
			return nil, fmt.Errorf("failed to check token status: %w", err)
		}

		if activeTokenCount == 0 {
			return nil, errors.New("device has been logged out")
		}
	}

	// 3. 获取用户（确保用户仍然存在）
	user, err := s.userRepo.GetUserByID(claims.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// issueTokenPair 签发 Token 对（Access + Refresh）
func (s *TokenService) issueTokenPair(user *User, deviceID, userAgent, ipAddress string) (*TokenPair, error) {
	// 1. 生成 Access Token
	accessToken, expiresIn, err := s.tokenMgr.GenerateAccessToken(user, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// 2. 生成 Refresh Token
	refreshToken, jti, err := s.tokenMgr.GenerateRefreshToken(user.ID, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 3. 保存 Refresh Token（哈希存储）
	tokenHash := HashRefreshToken(refreshToken)
	tokenRecord := &RefreshToken{
		ID:        jti,
		UserID:    user.ID,
		DeviceID:  deviceID,
		TokenHash: tokenHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(RefreshTokenDuration),
		UserAgent: userAgent,
		IPAddress: ipAddress,
	}

	if err := s.tokenRepo.CreateRefreshToken(tokenRecord); err != nil {
		return nil, fmt.Errorf("failed to save refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		TokenType:    "Bearer",
	}, nil
}

// CleanupExpiredData 清理过期数据（定期任务）
func (s *TokenService) CleanupExpiredData() error {
	if err := s.tokenRepo.DeleteExpiredRefreshTokens(); err != nil {
		return fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}
	if err := s.tokenRepo.DeleteExpiredOAuthStates(); err != nil {
		return fmt.Errorf("failed to delete expired oauth states: %w", err)
	}
	return nil
}
