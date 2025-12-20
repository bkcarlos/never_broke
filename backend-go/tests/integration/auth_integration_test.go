package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"neverbroke/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthFlow_RegisterLoginLogout 测试完整的认证流程
func TestAuthFlow_RegisterLoginLogout(t *testing.T) {
	// 清理测试数据
	cleanupTestData(testDB)

	// 1. 注册新用户
	t.Run("Register", func(t *testing.T) {
		registerReq := auth.RegisterRequest{
			Email:    "integration_test@example.com",
			Password: "TestPassword123!",
			Nickname: "Integration Test User",
			DeviceID: "test-device-1",
		}

		body, _ := json.Marshal(registerReq)
		req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Type", "app")

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code, "Registration should succeed")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")

		// 等待数据库写入完成
		waitForDB()

		// 验证用户已写入数据库
		var user auth.User
		err = testDB.Where("email = ?", "integration_test@example.com").First(&user).Error
		require.NoError(t, err, "User should be persisted in database")
		assert.Equal(t, "integration_test@example.com", user.Email)
		assert.Equal(t, "Integration Test User", user.Nickname)
		assert.NotEmpty(t, user.PasswordHash)

		// 验证 refresh token 已写入数据库
		var tokenCount int64
		testDB.Model(&auth.RefreshToken{}).Where("user_id = ?", user.ID).Count(&tokenCount)
		assert.Equal(t, int64(1), tokenCount, "Refresh token should be persisted")
	})

	// 2. 使用正确的密码登录
	t.Run("Login_Success", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			Email:    "integration_test@example.com",
			Password: "TestPassword123!",
			DeviceID: "test-device-2",
		}

		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Type", "app")

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		require.Equal(t, http.StatusOK, w.Code, "Login should succeed")

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")

		// 等待数据库写入
		waitForDB()

		// 验证新的 refresh token 已写入（现在应该有2个设备的token）
		var user auth.User
		testDB.Where("email = ?", "integration_test@example.com").First(&user)

		var tokenCount int64
		testDB.Model(&auth.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", user.ID).
			Count(&tokenCount)
		assert.Equal(t, int64(2), tokenCount, "Should have 2 active refresh tokens")
	})

	// 3. 使用错误的密码登录
	t.Run("Login_WrongPassword", func(t *testing.T) {
		loginReq := auth.LoginRequest{
			Email:    "integration_test@example.com",
			Password: "WrongPassword123!",
			DeviceID: "test-device-3",
		}

		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Client-Type", "app")

		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code, "Login should fail with wrong password")

		// 验证没有新增 token
		var user auth.User
		testDB.Where("email = ?", "integration_test@example.com").First(&user)

		var tokenCount int64
		testDB.Model(&auth.RefreshToken{}).
			Where("user_id = ? AND revoked_at IS NULL", user.ID).
			Count(&tokenCount)
		assert.Equal(t, int64(2), tokenCount, "Token count should not change")
	})

	// 4. 登出
	t.Run("Logout", func(t *testing.T) {
		// 先登录获取 access token
		loginReq := auth.LoginRequest{
			Email:    "integration_test@example.com",
			Password: "TestPassword123!",
			DeviceID: "test-device-logout",
		}

		body, _ := json.Marshal(loginReq)
		loginW := httptest.NewRecorder()
		loginR, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
		loginR.Header.Set("Content-Type", "application/json")
		loginR.Header.Set("X-Client-Type", "app")
		testRouter.ServeHTTP(loginW, loginR)

		var loginResponse map[string]interface{}
		json.Unmarshal(loginW.Body.Bytes(), &loginResponse)
		accessToken := loginResponse["access_token"].(string)

		waitForDB()

		// 执行登出
		logoutReq := auth.LogoutRequest{
			DeviceID: "test-device-logout",
		}
		logoutBody, _ := json.Marshal(logoutReq)
		logoutR, _ := http.NewRequest("POST", "/api/auth/logout", bytes.NewBuffer(logoutBody))
		logoutR.Header.Set("Content-Type", "application/json")
		logoutR.Header.Set("Authorization", "Bearer "+accessToken)

		logoutW := httptest.NewRecorder()
		testRouter.ServeHTTP(logoutW, logoutR)

		require.Equal(t, http.StatusOK, logoutW.Code, "Logout should succeed")

		waitForDB()

		// 验证该设备的 token 已被撤销
		var user auth.User
		testDB.Where("email = ?", "integration_test@example.com").First(&user)

		var revokedToken auth.RefreshToken
		err := testDB.Where("user_id = ? AND device_id = ?", user.ID, "test-device-logout").
			First(&revokedToken).Error
		require.NoError(t, err)
		assert.NotNil(t, revokedToken.RevokedAt, "Token should be revoked")
	})
}

// TestAuthFlow_TokenRefresh 测试 Token 刷新流程
func TestAuthFlow_TokenRefresh(t *testing.T) {
	cleanupTestData(testDB)

	// 注册并登录
	registerReq := auth.RegisterRequest{
		Email:    "refresh_test@example.com",
		Password: "TestPassword123!",
		Nickname: "Refresh Test User",
		DeviceID: "refresh-device-1",
	}

	body, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Client-Type", "app")

	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	var registerResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResponse)
	oldRefreshToken := registerResponse["refresh_token"].(string)

	waitForDB()

	// 测试 Token 刷新
	t.Run("Refresh_Success", func(t *testing.T) {
		refreshReq := auth.RefreshRequest{
			RefreshToken: oldRefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		refreshR, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(refreshBody))
		refreshR.Header.Set("Content-Type", "application/json")
		refreshR.Header.Set("X-Client-Type", "app")

		refreshW := httptest.NewRecorder()
		testRouter.ServeHTTP(refreshW, refreshR)

		require.Equal(t, http.StatusOK, refreshW.Code, "Token refresh should succeed")

		var refreshResponse map[string]interface{}
		json.Unmarshal(refreshW.Body.Bytes(), &refreshResponse)

		assert.Contains(t, refreshResponse, "access_token")
		assert.Contains(t, refreshResponse, "refresh_token")

		newRefreshToken := refreshResponse["refresh_token"].(string)
		assert.NotEqual(t, oldRefreshToken, newRefreshToken, "New refresh token should be different")

		waitForDB()

		// 验证旧 token 已被标记为 replaced
		var user auth.User
		testDB.Where("email = ?", "refresh_test@example.com").First(&user)

		var oldToken auth.RefreshToken
		err := testDB.Where("user_id = ? AND revoked_at IS NOT NULL", user.ID).
			Order("issued_at ASC").First(&oldToken).Error
		require.NoError(t, err)
		assert.NotNil(t, oldToken.ReplacedBy, "Old token should have replacedBy set")
	})

	// 测试使用已撤销的 token 刷新（应该失败）
	t.Run("Refresh_RevokedToken", func(t *testing.T) {
		refreshReq := auth.RefreshRequest{
			RefreshToken: oldRefreshToken, // 使用已经被替换的旧 token
		}

		refreshBody, _ := json.Marshal(refreshReq)
		refreshR, _ := http.NewRequest("POST", "/api/auth/refresh", bytes.NewBuffer(refreshBody))
		refreshR.Header.Set("Content-Type", "application/json")
		refreshR.Header.Set("X-Client-Type", "app")

		refreshW := httptest.NewRecorder()
		testRouter.ServeHTTP(refreshW, refreshR)

		assert.Equal(t, http.StatusUnauthorized, refreshW.Code, "Refresh with revoked token should fail")
	})
}

// TestAuthFlow_ConcurrentRegistrations 测试并发注册（防止重复邮箱）
func TestAuthFlow_ConcurrentRegistrations(t *testing.T) {
	cleanupTestData(testDB)

	email := "concurrent_test@example.com"
	results := make(chan int, 5)

	// 并发尝试注册相同邮箱
	for i := 0; i < 5; i++ {
		go func(index int) {
			registerReq := auth.RegisterRequest{
				Email:    email,
				Password: "TestPassword123!",
				Nickname: "Concurrent User",
				DeviceID: "device-" + string(rune(index)),
			}

			body, _ := json.Marshal(registerReq)
			req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Client-Type", "app")

			w := httptest.NewRecorder()
			testRouter.ServeHTTP(w, req)

			results <- w.Code
		}(i)
	}

	// 收集结果
	successCount := 0
	failCount := 0

	for i := 0; i < 5; i++ {
		code := <-results
		if code == http.StatusCreated {
			successCount++
		} else {
			failCount++
		}
	}

	// 验证只有一个注册成功
	assert.Equal(t, 1, successCount, "Only one registration should succeed")
	assert.Equal(t, 4, failCount, "Four registrations should fail")

	waitForDB()

	// 验证数据库中只有一个用户
	var count int64
	testDB.Model(&auth.User{}).Where("email = ?", email).Count(&count)
	assert.Equal(t, int64(1), count, "Only one user should exist in database")
}

// TestAuthFlow_SessionExpiry 测试会话过期
func TestAuthFlow_SessionExpiry(t *testing.T) {
	t.Skip("This test requires time manipulation, skipping for now")

	// TODO: 实现会话过期测试
	// 需要模拟时间流逝或修改 token 过期时间
}
