package recurringincome

import (
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/http/response"
	"neverbroke/internal/ledger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler 周期收入处理器
type Handler struct{}

// NewHandler 创建周期收入处理器
func NewHandler() *Handler {
	return &Handler{}
}

// ListPlansHandler 列出所有周期收入计划
func (h *Handler) ListPlansHandler(c *gin.Context) {
	plans := ListPlans()
	logger.Info("List recurring income plans", zap.Int("count", len(plans)))
	response.OK(c, plans)
}

// UpsertPlanHandler 创建或更新周期收入计划
func (h *Handler) UpsertPlanHandler(c *gin.Context) {
	var req UpsertRecurringIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	// 验证必填字段
	if req.Amount <= 0 {
		response.BadRequest(c, "amount must be > 0")
		return
	}
	if req.AccountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}
	if req.NextDueDate == "" {
		response.BadRequest(c, "nextDueDate is required")
		return
	}

	// 验证账户是否存在
	_, exists := accounts.FindAccountByID(req.AccountID)
	if !exists {
		response.BadRequest(c, "account not found")
		return
	}

	var plan RecurringIncomePlan
	if req.ID != "" {
		// 更新现有计划
		existing := GetPlanByID(req.ID)
		if existing == nil {
			response.BadRequest(c, "plan not found")
			return
		}
		plan = *existing
		plan.Type = req.Type
		plan.Label = req.Label
		plan.AccountID = req.AccountID
		plan.Amount = req.Amount
		plan.Currency = req.Currency
		plan.Frequency = req.Frequency
		plan.DayOfMonth = req.DayOfMonth
		plan.Weekday = req.Weekday
		plan.NextDueDate = req.NextDueDate
		plan.Active = req.Active
	} else {
		// 新建计划
		plan = RecurringIncomePlan{
			ID:          uuid.NewString(),
			Type:        req.Type,
			Label:       req.Label,
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Frequency:   req.Frequency,
			DayOfMonth:  req.DayOfMonth,
			Weekday:     req.Weekday,
			NextDueDate: req.NextDueDate,
			Active:      req.Active,
			CreatedAt:   time.Now().Format(time.RFC3339),
		}
	}

	saved := SavePlan(plan)
	logger.Info("Recurring income plan saved",
		zap.String("id", saved.ID),
		zap.String("type", string(saved.Type)),
		zap.Float64("amount", saved.Amount),
	)
	response.OK(c, saved)
}

// ListDuePlansHandler 查询指定日期到期的计划
func (h *Handler) ListDuePlansHandler(c *gin.Context) {
	date := c.Query("date")
	if date == "" {
		// 默认使用今天
		date = time.Now().Format("2006-01-02")
	}

	plans := FindDuePlans(date)
	logger.Info("List due recurring income plans",
		zap.String("date", date),
		zap.Int("count", len(plans)),
	)
	response.OK(c, plans)
}

// RealizePlanRequest 确认入账请求
type RealizePlanRequest struct {
	Date string `json:"date,omitempty"` // YYYY-MM-DD，可选，默认今天
}

// RealizePlanResponse 确认入账响应
type RealizePlanResponse struct {
	Plan        RecurringIncomePlan `json:"plan"`
	Transaction ledger.Transaction  `json:"transaction"`
}

// RealizePlanHandler 确认周期收入到账
func (h *Handler) RealizePlanHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	var req RealizePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 如果没有 body，使用默认值
		req.Date = ""
	}

	// 获取计划
	plan := GetPlanByID(id)
	if plan == nil {
		response.BadRequest(c, "plan not found")
		return
	}

	if !plan.Active {
		response.BadRequest(c, "plan is not active")
		return
	}

	// 确定日期
	realizeDate := req.Date
	if realizeDate == "" {
		realizeDate = time.Now().Format("2006-01-02")
	}

	// 验证日期格式
	_, err := time.Parse("2006-01-02", realizeDate)
	if err != nil {
		response.BadRequest(c, "invalid date format, expected YYYY-MM-DD")
		return
	}

	// 映射类型到分类
	var categoryKey string
	switch plan.Type {
	case RecurringIncomeTypeSalary:
		categoryKey = "INCOME_SALARY"
	case RecurringIncomeTypeRent:
		categoryKey = "INCOME_RENT"
	case RecurringIncomeTypeSide:
		categoryKey = "INCOME_SIDE"
	case RecurringIncomeTypeOther:
		categoryKey = "INCOME_OTHER"
	default:
		categoryKey = "INCOME_OTHER"
	}

	// 创建收入交易
	occurredAt, err := time.Parse("2006-01-02", realizeDate)
	if err != nil {
		response.BadRequest(c, "invalid date")
		return
	}

	// 更新账户余额
	ok := accounts.AddToBalance(plan.AccountID, plan.Amount)
	if !ok {
		response.BadRequest(c, "account not found")
		return
	}

	// 创建交易记录
	tx := ledger.Transaction{
		ID: uuid.NewString(),
		CreateTransactionRequest: ledger.CreateTransactionRequest{
			Type:        ledger.TransactionTypeIncome,
			Amount:      plan.Amount,
			CategoryKey: categoryKey,
			Remark:      plan.Label,
			OccurredAt:  occurredAt.Format(time.RFC3339),
			AccountID:   plan.AccountID,
		},
	}

	// 更新计划：设置 LastFiredDate 和计算 NextDueDate
	plan.LastFiredDate = realizeDate
	plan.NextDueDate = CalcNextDueDate(*plan)

	// 保存更新后的计划
	savedPlan := SavePlan(*plan)

	logger.Info("Recurring income realized",
		zap.String("planId", savedPlan.ID),
		zap.String("date", realizeDate),
		zap.Float64("amount", plan.Amount),
		zap.String("transactionId", tx.ID),
	)

	resp := RealizePlanResponse{
		Plan:        savedPlan,
		Transaction: tx,
	}
	response.OK(c, resp)
}
