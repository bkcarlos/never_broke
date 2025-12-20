package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// RefreshToken Refresh Token 模型
type RefreshToken struct {
	ID         string     `gorm:"primaryKey" json:"id"`
	UserID     int        `gorm:"not null;index" json:"user_id"`
	DeviceID   string     `gorm:"not null;index" json:"device_id"`
	TokenHash  string     `gorm:"not null;index" json:"-"` // 永不序列化
	IssuedAt   time.Time  `gorm:"not null" json:"issued_at"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	ReplacedBy *string    `json:"replaced_by,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
}

// TableName 指定表名
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}

// OAuthIdentity 第三方登录身份
type OAuthIdentity struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	UserID    int       `gorm:"not null;index" json:"user_id"`
	Provider  string    `gorm:"not null;size:50" json:"provider"`
	Subject   string    `gorm:"not null" json:"subject"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName 指定表名
func (OAuthIdentity) TableName() string {
	return "oauth_identities"
}

// OAuthState OAuth 流程 state
type OAuthState struct {
	State       string    `gorm:"primaryKey" json:"state"`
	Provider    string    `gorm:"not null;size:50" json:"provider"`
	RedirectURI string    `json:"redirect_uri,omitempty"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	ExpiresAt   time.Time `gorm:"not null;index" json:"expires_at"`
}

// TableName 指定表名
func (OAuthState) TableName() string {
	return "oauth_states"
}

// AccessTokenClaims Access Token JWT Claims
type AccessTokenClaims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	DeviceID string `json:"device_id"` // 设备ID，用于验证token是否已被撤销
	jwt.RegisteredClaims
}

// RefreshTokenClaims Refresh Token JWT Claims
type RefreshTokenClaims struct {
	UserID   int    `json:"user_id"`
	DeviceID string `json:"device_id"`
	jwt.RegisteredClaims
}

// TokenPair Token 对（Access + Refresh）
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"` // Web 端不返回（用 Cookie）
	ExpiresIn    int64  `json:"expires_in"`              // Access token 过期时间（秒）
	TokenType    string `json:"token_type"`              // "Bearer"
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Nickname string `json:"nickname" binding:"required,min=1,max=100"`
	DeviceID string `json:"device_id" binding:"required"` // 客户端设备 ID
}

// LoginRequest 登录请求
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	DeviceID string `json:"device_id" binding:"required"` // 客户端设备 ID
}

// RefreshRequest 刷新 token 请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"` // App/小程序从 body 传
	DeviceID     string `json:"device_id"`     // 可选，用于验证
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	DeviceID string `json:"device_id"` // 可选，用于指定设备
}

// OAuthStartResponse OAuth 开始响应
type OAuthStartResponse struct {
	AuthURL string `json:"auth_url"` // 授权 URL（扫码或跳转）
	State   string `json:"state"`    // State（用于回调验证）
}

// OAuthCallbackRequest OAuth 回调请求
type OAuthCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}
