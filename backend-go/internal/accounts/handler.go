package accounts

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	dbModels "neverbroke/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 账户处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建账户处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// ListAccountsHandler 列出所有账户
func (h *Handler) ListAccountsHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	// 从数据库查询用户的账户
	var dbAccounts []dbModels.Account
	if err := h.db.Where("user_id = ? AND archived = ?", userID.(string), false).Find(&dbAccounts).Error; err != nil {
		logger.Error("Failed to fetch accounts from database", zap.Error(err))
		response.InternalServerError(c, "failed to fetch accounts")
		return
	}

	// 转换为响应格式
	accounts := make([]Account, len(dbAccounts))
	for i, dbAcc := range dbAccounts {
		accounts[i] = Account{
			ID:          dbAcc.ID,
			Type:        AccountType(dbAcc.Type),
			Institution: AccountInstitution(stringPtrToString(dbAcc.Institution)),
			CardLast4:   stringPtrToString(dbAcc.CardLast4),
			Currency:    dbAcc.Currency,
			Balance:     dbAcc.Balance,
			CreditLimit: float64PtrToFloat(dbAcc.CreditLimit),
			TempLimit:   float64PtrToFloat(dbAcc.TempLimit),
			Archived:    dbAcc.Archived,
		}
	}

	response.OK(c, accounts)
}

// 辅助函数：将 *string 转换为 string
func stringPtrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// 辅助函数：将 *float64 转换为 float64
func float64PtrToFloat(ptr *float64) float64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// CreateAccountHandler 创建账户
func (h *Handler) CreateAccountHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

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
		var count int64
		h.db.Model(&dbModels.Account{}).Where("user_id = ? AND type = ? AND archived = ?", userID.(string), AccountTypeCash, false).Count(&count)
		if count > 0 {
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

	// 准备数据库模型
	dbAcc := dbModels.Account{
		ID:       id,
		UserID:   userID.(string),
		Type:     string(req.Type),
		Currency: req.Currency,
		Balance:  req.Balance,
		Archived: false,
	}

	// 设置可选字段
	if req.Institution != "" {
		inst := string(req.Institution)
		dbAcc.Institution = &inst
	}
	if req.CardLast4 != "" {
		dbAcc.CardLast4 = &req.CardLast4
	}
	if req.CreditLimit > 0 {
		dbAcc.CreditLimit = &req.CreditLimit
	}
	if req.TempLimit > 0 {
		dbAcc.TempLimit = &req.TempLimit
	}

	// 保存到数据库
	if err := h.db.Create(&dbAcc).Error; err != nil {
		logger.Error("Failed to create account in database", zap.Error(err))
		response.InternalServerError(c, "failed to create account")
		return
	}

	// 创建响应
	acc := Account{
		ID:          dbAcc.ID,
		Type:        req.Type,
		Institution: req.Institution,
		CardLast4:   req.CardLast4,
		Currency:    dbAcc.Currency,
		Balance:     dbAcc.Balance,
		CreditLimit: req.CreditLimit,
		TempLimit:   req.TempLimit,
		Archived:    false,
	}

	logger.Info("Account created", zap.String("id", acc.ID), zap.String("type", string(acc.Type)))
	response.OK(c, acc)
}

// UpdateCreditLimitsHandler 更新信用卡额度
func (h *Handler) UpdateCreditLimitsHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	id := c.Param("id")
	var req UpdateCreditLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	// 从数据库查询账户
	var dbAcc dbModels.Account
	if err := h.db.Where("id = ? AND user_id = ?", id, userID.(string)).First(&dbAcc).Error; err != nil {
		logger.Warn("Account not found", zap.String("id", id), zap.Error(err))
		response.BadRequest(c, "account not found")
		return
	}

	if dbAcc.Type != string(AccountTypeCredit) {
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
	updates := make(map[string]interface{})
	if req.CreditLimit != nil {
		updates["credit_limit"] = *req.CreditLimit
	}
	if req.TempLimit != nil {
		updates["temp_limit"] = *req.TempLimit
	}

	if err := h.db.Model(&dbAcc).Updates(updates).Error; err != nil {
		logger.Error("Failed to update account limits", zap.Error(err))
		response.InternalServerError(c, "failed to update account")
		return
	}

	// 重新查询以获取最新数据
	if err := h.db.Where("id = ?", id).First(&dbAcc).Error; err != nil {
		logger.Error("Failed to fetch updated account", zap.Error(err))
		response.InternalServerError(c, "failed to fetch updated account")
		return
	}

	// 创建响应
	result := Account{
		ID:          dbAcc.ID,
		Type:        AccountType(dbAcc.Type),
		Institution: AccountInstitution(stringPtrToString(dbAcc.Institution)),
		CardLast4:   stringPtrToString(dbAcc.CardLast4),
		Currency:    dbAcc.Currency,
		Balance:     dbAcc.Balance,
		CreditLimit: float64PtrToFloat(dbAcc.CreditLimit),
		TempLimit:   float64PtrToFloat(dbAcc.TempLimit),
		Archived:    dbAcc.Archived,
	}

	response.OK(c, result)
}
