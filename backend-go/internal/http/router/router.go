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
	"neverbroke/internal/storage/memory"
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
	// 安全说明：
	// 1. JWT 密钥从环境变量读取（JWT_SECRET）
	// 2. 生产环境必须设置 JWT_SECRET 环境变量（至少 32 字节）
	// 3. Refresh Token 哈希存储，禁止明文
	// 4. 支持 Token Rotation 和重放检测
	tokenMgr := auth.NewTokenManager()
	userRepo := auth.NewGormRepository(db)
	tokenRepo := auth.NewGormTokenRepository(db)
	tokenService := auth.NewTokenService(userRepo, tokenRepo, tokenMgr)
	tokenHandler := auth.NewTokenHandler(tokenService)

	// OAuth 服务
	oauthService := auth.NewOAuthService(userRepo, tokenRepo, tokenMgr)
	oauthHandler := auth.NewOAuthHandler(oauthService)

	// 启动定期清理协程（每小时清理一次过期数据）
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

	// API v1 路由组
	v1 := r.Group("/api/v1")
	{
		// 首页
		homeHandler := home.NewHandler(store)
		v1.GET("/home/overview", homeHandler.GetOverview)

		// 账户
		accountsHandler := accounts.NewHandler()
		v1.GET("/accounts", accountsHandler.ListAccountsHandler)
		v1.POST("/accounts", accountsHandler.CreateAccountHandler)
		v1.PATCH("/accounts/:id/credit-limits", accountsHandler.UpdateCreditLimitsHandler)

		// 账本明细
		ledgerHandler := ledger.NewHandler(store.Transactions())
		v1.GET("/transactions/daily", ledgerHandler.GetDailyTransactions)
		v1.POST("/transactions", ledgerHandler.CreateTransaction)

		// 报表
		reportHandler := report.NewHandler(store)
		v1.GET("/reports/monthly", reportHandler.GetMonthlyReport)

		// 用户设置 (使用内存存储)
		memoryStore := memory.NewStore()
		userHandler := user.NewHandler(memoryStore)
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
		v1.GET("/cashflow/forecast", cashflow.ForecastHandler(store))

		// 报表分析
		v1.GET("/reports/monthly-summary", reports.MonthlySummaryHandler(store))

		// 时间线
		v1.GET("/timeline/daily-summary", timeline.DailySummaryHandler(store))
		v1.GET("/timeline/daily-detail", timeline.DailyDetailHandler(store))
	}

	return r
}
