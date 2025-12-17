package auth

import (
	"errors"
	"fmt"
	"time"
)

// OAuthService OAuth 服务
type OAuthService struct {
	userRepo  Repository
	tokenRepo TokenRepository
	tokenMgr  *TokenManager
}

// NewOAuthService 创建 OAuth 服务
func NewOAuthService(userRepo Repository, tokenRepo TokenRepository, tokenMgr *TokenManager) *OAuthService {
	return &OAuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		tokenMgr:  tokenMgr,
	}
}

// StartOAuth 开始 OAuth 流程
// 生成 state 并返回授权 URL
func (s *OAuthService) StartOAuth(provider, redirectURI string) (*OAuthStartResponse, error) {
	// 1. 生成安全随机 state
	state, err := GenerateSecureToken(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}

	// 2. 保存 state（5 分钟过期）
	oauthState := &OAuthState{
		State:       state,
		Provider:    provider,
		RedirectURI: redirectURI,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(5 * time.Minute),
	}

	if err := s.tokenRepo.CreateOAuthState(oauthState); err != nil {
		return nil, fmt.Errorf("failed to save state: %w", err)
	}

	// 3. 获取提供商并生成授权 URL
	oauthProvider, err := GetProvider(provider)
	if err != nil {
		return nil, err
	}

	authURL := oauthProvider.GetAuthURL(state)

	return &OAuthStartResponse{
		AuthURL: authURL,
		State:   state,
	}, nil
}

// HandleOAuthCallback 处理 OAuth 回调
// 验证 state，交换 code，创建或登录用户
func (s *OAuthService) HandleOAuthCallback(provider, code, state string, deviceID, userAgent, ipAddress string) (*TokenPair, *User, error) {
	// 1. 验证 state
	oauthState, err := s.tokenRepo.GetOAuthState(state)
	if err != nil {
		if errors.Is(err, ErrOAuthStateExpired) {
			return nil, nil, errors.New("state expired")
		}
		return nil, nil, errors.New("invalid state")
	}

	// 验证 provider 匹配
	if oauthState.Provider != provider {
		return nil, nil, errors.New("provider mismatch")
	}

	// 删除 state（一次性使用）
	s.tokenRepo.DeleteOAuthState(state)

	// 2. 获取提供商并交换 code
	oauthProvider, err := GetProvider(provider)
	if err != nil {
		return nil, nil, err
	}

	userInfo, err := oauthProvider.ExchangeCode(code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 3. 查找或创建用户
	user, err := s.findOrCreateUserByOAuth(userInfo)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find or create user: %w", err)
	}

	// 4. 签发 token
	tokenPair, err := s.issueTokenPair(user, deviceID, userAgent, ipAddress)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to issue tokens: %w", err)
	}

	return tokenPair, user, nil
}

// findOrCreateUserByOAuth 根据 OAuth 信息查找或创建用户
func (s *OAuthService) findOrCreateUserByOAuth(userInfo *OAuthUserInfo) (*User, error) {
	// 1. 查找 OAuth 身份
	identity, err := s.tokenRepo.GetOAuthIdentity(userInfo.Provider, userInfo.Subject)
	if err == nil {
		// 身份已存在，获取用户
		user, err := s.userRepo.GetUserByID(identity.UserID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user: %w", err)
		}
		return user, nil
	}

	if !errors.Is(err, ErrOAuthIdentityNotFound) {
		return nil, fmt.Errorf("failed to get oauth identity: %w", err)
	}

	// 2. 身份不存在，创建新用户
	// 生成默认昵称（如果 OAuth 没有提供）
	nickname := userInfo.Nickname
	if nickname == "" {
		nickname = fmt.Sprintf("%s_user_%s", userInfo.Provider, userInfo.Subject[:8])
	}

	user := &User{
		Email:        "", // OAuth 用户可以没有邮箱
		PasswordHash: "", // OAuth 用户没有密码
		Nickname:     nickname,
	}

	// 注意：这里需要修改 CreateUser 方法，允许 email 为空
	// 或者生成一个虚拟邮箱，例如：{provider}_{subject}@oauth.local
	user.Email = fmt.Sprintf("%s_%s@oauth.local", userInfo.Provider, userInfo.Subject)

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// 3. 创建 OAuth 身份映射
	identity = &OAuthIdentity{
		UserID:   user.ID,
		Provider: userInfo.Provider,
		Subject:  userInfo.Subject,
	}

	if err := s.tokenRepo.CreateOAuthIdentity(identity); err != nil {
		return nil, fmt.Errorf("failed to create oauth identity: %w", err)
	}

	return user, nil
}

// issueTokenPair 签发 Token 对（复用 TokenService 的逻辑）
func (s *OAuthService) issueTokenPair(user *User, deviceID, userAgent, ipAddress string) (*TokenPair, error) {
	// 1. 生成 Access Token
	accessToken, expiresIn, err := s.tokenMgr.GenerateAccessToken(user)
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
