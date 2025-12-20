package auth

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrRefreshTokenNotFound  = errors.New("refresh token not found")
	ErrRefreshTokenExpired   = errors.New("refresh token expired")
	ErrRefreshTokenRevoked   = errors.New("refresh token revoked")
	ErrRefreshTokenReplaced  = errors.New("refresh token has been replaced")
	ErrOAuthIdentityNotFound = errors.New("oauth identity not found")
	ErrOAuthIdentityExists   = errors.New("oauth identity already exists")
	ErrOAuthStateNotFound    = errors.New("oauth state not found")
	ErrOAuthStateExpired     = errors.New("oauth state expired")
)

// TokenRepository Token 仓储接口
type TokenRepository interface {
	// RefreshToken operations
	CreateRefreshToken(token *RefreshToken) error
	GetRefreshToken(jti string) (*RefreshToken, error)
	GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error)
	UpdateRefreshToken(token *RefreshToken) error
	RevokeRefreshToken(jti string, replacedBy *string) error
	RevokeAllUserTokens(userID int) error
	RevokeDeviceTokens(userID int, deviceID string) error
	DeleteExpiredRefreshTokens() error

	// OAuthIdentity operations
	CreateOAuthIdentity(identity *OAuthIdentity) error
	GetOAuthIdentity(provider, subject string) (*OAuthIdentity, error)
	GetOAuthIdentitiesByUser(userID int) ([]OAuthIdentity, error)

	// OAuthState operations
	CreateOAuthState(state *OAuthState) error
	SaveOAuthState(state *OAuthState) error
	GetOAuthState(state string) (*OAuthState, error)
	DeleteOAuthState(state string) error
	DeleteExpiredOAuthStates() error
}

// GormTokenRepository GORM 实现的 Token 仓储
type GormTokenRepository struct {
	db *gorm.DB
}

// NewGormTokenRepository 创建 GORM Token 仓储
func NewGormTokenRepository(db *gorm.DB) TokenRepository {
	return &GormTokenRepository{db: db}
}

// CreateRefreshToken 创建 Refresh Token
func (r *GormTokenRepository) CreateRefreshToken(token *RefreshToken) error {
	return r.db.Create(token).Error
}

// GetRefreshToken 根据 jti 获取 Refresh Token
func (r *GormTokenRepository) GetRefreshToken(jti string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.Where("id = ?", jti).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	// 检查是否已撤销
	if token.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}

	// 检查是否已被替换
	if token.ReplacedBy != nil {
		return nil, ErrRefreshTokenReplaced
	}

	return &token, nil
}

// GetRefreshTokenByHash 根据 token hash 获取 Refresh Token
func (r *GormTokenRepository) GetRefreshTokenByHash(tokenHash string) (*RefreshToken, error) {
	var token RefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefreshTokenNotFound
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(token.ExpiresAt) {
		return nil, ErrRefreshTokenExpired
	}

	// 检查是否已撤销
	if token.RevokedAt != nil {
		return nil, ErrRefreshTokenRevoked
	}

	// 检查是否已被替换
	if token.ReplacedBy != nil {
		return nil, ErrRefreshTokenReplaced
	}

	return &token, nil
}

// UpdateRefreshToken 更新 Refresh Token
func (r *GormTokenRepository) UpdateRefreshToken(token *RefreshToken) error {
	return r.db.Save(token).Error
}

// RevokeRefreshToken 撤销 Refresh Token
func (r *GormTokenRepository) RevokeRefreshToken(jti string, replacedBy *string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"revoked_at": now,
	}
	if replacedBy != nil {
		updates["replaced_by"] = *replacedBy
	}

	result := r.db.Model(&RefreshToken{}).
		Where("id = ?", jti).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRefreshTokenNotFound
	}
	return nil
}

// RevokeAllUserTokens 撤销用户所有 Refresh Token
func (r *GormTokenRepository) RevokeAllUserTokens(userID int) error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// RevokeDeviceTokens 撤销指定设备的 Refresh Token
func (r *GormTokenRepository) RevokeDeviceTokens(userID int, deviceID string) error {
	now := time.Now()
	return r.db.Model(&RefreshToken{}).
		Where("user_id = ? AND device_id = ? AND revoked_at IS NULL", userID, deviceID).
		Update("revoked_at", now).Error
}

// DeleteExpiredRefreshTokens 删除过期的 Refresh Token
func (r *GormTokenRepository) DeleteExpiredRefreshTokens() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&RefreshToken{}).Error
}

// CreateOAuthIdentity 创建 OAuth 身份
func (r *GormTokenRepository) CreateOAuthIdentity(identity *OAuthIdentity) error {
	return r.db.Create(identity).Error
}

// GetOAuthIdentity 获取 OAuth 身份
func (r *GormTokenRepository) GetOAuthIdentity(provider, subject string) (*OAuthIdentity, error) {
	var identity OAuthIdentity
	err := r.db.Where("provider = ? AND subject = ?", provider, subject).First(&identity).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOAuthIdentityNotFound
		}
		return nil, err
	}
	return &identity, nil
}

// GetOAuthIdentitiesByUser 获取用户的所有 OAuth 身份
func (r *GormTokenRepository) GetOAuthIdentitiesByUser(userID int) ([]OAuthIdentity, error) {
	var identities []OAuthIdentity
	err := r.db.Where("user_id = ?", userID).Find(&identities).Error
	return identities, err
}

// CreateOAuthState 创建 OAuth State
func (r *GormTokenRepository) CreateOAuthState(state *OAuthState) error {
	return r.db.Create(state).Error
}

// SaveOAuthState 保存 OAuth State（别名，用于兼容）
func (r *GormTokenRepository) SaveOAuthState(state *OAuthState) error {
	return r.CreateOAuthState(state)
}

// GetOAuthState 获取 OAuth State
func (r *GormTokenRepository) GetOAuthState(state string) (*OAuthState, error) {
	var oauthState OAuthState
	err := r.db.Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOAuthStateNotFound
		}
		return nil, err
	}

	// 检查是否过期
	if time.Now().After(oauthState.ExpiresAt) {
		return nil, ErrOAuthStateExpired
	}

	return &oauthState, nil
}

// DeleteOAuthState 删除 OAuth State
func (r *GormTokenRepository) DeleteOAuthState(state string) error {
	return r.db.Where("state = ?", state).Delete(&OAuthState{}).Error
}

// DeleteExpiredOAuthStates 删除过期的 OAuth State
func (r *GormTokenRepository) DeleteExpiredOAuthStates() error {
	return r.db.Where("expires_at < ?", time.Now()).Delete(&OAuthState{}).Error
}
