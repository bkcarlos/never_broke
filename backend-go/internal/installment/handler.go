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

// PreviewInstallmentHandler 预览分期
func (h *Handler) PreviewInstallmentHandler(c *gin.Context) {
	var req PreviewInstallmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	resp, err := BuildSchedule(req)
	if err != nil {
		logger.Warn("Build schedule failed", zap.Error(err))
		response.BadRequest(c, "build schedule failed")
		return
	}

	response.OK(c, resp)
}

// CreateExpenseInstallmentHandler 创建支出分期
func (h *Handler) CreateExpenseInstallmentHandler(c *gin.Context) {
	var req CreateExpenseInstallmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.AccountID == "" || req.Principal <= 0 || req.TermMonths <= 1 {
		response.BadRequest(c, "invalid params")
		return
	}

	if req.Currency == "" {
		req.Currency = "CNY"
	}

	previewReq := PreviewInstallmentRequest{
		Principal:          req.Principal,
		TermMonths:         req.TermMonths,
		MonthlyRatePercent: req.MonthlyRatePercent,
		RepayMethod:        req.RepayMethod,
		FirstDueDate:       req.FirstDueDate,
		Currency:           req.Currency,
	}

	scheduleResp, err := BuildSchedule(previewReq)
	if err != nil {
		logger.Warn("Build schedule failed", zap.Error(err))
		response.BadRequest(c, "build schedule failed")
		return
	}

	now := time.Now().UTC()
	plan := InstallmentPlan{
		ID:            uuid.NewString(),
		AccountID:     req.AccountID,
		Principal:     scheduleResp.Principal,
		TotalInterest: scheduleResp.TotalInterest,
		TotalAmount:   scheduleResp.TotalAmount,
		TermMonths:    scheduleResp.TermMonths,
		MonthlyRate:   scheduleResp.MonthlyRate,
		RepayMethod:   scheduleResp.RepayMethod,
		Currency:      req.Currency,
		FirstDueDate:  req.FirstDueDate,
		Status:        InstallmentStatusActive,
		CreatedAt:     now.Format(time.RFC3339),
	}
	plan = AddPlan(plan)

	// 生成每期交易
	var txs []ledger.Transaction
	for _, item := range scheduleResp.Schedule {
		occ, err := time.Parse("2006-01-02", item.DueDate)
		if err != nil {
			logger.Warn("Invalid dueDate format", zap.String("dueDate", item.DueDate), zap.Error(err))
			continue
		}

		tx := ledger.Transaction{
			ID: uuid.NewString(),
			CreateTransactionRequest: ledger.CreateTransactionRequest{
				Type:        ledger.TransactionTypeExpense,
				Amount:      item.Total,
				CategoryKey: req.CategoryKey,
				Remark:      req.Remark,
				OccurredAt:  occ.Format(time.RFC3339),
				AccountID:   req.AccountID,
			},
			InstallmentPlanID: plan.ID,
			InstallmentPeriod: item.Period,
			InstallmentTotal:  scheduleResp.TermMonths,
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
		zap.String("repayMethod", string(req.RepayMethod)),
		zap.Int("transactionsCount", len(txs)),
	)

	response.OK(c, resp)
}
