package router

import (
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/budget"
	"neverbroke/internal/cashflow"
	"neverbroke/internal/citysocialrate"
	"neverbroke/internal/home"
	"neverbroke/internal/importer"
	"neverbroke/internal/incomeplan"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/recurringincome"
	"neverbroke/internal/report"
	"neverbroke/internal/reports"
	"neverbroke/internal/repository/memoryadapter"
	"neverbroke/internal/storage/memory"
	"neverbroke/internal/timeline"
	"neverbroke/internal/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Setup 设置路由
func Setup(store *memory.Store) *gin.Engine {
	// 创建不带默认中间件的 Gin 引擎
	r := gin.New()

	// 使用自定义的日志中间件
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 首页
		homeHandler := home.NewHandler()
		v1.GET("/home/overview", homeHandler.GetOverview)

		// 账户
		accountsHandler := accounts.NewHandler()
		v1.GET("/accounts", accountsHandler.ListAccountsHandler)
		v1.POST("/accounts", accountsHandler.CreateAccountHandler)
		v1.PATCH("/accounts/:id/credit-limits", accountsHandler.UpdateCreditLimitsHandler)

		// 账本明细
		ledgerHandler := ledger.NewHandler()
		v1.GET("/transactions/daily", ledgerHandler.GetDailyTransactions)
		v1.POST("/transactions", ledgerHandler.CreateTransaction)

		// 报表
		reportHandler := report.NewHandler()
		v1.GET("/reports/monthly", reportHandler.GetMonthlyReport)

		// 用户设置
		userHandler := user.NewHandler(store)
		v1.GET("/user/settings", userHandler.GetSettings)
		v1.PATCH("/user/settings", userHandler.UpdateSettings)

		// 预算
		budgetHandler := budget.NewHandler()
		v1.GET("/budget/monthly", budgetHandler.GetMonthlyBudgetHandler)
		v1.PUT("/budget/monthly", budgetHandler.PutMonthlyBudgetHandler)

		// 导入
		importerHandler := importer.NewHandler()
		v1.POST("/import/preview", importerHandler.PreviewHandler)

		// 收入计划
		incomePlanHandler := incomeplan.NewHandler()
		v1.GET("/income-plans", incomePlanHandler.ListIncomePlansHandler)
		v1.POST("/income-plans", incomePlanHandler.CreateIncomePlanHandler)

		// 分期
		installmentHandler := installment.NewHandler()
		v1.POST("/installments/preview", installmentHandler.PreviewInstallmentHandler)
		v1.POST("/installments/expense", installmentHandler.CreateExpenseInstallmentHandler)

		// 周期收入配置
		recurringIncomeHandler := recurringincome.NewHandler()
		v1.GET("/recurring-incomes", recurringIncomeHandler.ListPlansHandler)
		v1.POST("/recurring-incomes", recurringIncomeHandler.UpsertPlanHandler)
		v1.GET("/recurring-incomes/due", recurringIncomeHandler.ListDuePlansHandler)
		v1.POST("/recurring-incomes/:id/realize", recurringIncomeHandler.RealizePlanHandler)
		v1.POST("/recurring-incomes/salary/preview", recurringincome.SalaryPreviewHandler)
		v1.GET("/recurring-incomes/projected-transactions", recurringIncomeHandler.ProjectedTransactionsHandler)

		// 城市社保比例
		citySocialRateHandler := citysocialrate.NewHandler()
		v1.GET("/city-social-rates", citySocialRateHandler.ListCitySocialRatesHandler)
		v1.GET("/city-social-rates/:cityCode", citySocialRateHandler.GetCitySocialRateDetailHandler)

		// 薪资控制（我的 -> 薪资控制）
		v1.GET("/me/salary-plan", recurringIncomeHandler.GetSalaryPlanHandler)
		v1.POST("/me/salary-plan", recurringIncomeHandler.UpsertSalaryPlanHandler)
		v1.DELETE("/me/salary-plan/:id", recurringIncomeHandler.DeleteSalaryPlanHandler)

		// 现金流预测
		memoryStoreAdapter := memoryadapter.NewMemoryStoreAdapter()
		v1.GET("/cashflow/forecast", cashflow.ForecastHandler(memoryStoreAdapter))

		// 报表分析
		v1.GET("/reports/monthly-summary", reports.MonthlySummaryHandler(memoryStoreAdapter))

		// 时间线
		v1.GET("/timeline/daily-summary", timeline.DailySummaryHandler(memoryStoreAdapter))
		v1.GET("/timeline/daily-detail", timeline.DailyDetailHandler(memoryStoreAdapter))
	}

	return r
}
