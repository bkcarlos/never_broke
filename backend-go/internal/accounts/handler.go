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

	// 校验 Currency 非空，默认 CNY
	if req.Currency == "" {
		req.Currency = "CNY"
	}

	// 根据账户类型校验必填字段
	switch req.Type {
	case AccountTypeCash:
		// cash 时：检查是否已存在现金账户
		if HasCashAccount() {
			response.BadRequest(c, "cash account already exists")
			return
		}
	case AccountTypeBank, AccountTypeCredit:
		// bank/credit 时：institution 必填
		if req.Institution == "" {
			response.BadRequest(c, "institution is required for bank/credit accounts")
			return
		}
		// cardLast4 建议长度 2-4
		if req.CardLast4 != "" {
			cardLen := len(req.CardLast4)
			if cardLen < 2 || cardLen > 4 {
				response.BadRequest(c, "cardLast4 must be 2-4 digits")
				return
			}
		}
	case AccountTypeWallet:
		// wallet 时：institution 必填（wechat/alipay/other）
		if req.Institution == "" {
			response.BadRequest(c, "institution is required for wallet accounts")
			return
		}
	default:
		response.BadRequest(c, "invalid account type")
		return
	}

	// 生成 ID
	id := uuid.NewString()

	// 创建账户
	acc := Account{
		ID:          id,
		Type:        req.Type,
		Institution: req.Institution,
		CardLast4:   req.CardLast4,
		Currency:    req.Currency,
		Balance:     req.Balance,
		CreditLimit: req.CreditLimit,
		TempLimit:   req.TempLimit,
		Archived:    false,
	}

	// 添加到存储
	result := AddAccount(acc)

	logger.Info("Account created", zap.String("id", result.ID), zap.String("type", string(result.Type)))
	response.OK(c, result)
}

// UpdateCreditLimitsHandler 更新信用卡额度
func (h *Handler) UpdateCreditLimitsHandler(c *gin.Context) {
	id := c.Param("id")
	var req UpdateCreditLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	accPtr, ok := FindAccountByID(id)
	if !ok {
		response.BadRequest(c, "account not found")
		return
	}

	if accPtr.Type != AccountTypeCredit {
		response.BadRequest(c, "not a credit account")
		return
	}

	// 验证额度值
	if req.CreditLimit != nil && *req.CreditLimit < 0 {
		response.BadRequest(c, "creditLimit must be >= 0")
		return
	}
	if req.TempLimit != nil && *req.TempLimit < 0 {
		response.BadRequest(c, "tempLimit must be >= 0")
		return
	}

	// 更新账户额度
	var updatedAccount *Account
	ok = UpdateAccount(id, func(acc *Account) {
		if req.CreditLimit != nil {
			acc.CreditLimit = *req.CreditLimit
		}
		if req.TempLimit != nil {
			acc.TempLimit = *req.TempLimit
		}
		updatedAccount = acc
	})

	if !ok {
		response.BadRequest(c, "failed to update account")
		return
	}

	// 返回更新后的账户（需要复制以避免并发问题）
	result := *updatedAccount
	response.OK(c, result)
}
