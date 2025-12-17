package ledger

import (
	"context"
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransactionStore 交易存储接口（避免循环依赖）
type TransactionStore interface {
	ListByUserBetween(ctx context.Context, userID, from, to string) ([]*Transaction, error)
}

// Handler 账本处理器
type Handler struct {
	txStore TransactionStore
}

// NewHandler 创建账本处理器
func NewHandler(txStore TransactionStore) *Handler {
	return &Handler{
		txStore: txStore,
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

	// 使用固定用户 ID
	userID := "demo"
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

	// 更新账户余额
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

	// 创建交易记录
	tx := Transaction{
		ID:                       uuid.NewString(),
		CreateTransactionRequest: req,
	}

	// TODO: 将交易记录保存到内存/DB，这里先返回成功
	logger.Info("Transaction created successfully",
		zap.String("id", tx.ID),
		zap.String("type", string(req.Type)),
		zap.Float64("amount", req.Amount),
	)
	response.OK(c, tx)
}
