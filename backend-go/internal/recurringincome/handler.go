package recurringincome

import (
	"encoding/json"
	"neverbroke/common/logger"
	"neverbroke/internal/http/response"
	"neverbroke/internal/ledger"
	"neverbroke/internal/models"
	"neverbroke/internal/payroll"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handler 周期收入处理器
type Handler struct {
	db *gorm.DB
}

// NewHandler 创建周期收入处理器
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// ListPlansHandler 列出所有周期收入计划
func (h *Handler) ListPlansHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var dbPlans []models.RecurringIncomePlan
	if err := h.db.Where("user_id = ?", userID.(string)).
		Order("active DESC, created_at DESC").
		Find(&dbPlans).Error; err != nil {
		logger.Error("Failed to list recurring income plans", zap.Error(err))
		response.InternalError(c, "failed to list plans")
		return
	}

	result := make([]RecurringIncomePlan, len(dbPlans))
	for i, p := range dbPlans {
		result[i] = toBusinessModel(&p)
	}

	logger.Info("List recurring income plans",
		zap.String("userID", userID.(string)),
		zap.Int("count", len(result)),
	)
	response.OK(c, result)
}

// toBusinessModel 转换数据库模型为业务模型
func toBusinessModel(p *models.RecurringIncomePlan) RecurringIncomePlan {
	plan := RecurringIncomePlan{
		ID:          p.ID,
		Type:        RecurringIncomeType(p.Type),
		Label:       p.Label,
		AccountID:   p.AccountID,
		Amount:      p.Amount,
		Currency:    p.Currency,
		Frequency:   RecurringIncomeFrequency(p.Frequency),
		NextDueDate: p.NextDueDate.Format("2006-01-02"),
		Active:      p.Active,
		CreatedAt:   p.CreatedAt.Format(time.RFC3339),
	}

	if p.DayOfMonth != nil {
		plan.DayOfMonth = *p.DayOfMonth
	}
	if p.Weekday != nil {
		plan.Weekday = *p.Weekday
	}
	if p.LastFiredDate != nil && !p.LastFiredDate.IsZero() {
		plan.LastFiredDate = p.LastFiredDate.Format("2006-01-02")
	}
	if p.CityCode != nil {
		plan.CityCode = *p.CityCode
	}
	if p.SalaryPreTaxMonthly != nil {
		plan.SalaryPreTaxMonthly = *p.SalaryPreTaxMonthly
	}
	if p.SalarySocialRate != nil {
		plan.SalarySocialRate = *p.SalarySocialRate
	}
	if p.SalaryExtraDeductions != nil {
		plan.SalaryExtraDeductions = *p.SalaryExtraDeductions
	}
	if p.SalaryPayDay != nil {
		plan.SalaryPayDay = *p.SalaryPayDay
	}
	if p.SalaryStartDate != nil && !p.SalaryStartDate.IsZero() {
		plan.SalaryStartDate = p.SalaryStartDate.Format("2006-01-02")
	}

	plan.SalaryVersions, _ = SalaryVersions(*p)
	if p.SalaryContributions != nil {
		_ = json.Unmarshal([]byte(*p.SalaryContributions), &plan.SalaryContributions)
	}
	return plan
}

// UpsertPlanHandler 创建或更新周期收入计划
func (h *Handler) UpsertPlanHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req UpsertRecurringIncomeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

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

	// 验证账户属于当前用户
	var account models.Account
	if err := h.db.Where("id = ? AND user_id = ?", req.AccountID, userID.(string)).First(&account).Error; err != nil {
		response.BadRequest(c, "account not found")
		return
	}

	nextDueDate, err := time.Parse("2006-01-02", req.NextDueDate)
	if err != nil {
		response.BadRequest(c, "invalid nextDueDate format")
		return
	}

	var dbPlan *models.RecurringIncomePlan

	if req.ID != "" {
		// 更新现有计划
		dbPlan = &models.RecurringIncomePlan{}
		if err := h.db.Where("user_id = ? AND id = ?", userID.(string), req.ID).First(dbPlan).Error; err != nil {
			response.BadRequest(c, "plan not found")
			return
		}

		dbPlan.Type = string(req.Type)
		dbPlan.Label = req.Label
		dbPlan.AccountID = req.AccountID
		dbPlan.Amount = req.Amount
		dbPlan.Currency = req.Currency
		dbPlan.Frequency = string(req.Frequency)
		dbPlan.NextDueDate = nextDueDate
		dbPlan.Active = req.Active

		if req.DayOfMonth > 0 {
			dbPlan.DayOfMonth = &req.DayOfMonth
		}
		if req.Weekday > 0 {
			dbPlan.Weekday = &req.Weekday
		}
		if req.CityCode != "" {
			dbPlan.CityCode = &req.CityCode
		}

		if err := h.db.Save(dbPlan).Error; err != nil {
			logger.Error("Failed to update plan", zap.Error(err))
			response.InternalError(c, "failed to update plan")
			return
		}
	} else {
		// 新建计划
		dbPlan = &models.RecurringIncomePlan{
			ID:          uuid.NewString(),
			UserID:      userID.(string),
			Type:        string(req.Type),
			Label:       req.Label,
			AccountID:   req.AccountID,
			Amount:      req.Amount,
			Currency:    req.Currency,
			Frequency:   string(req.Frequency),
			NextDueDate: nextDueDate,
			Active:      req.Active,
			CreatedAt:   time.Now(),
		}

		if req.DayOfMonth > 0 {
			dbPlan.DayOfMonth = &req.DayOfMonth
		}
		if req.Weekday > 0 {
			dbPlan.Weekday = &req.Weekday
		}
		if req.CityCode != "" {
			dbPlan.CityCode = &req.CityCode
		}

		if err := h.db.Create(dbPlan).Error; err != nil {
			logger.Error("Failed to create plan", zap.Error(err))
			response.InternalError(c, "failed to create plan")
			return
		}
	}

	plan := toBusinessModel(dbPlan)
	logger.Info("Recurring income plan saved",
		zap.String("userID", userID.(string)),
		zap.String("id", plan.ID),
		zap.String("type", string(plan.Type)),
		zap.Float64("amount", plan.Amount),
	)
	response.OK(c, plan)
}

// ListDuePlansHandler 查询指定日期到期的计划
func (h *Handler) ListDuePlansHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	targetDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		response.BadRequest(c, "invalid date format")
		return
	}

	var dbPlans []models.RecurringIncomePlan
	if err := h.db.Where("user_id = ? AND active = ? AND next_due_date <= ?",
		userID.(string), true, targetDate).Find(&dbPlans).Error; err != nil {
		logger.Error("Failed to list due plans", zap.Error(err))
		response.InternalError(c, "failed to list plans")
		return
	}

	result := make([]RecurringIncomePlan, len(dbPlans))
	for i, p := range dbPlans {
		result[i] = toBusinessModel(&p)
	}

	logger.Info("List due recurring income plans",
		zap.String("userID", userID.(string)),
		zap.String("date", date),
		zap.Int("count", len(result)),
	)
	response.OK(c, result)
}

// RealizePlanRequest 确认入账请求
type RealizePlanRequest struct {
	Date string `json:"date,omitempty"`
}

// RealizePlanResponse 确认入账响应
type RealizePlanResponse struct {
	Plan        RecurringIncomePlan `json:"plan"`
	Transaction ledger.Transaction  `json:"transaction"`
}

// RealizePlanHandler 确认周期收入到账
func (h *Handler) RealizePlanHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	var req RealizePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Date = ""
	}

	var dbPlan models.RecurringIncomePlan
	if err := h.db.Where("user_id = ? AND id = ?", userID.(string), id).First(&dbPlan).Error; err != nil {
		response.BadRequest(c, "plan not found")
		return
	}

	if !dbPlan.Active {
		response.BadRequest(c, "plan is not active")
		return
	}

	realizeDate := req.Date
	if realizeDate == "" {
		realizeDate = time.Now().Format("2006-01-02")
	}

	occurredAt, err := time.Parse("2006-01-02", realizeDate)
	if err != nil {
		response.BadRequest(c, "invalid date format")
		return
	}

	if dbPlan.Type == "salary" && dbPlan.SalaryPreTaxMonthly != nil {
		versions, e := SalaryVersions(dbPlan)
		if e != nil {
			response.InternalError(c, "薪资历史读取失败")
			return
		}
		items, e := PreviewSalaryHistory(versions, occurredAt.Year(), payroll.NewService(payroll.NewRepository(h.db)).GetRegionRuleAt)
		if e != nil {
			response.BadRequest(c, "薪资预估失败")
			return
		}
		matched := false
		for _, item := range items {
			if item.Date == realizeDate {
				dbPlan.Amount = item.PostTax
				dbPlan.AccountID = item.Config.AccountID
				matched = true
				break
			}
		}
		if !matched {
			response.BadRequest(c, "所选日期不是此工资方案的发薪日")
			return
		}
	}

	// 映射类型到分类
	var categoryKey string
	switch RecurringIncomeType(dbPlan.Type) {
	case RecurringIncomeTypeSalary:
		categoryKey = "INCOME_SALARY"
	case RecurringIncomeTypeRent:
		categoryKey = "INCOME_RENT"
	case RecurringIncomeTypeSide:
		categoryKey = "INCOME_SIDE"
	default:
		categoryKey = "INCOME_OTHER"
	}

	// 使用事务创建交易并更新计划
	var dbTx *models.Transaction
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 创建交易记录
		dbTx = &models.Transaction{
			ID:         uuid.NewString(),
			UserID:     userID.(string),
			Type:       "income",
			Amount:     dbPlan.Amount,
			OccurredAt: occurredAt.Format(time.RFC3339),
			AccountID:  &dbPlan.AccountID,
		}
		dbTx.CategoryKey = &categoryKey
		dbTx.Remark = &dbPlan.Label

		if err := tx.Create(dbTx).Error; err != nil {
			return err
		}

		// 更新账户余额
		if err := tx.Model(&models.Account{}).Where("id = ?", dbPlan.AccountID).
			Update("balance", gorm.Expr("balance + ?", dbPlan.Amount)).Error; err != nil {
			return err
		}

		// 更新计划
		dbPlan.LastFiredDate = &occurredAt
		plan := toBusinessModel(&dbPlan)
		nextDue := CalcNextDueDate(plan)
		if nextDue != "" {
			nextDueDate, _ := time.Parse("2006-01-02", nextDue)
			dbPlan.NextDueDate = nextDueDate
		}

		return tx.Save(&dbPlan).Error
	})

	if err != nil {
		logger.Error("Failed to realize plan", zap.Error(err))
		response.InternalError(c, "failed to realize plan")
		return
	}

	plan := toBusinessModel(&dbPlan)
	txResp := ledger.Transaction{
		ID: dbTx.ID,
		CreateTransactionRequest: ledger.CreateTransactionRequest{
			Type:        ledger.TransactionTypeIncome,
			Amount:      dbPlan.Amount,
			CategoryKey: categoryKey,
			Remark:      dbPlan.Label,
			OccurredAt:  occurredAt.Format(time.RFC3339),
			AccountID:   dbPlan.AccountID,
		},
	}

	logger.Info("Recurring income realized",
		zap.String("userID", userID.(string)),
		zap.String("planId", plan.ID),
		zap.String("date", realizeDate),
		zap.Float64("amount", plan.Amount),
	)

	response.OK(c, RealizePlanResponse{
		Plan:        plan,
		Transaction: txResp,
	})
}

// SalaryPreviewHandler 工资预览接口
func SalaryPreviewHandler(c *gin.Context) {
	var req SalaryPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.Currency == "" {
		req.Currency = "CNY"
	}

	resp, err := PreviewSalaryToYearEnd(req)
	if err != nil {
		logger.Warn("Preview salary failed", zap.Error(err))
		response.BadRequest(c, "preview salary failed")
		return
	}

	response.OK(c, resp)
}

// ProjectedTransactionsHandler 未来明细投射接口
func (h *Handler) ProjectedTransactionsHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	from := c.Query("from")
	to := c.Query("to")

	if from == "" || to == "" {
		response.BadRequest(c, "from and to parameters are required")
		return
	}

	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		response.BadRequest(c, "invalid from date format")
		return
	}

	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		response.BadRequest(c, "invalid to date format")
		return
	}

	if toDate.Before(fromDate) {
		response.BadRequest(c, "to date must be after from date")
		return
	}

	var dbPlans []models.RecurringIncomePlan
	if err := h.db.Where("user_id = ? AND active = ?", userID.(string), true).Find(&dbPlans).Error; err != nil {
		logger.Error("Failed to list plans", zap.Error(err))
		response.InternalError(c, "failed to list plans")
		return
	}

	var projected []ProjectedTransaction

	for _, dbPlan := range dbPlans {
		plan := toBusinessModel(&dbPlan)
		if dbPlan.Type == "salary" && dbPlan.SalaryPreTaxMonthly != nil {
			versions, err := SalaryVersions(dbPlan)
			if err != nil {
				response.InternalError(c, "薪资历史读取失败")
				return
			}
			service := payroll.NewService(payroll.NewRepository(h.db))
			for year := fromDate.Year(); year <= toDate.Year(); year++ {
				items, err := PreviewSalaryHistory(versions, year, service.GetRegionRuleAt)
				if err != nil {
					response.BadRequest(c, "薪资预估失败")
					return
				}
				for _, item := range items {
					if item.Date < from || item.Date > to {
						continue
					}
					projected = append(projected, ProjectedTransaction{Date: item.Date, AccountID: item.Config.AccountID, Amount: item.PostTax, Type: "income", Label: plan.Label, PlanID: plan.ID, IsSalary: true})
				}
			}
			continue
		}
		dates := generateIncomeDates(plan, fromDate, toDate)
		isSalary := plan.Type == RecurringIncomeTypeSalary

		for _, date := range dates {
			amount := plan.Amount

			if isSalary && plan.CityCode != "" {
				socialRate := GetSocialSecurityRate(plan.CityCode)
				socialSecurity := amount * socialRate
				taxableIncome := amount - socialSecurity - MonthlyDeduction
				if taxableIncome > 0 {
					tax := CalculateTax(taxableIncome)
					amount = amount - socialSecurity - tax
				} else {
					amount = amount - socialSecurity
				}
			}

			projected = append(projected, ProjectedTransaction{
				Date:      date.Format("2006-01-02"),
				AccountID: plan.AccountID,
				Amount:    amount,
				Type:      "income",
				Label:     plan.Label,
				PlanID:    plan.ID,
				IsSalary:  isSalary,
			})
		}
	}

	if projected == nil {
		projected = []ProjectedTransaction{}
	}

	logger.Info("List projected transactions",
		zap.String("userID", userID.(string)),
		zap.String("from", from),
		zap.String("to", to),
		zap.Int("count", len(projected)),
	)

	response.OK(c, projected)
}

// GetSalaryPlanHandler 获取当前用户的工资方案
func (h *Handler) GetSalaryPlanHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var dbPlan models.RecurringIncomePlan
	err := h.db.Where("user_id = ? AND type = ? AND active = ?",
		userID.(string), string(RecurringIncomeTypeSalary), true).
		Order("created_at DESC").
		First(&dbPlan).Error

	var salaryPlan *RecurringIncomePlan
	if err == nil {
		plan := toBusinessModel(&dbPlan)
		salaryPlan = &plan
	} else if err != gorm.ErrRecordNotFound {
		response.InternalError(c, "读取薪资方案失败")
		return
	}

	logger.Info("Get salary plan",
		zap.String("userID", userID.(string)),
		zap.Bool("found", salaryPlan != nil),
	)

	response.OK(c, GetSalaryPlanResponse{
		Plan: salaryPlan,
	})
}

// UpsertSalaryPlanHandler 保存/更新工资方案
func (h *Handler) UpsertSalaryPlanHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	var req UpsertSalaryPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body", zap.Error(err))
		response.BadRequest(c, "invalid request")
		return
	}

	if req.PreTaxMonthly <= 0 {
		response.BadRequest(c, "preTaxMonthly must be > 0")
		return
	}
	if req.AccountID == "" {
		response.BadRequest(c, "accountId is required")
		return
	}
	if req.StartDate == "" {
		response.BadRequest(c, "startDate is required")
		return
	}
	if req.PayDay < 1 || req.PayDay > 31 {
		response.BadRequest(c, "payDay must be between 1 and 31")
		return
	}

	// 验证账户
	var account models.Account
	if err := h.db.Where("id = ? AND user_id = ?", req.AccountID, userID.(string)).First(&account).Error; err != nil {
		response.BadRequest(c, "account not found")
		return
	}

	if req.Currency == "" {
		req.Currency = "CNY"
	}
	startDate, err := time.Parse(time.DateOnly, req.StartDate)
	if err != nil {
		response.BadRequest(c, "生效日期无效")
		return
	}
	if req.BaseMode == "policy" && startDate.Day() != 1 {
		response.BadRequest(c, "自动政策方案请从月份首日生效")
		return
	}
	// Validate all values before attempting a transactional save.
	if _, err = MergeSalaryVersion(nil, req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	preview, err := PreviewSalaryToYearEnd(SalaryPreviewRequest{PreTaxAmount: req.PreTaxMonthly, CityCode: req.CityCode, StartDate: req.StartDate, PayDay: req.PayDay, Frequency: FrequencyMonthly, Currency: req.Currency, SocialRate: &req.SocialRate, ExtraDeductions: req.ExtraDeductions, Contributions: req.Contributions})
	if err != nil || len(preview.Items) == 0 {
		response.BadRequest(c, "无效薪资或缴费配置")
		return
	}
	postTaxAmount := preview.Items[0].PostTax
	nextDueDate, _ := time.Parse(time.DateOnly, preview.Items[0].Date)

	// 创建新方案
	dbPlan := &models.RecurringIncomePlan{
		ID:                    uuid.NewString(),
		UserID:                userID.(string),
		Type:                  string(RecurringIncomeTypeSalary),
		Label:                 "工资",
		AccountID:             req.AccountID,
		Amount:                postTaxAmount,
		Currency:              req.Currency,
		Frequency:             string(FrequencyMonthly),
		DayOfMonth:            &req.PayDay,
		NextDueDate:           nextDueDate,
		Active:                true,
		CreatedAt:             time.Now(),
		SalaryPreTaxMonthly:   &req.PreTaxMonthly,
		SalarySocialRate:      &req.SocialRate,
		SalaryExtraDeductions: &req.ExtraDeductions,
		SalaryPayDay:          &req.PayDay,
		SalaryStartDate:       &startDate,
	}

	if req.Contributions != nil {
		raw, _ := json.Marshal(req.Contributions)
		value := string(raw)
		dbPlan.SalaryContributions = &value
	}
	if req.CityCode != "" {
		dbPlan.CityCode = &req.CityCode
	}

	// Serialize per-user writes so concurrent raises cannot discard each other's history.
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", "salary:"+userID.(string)).Error; err != nil {
			return err
		}
		var old models.RecurringIncomePlan
		var versions []UpsertSalaryPlanRequest
		found := tx.Where("user_id = ? AND type = ? AND active = ?", userID.(string), "salary", true).Order("created_at DESC").First(&old).Error
		if found == nil {
			var e error
			versions, e = SalaryVersions(old)
			if e != nil {
				return e
			}
		} else if found != gorm.ErrRecordNotFound {
			return found
		}
		versions, e := MergeSalaryVersion(versions, req)
		if e != nil {
			return e
		}
		service := payroll.NewService(payroll.NewRepository(tx))
		previewYear := startDate.AddDate(0, req.PayMonthOffset, 0).Year()
		items, e := PreviewSalaryHistory(versions, previewYear, service.GetRegionRuleAt)
		if e != nil {
			return e
		}
		preview = HistoryPreview(items, previewYear)
		for _, item := range items {
			if item.Config.StartDate == req.StartDate {
				dbPlan.Amount = item.PostTax
				dbPlan.NextDueDate, _ = time.Parse(time.DateOnly, item.Date)
				break
			}
		}
		raw, e := json.Marshal(versions)
		if e != nil {
			return e
		}
		value := string(raw)
		dbPlan.SalaryVersions = &value
		if c.Query("preview") == "true" {
			return nil
		}
		if e := tx.Model(&models.RecurringIncomePlan{}).Where("user_id = ? AND type = ? AND active = ?", userID.(string), "salary", true).Update("active", false).Error; e != nil {
			return e
		}
		return tx.Create(dbPlan).Error
	})
	if err != nil {
		logger.Warn("Save salary history failed", zap.Error(err))
		response.BadRequest(c, "保存失败："+err.Error())
		return
	}

	plan := toBusinessModel(dbPlan)
	logger.Info("Salary plan saved",
		zap.String("userID", userID.(string)),
		zap.String("id", plan.ID),
		zap.Float64("preTax", req.PreTaxMonthly),
		zap.Float64("postTax", postTaxAmount),
	)

	response.OK(c, UpsertSalaryPlanResponse{
		Plan:    plan,
		Preview: preview,
	})
}

// DeleteSalaryPlanHandler 删除工资方案
func (h *Handler) DeleteSalaryPlanHandler(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "user not authenticated")
		return
	}

	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "id is required")
		return
	}

	var dbPlan models.RecurringIncomePlan
	if err := h.db.Where("user_id = ? AND id = ?", userID.(string), id).First(&dbPlan).Error; err != nil {
		response.BadRequest(c, "plan not found")
		return
	}

	if RecurringIncomeType(dbPlan.Type) != RecurringIncomeTypeSalary {
		response.BadRequest(c, "not a salary plan")
		return
	}

	dbPlan.Active = false
	if err := h.db.Save(&dbPlan).Error; err != nil {
		logger.Error("Failed to delete salary plan", zap.Error(err))
		response.InternalError(c, "failed to delete salary plan")
		return
	}

	logger.Info("Salary plan deleted",
		zap.String("userID", userID.(string)),
		zap.String("id", id),
	)

	response.OK[any](c, nil)
}

// generateIncomeDates 生成计划在指定时间范围内的所有收入日期
func generateIncomeDates(plan RecurringIncomePlan, fromDate, toDate time.Time) []time.Time {
	var dates []time.Time

	nextDue, err := time.Parse("2006-01-02", plan.NextDueDate)
	if err != nil {
		return dates
	}

	currentDate := nextDue
	if currentDate.Before(fromDate) {
		for currentDate.Before(fromDate) {
			currentDate = calcNextDate(currentDate, plan.Frequency, plan.DayOfMonth)
		}
	}

	for !currentDate.After(toDate) {
		dates = append(dates, currentDate)
		currentDate = calcNextDate(currentDate, plan.Frequency, plan.DayOfMonth)
	}

	return dates
}

// calcNextDate 计算下一个收入日期
func calcNextDate(current time.Time, frequency RecurringIncomeFrequency, dayOfMonth int) time.Time {
	switch frequency {
	case FrequencyMonthly:
		return AddMonths(current, 1)
	case FrequencyBimonthly:
		return AddMonths(current, 2)
	case FrequencyQuarterly:
		return AddMonths(current, 3)
	case FrequencySemiannual:
		return AddMonths(current, 6)
	case FrequencyWeekly:
		return current.AddDate(0, 0, 7)
	case FrequencyYearly:
		return current.AddDate(1, 0, 0)
	default:
		return AddMonths(current, 1)
	}
}
