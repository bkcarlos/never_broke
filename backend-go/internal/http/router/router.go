package router

import (
	"time"

	"neverbroke/common/logger"
	"neverbroke/internal/accounts"
	"neverbroke/internal/auth"
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
	"neverbroke/internal/repository"
	"neverbroke/internal/timeline"
	"neverbroke/internal/user"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Setup 设置路由
func Setup(store repository.Store, db *gorm.DB) *gin.Engine {
	// 创建不带默认中间件的 Gin 引擎
	r := gin.New()

	// 使用自定义的日志中间件
	r.Use(logger.GinLogger())
	r.Use(logger.GinRecovery())

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// 初始化认证系统（Token 方案）
	tokenMgr := auth.NewTokenManager()
	userRepo := auth.NewGormRepository(db)
	tokenRepo := auth.NewGormTokenRepository(db)
	tokenService := auth.NewTokenService(userRepo, tokenRepo, tokenMgr)
	tokenHandler := auth.NewTokenHandler(tokenService)

	// OAuth 服务
	oauthService := auth.NewOAuthService(userRepo, tokenRepo, tokenMgr)
	oauthHandler := auth.NewOAuthHandler(oauthService)

	// 启动定期清理协程
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := tokenService.CleanupExpiredData(); err != nil {
				logger.Error("Failed to cleanup expired data", zap.Error(err))
			}
		}
	}()

	// 认证路由（无需鉴权）
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/register", tokenHandler.Register)
		authGroup.POST("/login", tokenHandler.Login)
		authGroup.POST("/refresh", tokenHandler.Refresh)
		authGroup.POST("/logout", tokenHandler.AuthMiddleware(), tokenHandler.Logout)
		authGroup.POST("/logout_all", tokenHandler.AuthMiddleware(), tokenHandler.LogoutAll)
	}

	// OAuth 路由
	oauthGroup := r.Group("/api/oauth")
	{
		oauthGroup.GET("/:provider/start", oauthHandler.Start)
		oauthGroup.GET("/:provider/callback", oauthHandler.Callback)
	}

	// 用户信息路由（需要鉴权）
	r.GET("/api/me", tokenHandler.AuthMiddleware(), tokenHandler.Me)

	// 创建所有 handlers
	homeHandler := home.NewHandler(store, db)
	accountsHandler := accounts.NewHandler(db)
	ledgerHandler := ledger.NewHandler(store.Transactions(), db)
	reportHandler := report.NewHandler(store)
	userHandler := user.NewHandler(userRepo)

	// 直接使用 db 的 handlers（避免循环导入）
	budgetHandler := budget.NewHandler(db)
	incomePlanHandler := incomeplan.NewHandler(db)
	installmentHandler := installment.NewHandler(db)
	recurringIncomeHandler := recurringincome.NewHandler(db)

	importerHandler := importer.NewHandler()
	citySocialRateHandler := citysocialrate.NewHandler()

	// 注册路由函数
	registerRoutes := func(group *gin.RouterGroup) {
		// 首页
		group.GET("/home/overview", homeHandler.GetOverview)
		group.GET("/home/asset-summary", homeHandler.GetAssetSummary)
		group.GET("/home/budget-stats", homeHandler.GetBudgetStats)
		group.GET("/home/month-overview", homeHandler.GetMonthOverview)
		group.GET("/home/category-stats", homeHandler.GetCategoryStats)
		group.GET("/home/pay-cycle", homeHandler.GetPayCycle)

		// 账户
		group.GET("/accounts", accountsHandler.ListAccountsHandler)
		group.POST("/accounts", accountsHandler.CreateAccountHandler)
		group.PATCH("/accounts/:id/credit-limits", accountsHandler.UpdateCreditLimitsHandler)

		// 账本明细
		group.GET("/transactions/daily", ledgerHandler.GetDailyTransactions)
		group.POST("/transactions", ledgerHandler.CreateTransaction)

		// 报表
		group.GET("/reports/monthly", reportHandler.GetMonthlyReport)

		// 用户设置
		group.GET("/user/settings", userHandler.GetSettings)
		group.PATCH("/user/settings", userHandler.UpdateSettings)

		// 预算
		group.GET("/budget/monthly", budgetHandler.GetMonthlyBudgetHandler)
		group.PUT("/budget/monthly", budgetHandler.PutMonthlyBudgetHandler)

		// 导入
		group.POST("/import/preview", importerHandler.PreviewHandler)

		// 收入计划
		group.GET("/income-plans", incomePlanHandler.ListIncomePlansHandler)
		group.POST("/income-plans", incomePlanHandler.CreateIncomePlanHandler)

		// 分期
		group.GET("/installments", installmentHandler.ListInstallmentPlansHandler)
		group.POST("/installments/preview", installmentHandler.PreviewInstallmentHandler)
		group.POST("/installments/expense", installmentHandler.CreateExpenseInstallmentHandler)

		// 周期收入配置
		group.GET("/recurring-incomes", recurringIncomeHandler.ListPlansHandler)
		group.POST("/recurring-incomes", recurringIncomeHandler.UpsertPlanHandler)
		group.GET("/recurring-incomes/due", recurringIncomeHandler.ListDuePlansHandler)
		group.POST("/recurring-incomes/:id/realize", recurringIncomeHandler.RealizePlanHandler)
		group.POST("/recurring-incomes/salary/preview", recurringincome.SalaryPreviewHandler)
		group.GET("/recurring-incomes/projected-transactions", recurringIncomeHandler.ProjectedTransactionsHandler)

		// 城市社保比例
		group.GET("/city-social-rates", citySocialRateHandler.ListCitySocialRatesHandler)
		group.GET("/city-social-rates/:cityCode", citySocialRateHandler.GetCitySocialRateDetailHandler)

		// 薪资管理
		group.GET("/me/salary-plan", recurringIncomeHandler.GetSalaryPlanHandler)
		group.POST("/me/salary-plan", recurringIncomeHandler.UpsertSalaryPlanHandler)
		group.DELETE("/me/salary-plan/:id", recurringIncomeHandler.DeleteSalaryPlanHandler)

		// 现金流预测
		group.GET("/cashflow/forecast", cashflow.ForecastHandler(store))

		// 报表分析
		group.GET("/reports/monthly-summary", reports.MonthlySummaryHandler(store))

		// 时间线
		group.GET("/timeline/daily-summary", timeline.DailySummaryHandler(store))
		group.GET("/timeline/daily-detail", timeline.DailyDetailHandler(store))
	}

	// API 路由组（需要认证）
	api := r.Group("/api", tokenHandler.AuthMiddleware())
	registerRoutes(api)

	// API v1 路由组
	v1 := r.Group("/api/v1", tokenHandler.AuthMiddleware())
	registerRoutes(v1)

	return r
}
