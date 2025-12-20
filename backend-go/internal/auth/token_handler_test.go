package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockUserRepository 模拟用户仓储
type MockUserRepository struct {
	users  map[string]*User
	nextID int
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:  make(map[string]*User),
		nextID: 1,
	}
}

func (m *MockUserRepository) CreateUser(user *User) error {
	email := strings.ToLower(user.Email)
	if _, exists := m.users[email]; exists {
		return ErrEmailExists
	}
	user.ID = m.nextID
	m.nextID++
	m.users[email] = user
	return nil
}

func (m *MockUserRepository) GetUserByEmail(email string) (*User, error) {
	email = strings.ToLower(email)
	user, exists := m.users[email]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *MockUserRepository) GetUserByID(id int) (*User, error) {
	for _, user := range m.users {
		if user.ID == id {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

// MockTokenRepository 模拟 Token 仓储
type MockTokenRepository struct {
	refreshTokens map[string]*RefreshToken
	oauthStates   map[string]*OAuthState
}

func NewMockTokenRepository() *MockTokenRepository {
	return &MockTokenRepository{
		refreshTokens: make(map[string]*RefreshToken),
		oauthStates:   make(map[string]*OAuthState),
	}
}

func (m *MockTokenRepository) CreateRefreshToken(token *RefreshToken) error {
	m.refreshTokens[token.ID] = token
	return nil
}

func (m *MockTokenRepository) GetRefreshToken(jti string) (*RefreshToken, error) {
	if token, exists := m.refreshTokens[jti]; exists && token.RevokedAt == nil {
		return token, nil
	}
	return nil, ErrRefreshTokenNotFound
}

func (m *MockTokenRepository) GetRefreshTokenByHash(hash string) (*RefreshToken, error) {
	for _, token := range m.refreshTokens {
		if token.TokenHash == hash && token.RevokedAt == nil {
			return token, nil
		}
	}
	return nil, ErrRefreshTokenNotFound
}

func (m *MockTokenRepository) UpdateRefreshToken(token *RefreshToken) error {
	if existing, exists := m.refreshTokens[token.ID]; exists {
		*existing = *token
		return nil
	}
	return ErrRefreshTokenNotFound
}

func (m *MockTokenRepository) RevokeRefreshToken(id string, replacedBy *string) error {
	if token, exists := m.refreshTokens[id]; exists {
		token.RevokedAt = timePtr(time.Now())
		token.ReplacedBy = replacedBy
		return nil
	}
	return ErrRefreshTokenNotFound
}

func (m *MockTokenRepository) RevokeDeviceTokens(userID int, deviceID string) error {
	for _, token := range m.refreshTokens {
		if token.UserID == userID && token.DeviceID == deviceID && token.RevokedAt == nil {
			token.RevokedAt = timePtr(time.Now())
		}
	}
	return nil
}

func (m *MockTokenRepository) RevokeAllUserTokens(userID int) error {
	for _, token := range m.refreshTokens {
		if token.UserID == userID && token.RevokedAt == nil {
			token.RevokedAt = timePtr(time.Now())
		}
	}
	return nil
}

func (m *MockTokenRepository) DeleteExpiredRefreshTokens() error {
	now := time.Now()
	for id, token := range m.refreshTokens {
		if token.ExpiresAt.Before(now) {
			delete(m.refreshTokens, id)
		}
	}
	return nil
}

func (m *MockTokenRepository) CountActiveTokensForDevice(userID int, deviceID string) (int64, error) {
	var count int64
	for _, token := range m.refreshTokens {
		if token.UserID == userID && token.DeviceID == deviceID && token.RevokedAt == nil {
			count++
		}
	}
	return count, nil
}

func (m *MockTokenRepository) SaveOAuthState(state *OAuthState) error {
	m.oauthStates[state.State] = state
	return nil
}

func (m *MockTokenRepository) GetOAuthState(state string) (*OAuthState, error) {
	if s, exists := m.oauthStates[state]; exists {
		return s, nil
	}
	return nil, ErrOAuthStateNotFound
}

func (m *MockTokenRepository) DeleteOAuthState(state string) error {
	delete(m.oauthStates, state)
	return nil
}

func (m *MockTokenRepository) CreateOAuthState(state *OAuthState) error {
	m.oauthStates[state.State] = state
	return nil
}

func (m *MockTokenRepository) DeleteExpiredOAuthStates() error {
	now := time.Now()
	for state, s := range m.oauthStates {
		if s.ExpiresAt.Before(now) {
			delete(m.oauthStates, state)
		}
	}
	return nil
}

func (m *MockTokenRepository) CreateOAuthIdentity(identity *OAuthIdentity) error {
	// For simplicity, just return nil
	return nil
}

func (m *MockTokenRepository) GetOAuthIdentity(provider, subject string) (*OAuthIdentity, error) {
	return nil, ErrOAuthIdentityNotFound
}

func (m *MockTokenRepository) GetOAuthIdentitiesByUser(userID int) ([]OAuthIdentity, error) {
	return []OAuthIdentity{}, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}

// setupTestAuthRouter 设置测试路由
func setupTestAuthRouter() (*gin.Engine, *TokenService, *MockUserRepository, *MockTokenRepository) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// 创建 mock repositories
	userRepo := NewMockUserRepository()
	tokenRepo := NewMockTokenRepository()

	// 创建 token manager
	// 设置测试环境变量
	os.Setenv("JWT_SECRET", "test-secret-key-for-testing-only-min-32-chars")
	tokenMgr := NewTokenManager()

	// 创建服务
	service := NewTokenService(userRepo, tokenRepo, tokenMgr)

	// 创建处理器
	handler := NewTokenHandler(service)

	// 注册路由
	router.POST("/auth/register", handler.Register)
	router.POST("/auth/login", handler.Login)
	router.POST("/auth/refresh", handler.Refresh)
	router.POST("/auth/logout", handler.AuthMiddleware(), handler.Logout)
	router.POST("/auth/logout_all", handler.AuthMiddleware(), handler.LogoutAll)
	router.GET("/me", handler.AuthMiddleware(), handler.Me)

	return router, service, userRepo, tokenRepo
}

func TestRegister(t *testing.T) {
	router, _, userRepo, _ := setupTestAuthRouter()

	t.Run("successful registration as app client", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
			Nickname: "Test User",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Client-Type", "app") // 标记为 app 客户端

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "expires_in")
		assert.Contains(t, response, "token_type")
		assert.Contains(t, response, "user")

		// 验证用户已创建
		user, err := userRepo.GetUserByEmail("test@example.com")
		require.NoError(t, err)
		assert.Equal(t, "test@example.com", user.Email)
		assert.Equal(t, "Test User", user.Nickname)
	})

	t.Run("successful registration as web client", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "webtest@example.com",
			Password: "password123",
			Nickname: "Web User",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("User-Agent", "Mozilla/5.0") // Web 客户端

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.NotContains(t, response, "refresh_token") // Web 端不返回 refresh_token
		assert.Contains(t, response, "expires_in")
		assert.Contains(t, response, "token_type")

		// 验证 refresh_token 在 Cookie 中
		cookies := w.Result().Cookies()
		var refreshCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "refresh_token" {
				refreshCookie = cookie
				break
			}
		}
		assert.NotNil(t, refreshCookie, "refresh_token cookie should be set")
		assert.NotEmpty(t, refreshCookie.Value)
	})

	t.Run("duplicate email", func(t *testing.T) {
		// 先注册一次
		req := RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "password123",
			Nickname: "User 1",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w1 := httptest.NewRecorder()
		r1, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(body)))
		r1.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w1, r1)

		// 再次注册相同邮箱
		w2 := httptest.NewRecorder()
		r2, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(body)))
		r2.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w2, r2)

		assert.Equal(t, http.StatusBadRequest, w2.Code)

		var response map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "already exists")
	})

	t.Run("weak password", func(t *testing.T) {
		req := RegisterRequest{
			Email:    "weak@example.com",
			Password: "123",
			Nickname: "User",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/register", strings.NewReader("invalid json"))
		r.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestLogin(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	// 先注册一个用户
	regReq := RegisterRequest{
		Email:    "login@example.com",
		Password: "password123",
		Nickname: "Login User",
		DeviceID: "device-reg",
	}
	regBody, _ := json.Marshal(regReq)
	regW := httptest.NewRecorder()
	regR, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(regBody)))
	regR.Header.Set("Content-Type", "application/json")
	regR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(regW, regR)

	require.Equal(t, http.StatusCreated, regW.Code, "Registration should succeed")

	t.Run("successful login", func(t *testing.T) {
		req := LoginRequest{
			Email:    "login@example.com",
			Password: "password123",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Client-Type", "app") // 使用 app 客户端

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "expires_in")
		assert.Contains(t, response, "token_type")
	})

	t.Run("wrong password", func(t *testing.T) {
		req := LoginRequest{
			Email:    "login@example.com",
			Password: "wrongpassword",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid credentials", response["error"])
	})

	t.Run("non-existent user", func(t *testing.T) {
		req := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "invalid credentials", response["error"])
	})

	t.Run("invalid json", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader("invalid json"))
		r.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestRefresh(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	// 先注册一个用户
	regReq := RegisterRequest{
		Email:    "refresh@example.com",
		Password: "password123",
		Nickname: "Refresh User",
		DeviceID: "device-reg",
	}
	regBody, _ := json.Marshal(regReq)
	regW := httptest.NewRecorder()
	regR, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(regBody)))
	regR.Header.Set("Content-Type", "application/json")
	regR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(regW, regR)

	// 登录获取初始 token
	loginReq := LoginRequest{
		Email:    "refresh@example.com",
		Password: "password123",
		DeviceID: "device-1",
	}

	body, _ := json.Marshal(loginReq)
	w := httptest.NewRecorder()
	r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(body)))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Client-Type", "app") // 标记为 app，返回 refresh_token 在 body

	router.ServeHTTP(w, r)

	var loginResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResponse)

	if loginResponse["refresh_token"] == nil {
		t.Fatal("Failed to get refresh token from login response")
	}
	refreshToken := loginResponse["refresh_token"].(string)

	t.Run("successful refresh", func(t *testing.T) {
		req := RefreshRequest{
			RefreshToken: refreshToken,
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Client-Type", "app")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "expires_in")

		// 新的 refresh_token 应该与旧的不同（Token Rotation）
		newRefreshToken := response["refresh_token"].(string)
		assert.NotEqual(t, refreshToken, newRefreshToken)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		req := RefreshRequest{
			RefreshToken: "invalid-token",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Client-Type", "app")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing refresh token for app", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader("{}"))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("X-Client-Type", "app")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing refresh token cookie for web", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/refresh", nil)
		r.Header.Set("User-Agent", "Mozilla/5.0")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestAuthMiddleware(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	// 先注册一个用户并获取 token
	regReq := RegisterRequest{
		Email:    "middleware@example.com",
		Password: "password123",
		Nickname: "Middleware User",
		DeviceID: "device-1",
	}
	regBody, _ := json.Marshal(regReq)
	regW := httptest.NewRecorder()
	regR, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(regBody)))
	regR.Header.Set("Content-Type", "application/json")
	regR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(regW, regR)

	var regResponse map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResponse)
	accessToken := regResponse["access_token"].(string)

	t.Run("valid access token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/me", nil)
		r.Header.Set("Authorization", "Bearer "+accessToken)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		var response UserResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "middleware@example.com", response.Email)
		assert.Equal(t, "Middleware User", response.Nickname)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/me", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "authorization header required", response["error"])
	})

	t.Run("invalid token format", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/me", nil)
		r.Header.Set("Authorization", "InvalidFormat")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid access token", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/me", nil)
		r.Header.Set("Authorization", "Bearer invalid-token")

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestLogout(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	// 先注册并登录获取 token
	regReq := RegisterRequest{
		Email:    "logout@example.com",
		Password: "password123",
		Nickname: "Logout User",
		DeviceID: "device-1",
	}
	regBody, _ := json.Marshal(regReq)
	regW := httptest.NewRecorder()
	regR, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(regBody)))
	regR.Header.Set("Content-Type", "application/json")
	regR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(regW, regR)

	var regResponse map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResponse)

	type TokenPairLocal struct {
		AccessToken  string
		RefreshToken string
	}
	tokenPair := TokenPairLocal{
		AccessToken:  regResponse["access_token"].(string),
		RefreshToken: regResponse["refresh_token"].(string),
	}

	t.Run("successful logout", func(t *testing.T) {
		req := LogoutRequest{
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/logout", strings.NewReader(string(body)))
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证 token 已被撤销
		// 尝试使用 refresh token 应该失败
		refreshReq := RefreshRequest{
			RefreshToken: tokenPair.RefreshToken,
		}
		refreshBody, _ := json.Marshal(refreshReq)
		w2 := httptest.NewRecorder()
		r2, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(string(refreshBody)))
		r2.Header.Set("Content-Type", "application/json")
		r2.Header.Set("X-Client-Type", "app")

		router.ServeHTTP(w2, r2)

		assert.Equal(t, http.StatusUnauthorized, w2.Code)
	})

	t.Run("logout without auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/logout", nil)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestLogoutAll(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	// 先注册一个用户
	regReq := RegisterRequest{
		Email:    "logoutall@example.com",
		Password: "password123",
		Nickname: "Logout All User",
		DeviceID: "device-1",
	}
	regBody, _ := json.Marshal(regReq)
	regW := httptest.NewRecorder()
	regR, _ := http.NewRequest("POST", "/auth/register", strings.NewReader(string(regBody)))
	regR.Header.Set("Content-Type", "application/json")
	regR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(regW, regR)

	var regResponse map[string]interface{}
	json.Unmarshal(regW.Body.Bytes(), &regResponse)

	// 第一个设备的 token
	type TokenPairLocal struct {
		AccessToken  string
		RefreshToken string
	}
	tokenPair1 := TokenPairLocal{
		AccessToken:  regResponse["access_token"].(string),
		RefreshToken: regResponse["refresh_token"].(string),
	}

	// 再次登录获取第二个设备的 token
	loginReq := LoginRequest{
		Email:    "logoutall@example.com",
		Password: "password123",
		DeviceID: "device-2",
	}
	loginBody, _ := json.Marshal(loginReq)
	loginW := httptest.NewRecorder()
	loginR, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(loginBody)))
	loginR.Header.Set("Content-Type", "application/json")
	loginR.Header.Set("X-Client-Type", "app")
	router.ServeHTTP(loginW, loginR)

	var loginResponse map[string]interface{}
	json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
	tokenPair2 := TokenPairLocal{
		AccessToken:  loginResponse["access_token"].(string),
		RefreshToken: loginResponse["refresh_token"].(string),
	}

	t.Run("successful logout all", func(t *testing.T) {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/auth/logout_all", nil)
		r.Header.Set("Authorization", "Bearer "+tokenPair1.AccessToken)

		router.ServeHTTP(w, r)

		assert.Equal(t, http.StatusOK, w.Code)

		// 验证所有设备的 token 都已被撤销
		// 尝试使用任何 refresh token 都应该失败
		refreshReq1 := RefreshRequest{RefreshToken: tokenPair1.RefreshToken}
		refreshBody1, _ := json.Marshal(refreshReq1)
		w2 := httptest.NewRecorder()
		r2, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(string(refreshBody1)))
		r2.Header.Set("Content-Type", "application/json")
		r2.Header.Set("X-Client-Type", "app")
		router.ServeHTTP(w2, r2)
		assert.Equal(t, http.StatusUnauthorized, w2.Code)

		refreshReq2 := RefreshRequest{RefreshToken: tokenPair2.RefreshToken}
		refreshBody2, _ := json.Marshal(refreshReq2)
		w3 := httptest.NewRecorder()
		r3, _ := http.NewRequest("POST", "/auth/refresh", strings.NewReader(string(refreshBody2)))
		r3.Header.Set("Content-Type", "application/json")
		r3.Header.Set("X-Client-Type", "app")
		router.ServeHTTP(w3, r3)
		assert.Equal(t, http.StatusUnauthorized, w3.Code)
	})
}

func TestRateLimiting(t *testing.T) {
	router, _, _, _ := setupTestAuthRouter()

	t.Run("rate limit exceeded for login", func(t *testing.T) {
		req := LoginRequest{
			Email:    "ratelimit@example.com",
			Password: "wrongpassword",
			DeviceID: "device-1",
		}

		body, _ := json.Marshal(req)

		// 发送超过限制的请求（限流器配置为 10 次/5分钟）
		for i := 0; i < 11; i++ {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("POST", "/auth/login", strings.NewReader(string(body)))
			r.Header.Set("Content-Type", "application/json")
			r.RemoteAddr = "192.168.1.1:12345" // 固定 IP

			router.ServeHTTP(w, r)

			if i < 10 {
				// 前 10 次应该返回 401（密码错误）
				assert.Equal(t, http.StatusUnauthorized, w.Code)
			} else {
				// 第 11 次应该被限流
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
			}
		}
	})
}
