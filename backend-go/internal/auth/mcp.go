package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// MCPAuthMiddleware accepts an optional deployment-managed key scoped to one
// existing user. The key is accepted only at /mcp; REST keeps normal JWT auth.
// Rotate/revoke the key by updating configuration and restarting the service.
func (h *TokenHandler) MCPAuthMiddleware(key string, userID int) gin.HandlerFunc {
	jwtAuth := h.AuthMiddleware()
	expected := sha256.Sum256([]byte(key))
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		supplied := sha256.Sum256([]byte(strings.TrimPrefix(header, "Bearer ")))
		if key != "" && userID > 0 && strings.HasPrefix(header, "Bearer ") && subtle.ConstantTimeCompare(expected[:], supplied[:]) == 1 {
			user, err := h.service.userRepo.GetUserByID(userID)
			if err != nil || user == nil {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			token, _, err := h.service.tokenMgr.GenerateAccessToken(user, "")
			if err != nil {
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			// The in-process tool adapter forwards this short-lived credential through
			// ordinary REST authentication, preserving ownership checks.
			c.Request.Header.Set("Authorization", "Bearer "+token)
		}
		jwtAuth(c)
	}
}
