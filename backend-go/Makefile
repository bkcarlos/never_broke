.PHONY: migrate migrate-create migrate-gorm db-reset help

# 运行数据库迁移 (原生 SQL)
migrate:
	@echo "Running database migrations (SQL)..."
	@cd cmd/migrate && go run main.go

# 运行 GORM 自动迁移
migrate-gorm:
	@echo "Running GORM auto migrations..."
	@cd cmd/migrate-gorm && go run main.go

# 创建新的迁移文件
migrate-create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%Y%m%d%H%M%S); \
	filename="migrations/$${timestamp}_$${name}.sql"; \
	touch $$filename; \
	echo "Created migration file: $$filename"

# 重置数据库（危险操作！）
db-reset:
	@echo "⚠️  This will drop and recreate the database!"
	@read -p "Are you sure? (yes/no): " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		PGPASSWORD=app_password psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS neverbroke;"; \
		PGPASSWORD=app_password psql -h localhost -U postgres -c "CREATE DATABASE neverbroke;"; \
		echo "✅ Database reset complete. Run 'make migrate' to apply migrations."; \
	else \
		echo "Cancelled."; \
	fi

# 显示帮助信息
help:
	@echo "Available commands:"
	@echo "  make migrate        - Run database migrations (SQL)"
	@echo "  make migrate-gorm   - Run GORM auto migrations"
	@echo "  make migrate-create - Create a new migration file"
	@echo "  make db-reset       - Drop and recreate database (dangerous!)"
	@echo "  make help           - Show this help message"
