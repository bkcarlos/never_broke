package installment

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/ledger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler 分期处理器
type Handler struct{}

// NewHandler 创建分期处理器
func NewHandler() *Handler {
	return &Handler{}
}

// CreateExpenseInstallmentHandler 创建支出分期
func (h *Handler) CreateExpenseInstallmentHandler(c *gin.Context) {
	var req CreateExpenseInstallmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.AccountID == "" || req.TotalAmount <= 0 || req.TermMonths <= 1 {
		response.BadRequest(c, "invalid params")
		return
	}

	if req.Currency == "" {
		req.Currency = "CNY"
	}

	if req.Principal <= 0 {
		req.Principal = req.TotalAmount
	}

	// 计算每期金额（简单平均，最后一期补差）
	monthly := req.TotalAmount / float64(req.TermMonths)

	now := time.Now().UTC()
	plan := InstallmentPlan{
		ID:            uuid.NewString(),
		AccountID:     req.AccountID,
		TotalAmount:   req.TotalAmount,
		Principal:     req.Principal,
		FeeTotal:      req.FeeTotal,
		TermMonths:    req.TermMonths,
		MonthlyAmount: monthly,
		Currency:      req.Currency,
		FirstDueDate:  req.FirstDueDate,
		Status:        InstallmentStatusActive,
		CreatedAt:     now.Format(time.RFC3339),
	}
	plan = AddPlan(plan)

	// 生成每期还款交易
	var txs []ledger.Transaction
	dueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		logger.Warn("Invalid firstDueDate format", zap.Error(err))
		response.BadRequest(c, "invalid firstDueDate format, expected YYYY-MM-DD")
		return
	}

	totalAssigned := 0.0

	for i := 0; i < req.TermMonths; i++ {
		amount := monthly
		if i == req.TermMonths-1 {
			// 最后一期用总额减已分配，避免小数误差
			amount = req.TotalAmount - totalAssigned
		}
		totalAssigned += amount

		occ := dueDate.AddDate(0, i, 0) // 每期加一个月

		tx := ledger.Transaction{
			ID: uuid.NewString(),
			CreateTransactionRequest: ledger.CreateTransactionRequest{
				Type:        ledger.TransactionTypeExpense,
				Amount:      amount,
				CategoryKey: req.CategoryKey,
				Remark:      req.Remark,
				OccurredAt:  occ.Format(time.RFC3339),
				AccountID:   req.AccountID,
			},
			InstallmentPlanID: plan.ID,
			InstallmentPeriod: i + 1,
			InstallmentTotal:  req.TermMonths,
		}

		// TODO: 存入交易存储（如有），目前可以先只返回
		txs = append(txs, tx)
	}

	type Response struct {
		Plan         InstallmentPlan      `json:"plan"`
		Transactions []ledger.Transaction `json:"transactions"`
	}

	resp := Response{
		Plan:         plan,
		Transactions: txs,
	}

	logger.Info("Installment plan created",
		zap.String("planId", plan.ID),
		zap.Int("termMonths", req.TermMonths),
		zap.Int("transactionsCount", len(txs)),
	)

	response.OK(c, resp)
}
