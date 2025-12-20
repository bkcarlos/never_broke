package installment

import (
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 分期处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建分期处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
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

// ListInstallmentPlansHandler 列出所有分期计划
func (h *Handler) ListInstallmentPlansHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	status := c.Query("status")

	var dbPlans []models.InstallmentPlan
	query := h.db.Where("user_id = ?", userID.(string))
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&dbPlans).Error; err != nil {
		logger.Error("Failed to list installment plans", zap.Error(err))
		response.InternalError(c, "failed to list installment plans")
		return
	}

	// 转换为响应格式
	result := make([]InstallmentPlan, len(dbPlans))
	for i, p := range dbPlans {
		result[i] = InstallmentPlan{
			ID:            p.ID,
			AccountID:     p.AccountID,
			Principal:     p.Principal,
			TotalInterest: p.TotalInterest,
			TotalAmount:   p.TotalAmount,
			TermMonths:    p.TermMonths,
			MonthlyRate:   p.MonthlyRate,
			RepayMethod:   InstallmentRepayMethod(p.RepayMethod),
			Currency:      p.Currency,
			FirstDueDate:  p.FirstDueDate.Format("2006-01-02"),
			Status:        InstallmentPlanStatus(p.Status),
			CreatedAt:     p.CreatedAt.Format(time.RFC3339),
		}
	}

	response.OK(c, result)
}

// CreateExpenseInstallmentHandler 创建支出分期
func (h *Handler) CreateExpenseInstallmentHandler(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

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

	// 验证账户属于当前用户
	var account models.Account
	if err := h.db.Where("id = ? AND user_id = ?", req.AccountID, userID.(string)).First(&account).Error; err != nil {
		response.BadRequest(c, "account not found or access denied")
		return
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

	// 解析首期还款日
	firstDueDate, err := time.Parse("2006-01-02", req.FirstDueDate)
	if err != nil {
		response.BadRequest(c, "invalid firstDueDate format")
		return
	}

	now := time.Now().UTC()
	dbPlan := &models.InstallmentPlan{
		ID:            uuid.NewString(),
		UserID:        userID.(string),
		AccountID:     req.AccountID,
		Principal:     scheduleResp.Principal,
		TotalInterest: scheduleResp.TotalInterest,
		TotalAmount:   scheduleResp.TotalAmount,
		TermMonths:    scheduleResp.TermMonths,
		MonthlyRate:   scheduleResp.MonthlyRate,
		RepayMethod:   string(scheduleResp.RepayMethod),
		Currency:      req.Currency,
		FirstDueDate:  firstDueDate,
		Status:        string(InstallmentStatusActive),
		CreatedAt:     now,
	}

	// 使用事务保存分期计划和交易记录
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 保存分期计划
		if err := tx.Create(dbPlan).Error; err != nil {
			return err
		}

		// 生成每期交易并保存
		for _, item := range scheduleResp.Schedule {
			occ, err := time.Parse("2006-01-02", item.DueDate)
			if err != nil {
				continue
			}

			dbTx := &models.Transaction{
				ID:                uuid.NewString(),
				UserID:            userID.(string),
				Type:              "expense",
				Amount:            item.Total,
				OccurredAt:        occ.Format(time.RFC3339),
				InstallmentPlanID: &dbPlan.ID,
				InstallmentPeriod: &item.Period,
				InstallmentTotal:  &scheduleResp.TermMonths,
			}
			if req.CategoryKey != "" {
				dbTx.CategoryKey = &req.CategoryKey
			}
			if req.Remark != "" {
				dbTx.Remark = &req.Remark
			}
			if req.AccountID != "" {
				dbTx.AccountID = &req.AccountID
			}

			if err := tx.Create(dbTx).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logger.Error("Failed to create installment plan", zap.Error(err))
		response.InternalError(c, "failed to create installment plan")
		return
	}

	// 构造响应
	plan := InstallmentPlan{
		ID:            dbPlan.ID,
		AccountID:     dbPlan.AccountID,
		Principal:     dbPlan.Principal,
		TotalInterest: dbPlan.TotalInterest,
		TotalAmount:   dbPlan.TotalAmount,
		TermMonths:    dbPlan.TermMonths,
		MonthlyRate:   dbPlan.MonthlyRate,
		RepayMethod:   InstallmentRepayMethod(dbPlan.RepayMethod),
		Currency:      dbPlan.Currency,
		FirstDueDate:  dbPlan.FirstDueDate.Format("2006-01-02"),
		Status:        InstallmentStatusActive,
		CreatedAt:     dbPlan.CreatedAt.Format(time.RFC3339),
	}

	// 生成交易响应
	var txs []ledger.Transaction
	for _, item := range scheduleResp.Schedule {
		occ, _ := time.Parse("2006-01-02", item.DueDate)
		txs = append(txs, ledger.Transaction{
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
		})
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
		zap.String("userID", userID.(string)),
		zap.String("planId", plan.ID),
		zap.Int("termMonths", req.TermMonths),
		zap.String("repayMethod", string(req.RepayMethod)),
		zap.Int("transactionsCount", len(txs)),
	)

	response.OK(c, resp)
}
