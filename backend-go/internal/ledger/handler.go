package ledger

import (
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/storage/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler 账本处理器
type Handler struct{}

// NewHandler 创建账本处理器
func NewHandler() *Handler {
	return &Handler{}
}

// GetDailyTransactions 获取每日账单
func (h *Handler) GetDailyTransactions(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		logger.Warn("Missing date parameter in daily transactions request")
		response.BadRequest(c, "date parameter is required")
		return
	}

	logger.Debug("Fetching daily transactions", zap.String("date", date))

	// Mock 数据
	dailyData := models.DailyTransactionResponse{
		Date: date,
		Stats: models.DailyStats{
			TotalIncome:  500.00,
			TotalExpense: 156.80,
		},
		List: []models.DailyTransaction{
			{
				ID:          "tx001",
				Type:        "expense",
				Amount:      45.50,
				CategoryKey: "food",
				Remark:      "午餐",
				Time:        "12:30",
			},
			{
				ID:          "tx002",
				Type:        "expense",
				Amount:      28.00,
				CategoryKey: "transport",
				Remark:      "地铁",
				Time:        "08:15",
			},
			{
				ID:          "tx003",
				Type:        "expense",
				Amount:      83.30,
				CategoryKey: "shopping",
				Remark:      "日用品",
				Time:        "19:45",
			},
			{
				ID:          "tx004",
				Type:        "income",
				Amount:      500.00,
				CategoryKey: "salary",
				Remark:      "工资",
				Time:        "10:00",
			},
		},
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
