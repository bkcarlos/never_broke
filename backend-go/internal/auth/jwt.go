package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWT 配置
const (
	// Access Token 过期时间（10 分钟）
	AccessTokenDuration = 10 * time.Minute

	// Refresh Token 过期时间（7 天）
	RefreshTokenDuration = 7 * 24 * time.Hour

	// 默认 JWT 密钥（生产环境必须使用环境变量）
	defaultJWTSecret = "your-secret-key-change-this-in-production-min-32-bytes"
)

// TokenManager Token 管理器（Access + Refresh）
type TokenManager struct {
	secretKey []byte
}

// NewTokenManager 创建 Token 管理器
// 安全说明：
// 1. 优先从环境变量读取密钥
// 2. 生产环境必须设置 JWT_SECRET 环境变量
// 3. 密钥长度建议至少 32 字节
func NewTokenManager() *TokenManager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 开发环境使用默认密钥（生产环境必须设置环境变量）
		secret = defaultJWTSecret
	}

	return &TokenManager{
		secretKey: []byte(secret),
	}
}

// GenerateAccessToken 生成 Access Token
// 安全说明：
// 1. 使用 HS256 签名算法
// 2. 短期有效（10 分钟）
// 3. 包含用户基本信息（不包含敏感信息）
// 4. 包含设备ID，用于验证token是否已被撤销
func (tm *TokenManager) GenerateAccessToken(user *User, deviceID string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(AccessTokenDuration)

	claims := AccessTokenClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Nickname: user.Nickname,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "neverbroke-auth",
			Subject:   fmt.Sprintf("%d", user.ID),
			ID:        uuid.New().String(), // JTI: 确保每个token唯一
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		return "", 0, fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, int64(AccessTokenDuration.Seconds()), nil
}

// generateJTI 生成唯一的JTI
func generateJTI() string {
	return uuid.New().String()
}

// GenerateRefreshToken 生成 Refresh Token
// 安全说明：
// 1. 使用 UUID v4 作为 jti（唯一标识）
// 2. 长期有效（7 天）
// 3. 仅包含 user_id 和 device_id
// 4. 返回 token 和 jti（用于数据库存储）
func (tm *TokenManager) GenerateRefreshToken(userID int, deviceID string) (token string, jti string, err error) {
	jti = generateJTI()
	return tm.GenerateRefreshTokenWithJTI(userID, deviceID, jti)
}

// GenerateRefreshTokenWithJTI 使用指定的JTI生成 Refresh Token
func (tm *TokenManager) GenerateRefreshTokenWithJTI(userID int, deviceID string, jti string) (token string, _ string, err error) {
	now := time.Now()
	expiresAt := now.Add(RefreshTokenDuration)

	claims := RefreshTokenClaims{
		UserID:   userID,
		DeviceID: deviceID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "neverbroke-auth",
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString(tm.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, jti, nil
}

// ValidateAccessToken 验证 Access Token
func (tm *TokenManager) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	claims, ok := token.Claims.(*AccessTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	// 验证 issuer
	if claims.Issuer != "neverbroke-auth" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}

// ValidateRefreshToken 验证 Refresh Token
func (tm *TokenManager) ValidateRefreshToken(tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	// 验证 issuer
	if claims.Issuer != "neverbroke-auth" {
		return nil, errors.New("invalid token issuer")
	}

	return claims, nil
}

// HashRefreshToken 对 Refresh Token 进行 SHA-256 哈希
// 安全说明：数据库中禁止存储明文 refresh token
func HashRefreshToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GenerateSecureToken 生成安全随机 token（用于 OAuth state 等）
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
