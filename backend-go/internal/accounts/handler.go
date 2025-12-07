package accounts

import (
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
)

// Handler 账户处理器
type Handler struct{}

// NewHandler 创建账户处理器
func NewHandler() *Handler {
	return &Handler{}
}

// ListAccountsHandler 列出所有账户
func (h *Handler) ListAccountsHandler(c *gin.Context) {
	list := ListAccounts()
	response.OK(c, list)
}
