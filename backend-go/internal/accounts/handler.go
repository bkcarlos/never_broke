package accounts

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
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

// CreateAccountHandler 创建账户
func (h *Handler) CreateAccountHandler(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	// 校验 Name 非空
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	// 校验 Type 合法
	validTypes := map[AccountType]bool{
		AccountTypeCash:   true,
		AccountTypeBank:   true,
		AccountTypeWallet: true,
		AccountTypeCredit: true,
	}
	if !validTypes[req.Type] {
		response.BadRequest(c, "invalid account type")
		return
	}

	// 校验 Currency 非空
	if req.Currency == "" {
		response.BadRequest(c, "currency is required")
		return
	}

	// 生成 ID
	id := uuid.NewString()

	// 创建账户
	acc := Account{
		ID:       id,
		Name:     req.Name,
		Type:     req.Type,
		Currency: req.Currency,
		Balance:  req.Balance,
		Archived: false,
	}

	// 添加到存储
	result := AddAccount(acc)

	logger.Info("Account created", zap.String("id", result.ID), zap.String("name", result.Name))
	response.OK(c, result)
}
