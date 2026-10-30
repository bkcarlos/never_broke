package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

func TestMCPKeyScope(t *testing.T) {
	users := NewMockUserRepository()
	require.NoError(t, users.CreateUser(&User{Email: "mcp@example.test"}))
	h := NewTokenHandler(NewTokenService(users, NewMockTokenRepository(), NewTokenManager()))
	r := gin.New()
	r.GET("/mcp", h.MCPAuthMiddleware("test-mcp-key", 1), func(c *gin.Context) { require.Equal(t, "1", c.GetString("user_id")); c.Status(204) })
	r.GET("/api", h.AuthMiddleware(), func(c *gin.Context) { c.Status(204) })
	for _, tc := range []struct {
		path, key string
		status    int
	}{{"/mcp", "test-mcp-key", 204}, {"/mcp", "wrong", 401}, {"/mcp", "", 401}, {"/api", "test-mcp-key", 401}} {
		req := httptest.NewRequest("GET", tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+tc.key)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		require.Equal(t, tc.status, w.Code)
	}
	delete(users.users, "mcp@example.test")
	req := httptest.NewRequest("GET", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer test-mcp-key")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, 401, w.Code)
}
