package router

import (
	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/budget"
	"neverbroke/internal/home"
	"neverbroke/internal/importer"
	"neverbroke/internal/incomeplan"
	"neverbroke/internal/installment"
	"neverbroke/internal/ledger"
	"neverbroke/internal/report"
	"neverbroke/internal/storage/memory"
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
		v1.POST("/installments/expense", installmentHandler.CreateExpenseInstallmentHandler)
	}

	return r
}
