#!/bin/bash

# 集成测试数据库设置脚本
# 用途：创建和配置集成测试所需的 PostgreSQL 数据库

set -e

echo "🔧 Setting up integration test database..."

# 配置变量
DB_NAME="${NEVERBROKE_TEST_DB:-neverbroke_test}"
DB_USER="${NEVERBROKE_PG_USER:-neverbroke}"
DB_PASSWORD="${NEVERBROKE_PG_PASSWORD:-neverbroke}"
DB_HOST="${NEVERBROKE_PG_HOST:-127.0.0.1}"
DB_PORT="${NEVERBROKE_PG_PORT:-5432}"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查 PostgreSQL 是否安装
if ! command -v psql &> /dev/null; then
    echo -e "${RED}❌ PostgreSQL is not installed${NC}"
    echo "Please install PostgreSQL first:"
    echo "  macOS: brew install postgresql@15"
    echo "  Ubuntu: sudo apt-get install postgresql-15"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is installed${NC}"

# 检查 PostgreSQL 是否运行
if ! pg_isready -h $DB_HOST -p $DB_PORT &> /dev/null; then
    echo -e "${RED}❌ PostgreSQL is not running${NC}"
    echo "Please start PostgreSQL first:"
    echo "  macOS: brew services start postgresql@15"
    echo "  Ubuntu: sudo systemctl start postgresql"
    exit 1
fi

echo -e "${GREEN}✅ PostgreSQL is running${NC}"

# 创建数据库和用户（如果不存在）
echo "Creating database and user..."

psql -h $DB_HOST -p $DB_PORT -U postgres <<EOF
-- 创建用户（如果不存在）
DO \$\$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '$DB_USER') THEN
        CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
    END IF;
END
\$\$;

-- 删除已存在的测试数据库（如果存在）
DROP DATABASE IF EXISTS $DB_NAME;

-- 创建测试数据库
CREATE DATABASE $DB_NAME OWNER $DB_USER;

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Database '$DB_NAME' created successfully${NC}"
else
    echo -e "${RED}❌ Failed to create database${NC}"
    echo -e "${YELLOW}Tip: You may need to run this script with sudo or as postgres user${NC}"
    exit 1
fi

# 连接到测试数据库并设置权限
psql -h $DB_HOST -p $DB_PORT -U postgres -d $DB_NAME <<EOF
-- 授予 schema 权限
GRANT ALL ON SCHEMA public TO $DB_USER;

-- 授予默认权限
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $DB_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $DB_USER;
EOF

echo -e "${GREEN}✅ Database permissions configured${NC}"

# 验证连接
echo "Verifying database connection..."
if psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();" &> /dev/null; then
    echo -e "${GREEN}✅ Database connection verified${NC}"
else
    echo -e "${RED}❌ Failed to connect to database${NC}"
    exit 1
fi

# 创建 .env.test 文件
ENV_FILE="../../.env.test"
echo "Creating $ENV_FILE..."

cat > $ENV_FILE <<EOF
# Integration Test Environment Variables
NEVERBROKE_TEST_DB=$DB_NAME
NEVERBROKE_PG_HOST=$DB_HOST
NEVERBROKE_PG_PORT=$DB_PORT
NEVERBROKE_PG_USER=$DB_USER
NEVERBROKE_PG_PASSWORD=$DB_PASSWORD
JWT_SECRET=test-secret-key-for-integration-tests-min-32-chars
EOF

echo -e "${GREEN}✅ Environment file created: $ENV_FILE${NC}"

# 显示连接信息
echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✨ Test database setup completed!${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "Database Information:"
echo "  Host:     $DB_HOST"
echo "  Port:     $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User:     $DB_USER"
echo ""
echo "Connection String:"
echo "  postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable"
echo ""
echo "To connect manually:"
echo "  psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME"
echo ""
echo "To run integration tests:"
echo "  cd backend-go"
echo "  source .env.test"
echo "  go test ./tests/integration/... -v"
echo ""
echo -e "${YELLOW}Note: Tables will be auto-migrated when tests run${NC}"
echo ""
