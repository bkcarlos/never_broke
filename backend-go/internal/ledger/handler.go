package ledger

import (
	"context"
	"fmt"
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	dbModels "neverbroke/internal/models"
	"neverbroke/internal/storage/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionStore 交易存储接口（避免循环依赖）
type TransactionStore interface {
	ListByUserBetween(ctx context.Context, userID, from, to string) ([]*Transaction, error)
	Create(ctx context.Context, userID string, tx *Transaction) error
}

// Handler 账本处理器
type Handler struct {
	txStore TransactionStore
	db      *gorm.DB
}

// NewHandler 创建账本处理器
func NewHandler(txStore TransactionStore, db *gorm.DB) *Handler {
	return &Handler{
		txStore: txStore,
		db:      db,
	}
}

// GetDailyTransactions 获取每日账单
func (h *Handler) GetDailyTransactions(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		logger.Warn("Missing date parameter in daily transactions request")
		response.BadRequest(c, "date parameter is required")
		return
	}

	// 从上下文获取用户ID
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.InternalError(c, "user not authenticated")
		return
	}
	userID := userIDVal.(string)
	ctx := context.Background()

	logger.Debug("Fetching daily transactions",
		zap.String("date", date),
		zap.String("userID", userID))

	// 解析日期
	dayStart, err := time.Parse("2006-01-02", date)
	if err != nil {
		logger.Warn("Invalid date format", zap.String("date", date), zap.Error(err))
		response.BadRequest(c, "invalid date format, expected YYYY-MM-DD")
		return
	}
	dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)

	// 从数据库获取当日交易
	txs, err := h.txStore.ListByUserBetween(ctx, userID,
		dayStart.Format(time.RFC3339),
		dayEnd.Format(time.RFC3339))
	if err != nil {
		logger.Error("Failed to fetch daily transactions",
			zap.String("date", date),
			zap.Error(err))
		response.InternalError(c, "Failed to fetch transactions")
		return
	}

	// 计算统计数据
	var totalIncome, totalExpense float64
	var txList []models.DailyTransaction

	for _, tx := range txs {
		// 统计收支
		switch tx.Type {
		case "income":
			totalIncome += tx.Amount
		case "expense":
			totalExpense += tx.Amount
		}

		// 转换为响应格式
		occurredTime, _ := time.Parse(time.RFC3339, tx.OccurredAt)

		txList = append(txList, models.DailyTransaction{
			ID:          tx.ID,
			Type:        string(tx.Type),
			Amount:      tx.Amount,
			CategoryKey: tx.CategoryKey,
			Remark:      tx.Remark,
			Time:        occurredTime.Format("15:04"),
		})
	}

	dailyData := models.DailyTransactionResponse{
		Date: date,
		Stats: models.DailyStats{
			TotalIncome:  totalIncome,
			TotalExpense: totalExpense,
		},
		List: txList,
	}

	logger.Info("Daily transactions fetched successfully",
		zap.String("date", date),
		zap.Int("count", len(dailyData.List)),
	)
	response.OK(c, dailyData)
}

// CreateTransaction 创建交易
func (h *Handler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.Amount <= 0 {
		response.BadRequest(c, "amount must be > 0")
		return
	}

	// 验证交易类型和账户
	switch req.Type {
	case TransactionTypeExpense, TransactionTypeIncome:
		if req.AccountID == "" {
			response.BadRequest(c, "accountId is required")
			return
		}
	case TransactionTypeTransfer:
		if req.FromAccountID == "" || req.ToAccountID == "" || req.FromAccountID == req.ToAccountID {
			response.BadRequest(c, "invalid transfer accounts")
			return
		}
	default:
		response.BadRequest(c, "invalid type")
		return
	}

	// 创建交易记录
	tx := Transaction{
		ID:                       uuid.NewString(),
		CreateTransactionRequest: req,
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.InternalError(c, "user not authenticated")
		return
	}

	// 验证账户所有权
	if h.db != nil {
		// 对于expense/income，验证accountId属于当前用户
		if req.Type == TransactionTypeExpense || req.Type == TransactionTypeIncome {
			var account dbModels.Account
			if err := h.db.Where("id = ? AND user_id = ?", req.AccountID, userID.(string)).First(&account).Error; err != nil {
				logger.Warn("Account ownership validation failed",
					zap.String("userId", userID.(string)),
					zap.String("accountId", req.AccountID),
					zap.Error(err))
				response.BadRequest(c, "account not found or access denied")
				return
			}
		}

		// 对于transfer，验证fromAccountId和toAccountId都属于当前用户
		if req.Type == TransactionTypeTransfer {
			var fromAccount, toAccount dbModels.Account
			if err := h.db.Where("id = ? AND user_id = ?", req.FromAccountID, userID.(string)).First(&fromAccount).Error; err != nil {
				logger.Warn("From account ownership validation failed",
					zap.String("userId", userID.(string)),
					zap.String("accountId", req.FromAccountID))
				response.BadRequest(c, "from account not found or access denied")
				return
			}
			if err := h.db.Where("id = ? AND user_id = ?", req.ToAccountID, userID.(string)).First(&toAccount).Error; err != nil {
				logger.Warn("To account ownership validation failed",
					zap.String("userId", userID.(string)),
					zap.String("accountId", req.ToAccountID))
				response.BadRequest(c, "to account not found or access denied")
				return
			}
		}
	}

	// 在数据库事务中更新账户余额和保存交易记录
	if h.db != nil {
		err := h.db.Transaction(func(dbTx *gorm.DB) error {
			// 更新账户余额
			switch req.Type {
			case TransactionTypeExpense:
				result := dbTx.Model(&dbModels.Account{}).
					Where("id = ?", req.AccountID).
					Update("balance", gorm.Expr("balance - ?", req.Amount))
				if result.Error != nil {
					return fmt.Errorf("failed to update account balance: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("account not found")
				}
			case TransactionTypeIncome:
				result := dbTx.Model(&dbModels.Account{}).
					Where("id = ?", req.AccountID).
					Update("balance", gorm.Expr("balance + ?", req.Amount))
				if result.Error != nil {
					return fmt.Errorf("failed to update account balance: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("account not found")
				}
			case TransactionTypeTransfer:
				// 更新转出账户
				result := dbTx.Model(&dbModels.Account{}).
					Where("id = ?", req.FromAccountID).
					Update("balance", gorm.Expr("balance - ?", req.Amount))
				if result.Error != nil {
					return fmt.Errorf("failed to update from account balance: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("from account not found")
				}

				// 更新转入账户
				result = dbTx.Model(&dbModels.Account{}).
					Where("id = ?", req.ToAccountID).
					Update("balance", gorm.Expr("balance + ?", req.Amount))
				if result.Error != nil {
					return fmt.Errorf("failed to update to account balance: %w", result.Error)
				}
				if result.RowsAffected == 0 {
					return fmt.Errorf("to account not found")
				}
			}

			// 保存交易记录
			dbTx2 := toDBTransaction(userID.(string), &tx)
			if err := dbTx.Create(dbTx2).Error; err != nil {
				return fmt.Errorf("failed to save transaction: %w", err)
			}

			return nil
		})

		if err != nil {
			logger.Error("Transaction failed", zap.Error(err))
			response.InternalError(c, err.Error())
			return
		}

		// 记录成功日志
		switch req.Type {
		case TransactionTypeExpense:
			logger.Info("Account balance updated for expense",
				zap.String("accountId", req.AccountID),
				zap.Float64("amount", req.Amount),
			)
		case TransactionTypeIncome:
			logger.Info("Account balance updated for income",
				zap.String("accountId", req.AccountID),
				zap.Float64("amount", req.Amount),
			)
		case TransactionTypeTransfer:
			logger.Info("Account balance updated for transfer",
				zap.String("fromAccountId", req.FromAccountID),
				zap.String("toAccountId", req.ToAccountID),
				zap.Float64("amount", req.Amount),
			)
		}
	} else {
		// 回退到内存操作（用于单元测试）
		switch req.Type {
		case TransactionTypeExpense:
			ok := accounts.AddToBalance(req.AccountID, -req.Amount)
			if !ok {
				response.BadRequest(c, "account not found")
				return
			}
			logger.Info("Account balance updated for expense",
				zap.String("accountId", req.AccountID),
				zap.Float64("amount", req.Amount),
			)
		case TransactionTypeIncome:
			ok := accounts.AddToBalance(req.AccountID, req.Amount)
			if !ok {
				response.BadRequest(c, "account not found")
				return
			}
			logger.Info("Account balance updated for income",
				zap.String("accountId", req.AccountID),
				zap.Float64("amount", req.Amount),
			)
		case TransactionTypeTransfer:
			okFrom := accounts.AddToBalance(req.FromAccountID, -req.Amount)
			okTo := accounts.AddToBalance(req.ToAccountID, req.Amount)
			if !okFrom || !okTo {
				response.BadRequest(c, "account not found")
				return
			}
			logger.Info("Account balance updated for transfer",
				zap.String("fromAccountId", req.FromAccountID),
				zap.String("toAccountId", req.ToAccountID),
				zap.Float64("amount", req.Amount),
			)
		}

		// 保存交易记录到内存
		if err := h.txStore.Create(c.Request.Context(), userID.(string), &tx); err != nil {
			logger.Error("Failed to save transaction", zap.Error(err))
			response.InternalError(c, "failed to save transaction")
			return
		}
	}

	logger.Info("Transaction created successfully",
		zap.String("id", tx.ID),
		zap.String("type", string(req.Type)),
		zap.Float64("amount", req.Amount),
	)
	response.OK(c, tx)
}

// toDBTransaction 将业务模型转换为数据库模型
func toDBTransaction(userID string, tx *Transaction) *dbModels.Transaction {
	dbTx := &dbModels.Transaction{
		ID:         tx.ID,
		UserID:     userID,
		Type:       string(tx.Type),
		Amount:     tx.Amount,
		OccurredAt: tx.OccurredAt,
	}

	if tx.CategoryKey != "" {
		dbTx.CategoryKey = &tx.CategoryKey
	}
	if tx.Remark != "" {
		dbTx.Remark = &tx.Remark
	}
	if tx.AccountID != "" {
		dbTx.AccountID = &tx.AccountID
	}
	if tx.FromAccountID != "" {
		dbTx.FromAccountID = &tx.FromAccountID
	}
	if tx.ToAccountID != "" {
		dbTx.ToAccountID = &tx.ToAccountID
	}
	if tx.InstallmentPlanID != "" {
		dbTx.InstallmentPlanID = &tx.InstallmentPlanID
	}
	if tx.InstallmentPeriod > 0 {
		dbTx.InstallmentPeriod = &tx.InstallmentPeriod
	}
	if tx.InstallmentTotal > 0 {
		dbTx.InstallmentTotal = &tx.InstallmentTotal
	}

	return dbTx
}
