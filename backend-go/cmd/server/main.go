package main

import (
	"neverbroke/common/logger"
	"neverbroke/internal/citysocialrate"
	"neverbroke/internal/http/router"
	"neverbroke/internal/storage/memory"

	"go.uber.org/zap"
)

func main() {
	// 初始化日志 - 同时输出到控制台和文件
	if err := logger.Init(logger.Config{
		Level: "info", // 可配置：debug, info, warn, error
		OutputPaths: []string{
			"stdout",         // 输出到控制台
			"./logs/app.log", // 同时输出到文件
		},
		Format: "console", // 可配置：console 或 json
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("🚀 NeverBroke API Server starting...",
		zap.String("version", "v0.1.0"),
		zap.String("port", "8080"),
	)

	// 初始化城市社保比例数据
	citysocialrate.InitDefaultRates()
	logger.Info("City social rates initialized")

	// 创建内存存储
	store := memory.NewStore()
	logger.Info("Memory store initialized")

	// 设置路由
	r := router.Setup(store)
	logger.Info("Router setup completed")

	// 启动服务
	logger.Info("Server listening on :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
