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

	// 创建所有 handlers（复用于 /api 和 /api/v1 路由组）
	homeHandler := home.NewHandler(store)
	accountsHandler := accounts.NewHandler(db)
	ledgerHandler := ledger.NewHandler(store.Transactions(), db)
	reportHandler := report.NewHandler(store)
	memoryStore := memory.NewStore()
	userHandler := user.NewHandler(memoryStore)
	budgetHandler := budget.NewHandler()
	importerHandler := importer.NewHandler()
	incomePlanHandler := incomeplan.NewHandler()
	installmentHandler := installment.NewHandler()
	recurringIncomeHandler := recurringincome.NewHandler()
	citySocialRateHandler := citysocialrate.NewHandler()

	// 注册路由函数（DRY - Don't Repeat Yourself）
	registerRoutes := func(group *gin.RouterGroup) {
		// 首页 - 完整概览（保留用于兼容）
		group.GET("/home/overview", homeHandler.GetOverview)

		// 首页 - 独立 API
		group.GET("/home/asset-summary", homeHandler.GetAssetSummary)   // 资产总览
		group.GET("/home/budget-stats", homeHandler.GetBudgetStats)     // 本月预算
		group.GET("/home/month-overview", homeHandler.GetMonthOverview) // 本月概览
		group.GET("/home/category-stats", homeHandler.GetCategoryStats) // 分类统计
		group.GET("/home/pay-cycle", homeHandler.GetPayCycle)           // 支付周期（保留用于兼容）

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

		// 薪资控制（我的 -> 薪资控制）
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

	// API 路由组（不带版本号，与前端兼容）
	// 需要认证的路由
	api := r.Group("/api", tokenHandler.AuthMiddleware())
	registerRoutes(api)

	// API v1 路由组（保留用于未来版本管理）
	v1 := r.Group("/api/v1", tokenHandler.AuthMiddleware())
	registerRoutes(v1)

	return r
}
