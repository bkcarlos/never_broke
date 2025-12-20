#!/bin/bash

# 集成测试运行脚本
# 用途：便捷地运行集成测试并生成报告

set -e

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

# 默认参数
VERBOSE=false
RACE=false
COVERAGE=false
SPECIFIC_TEST=""

# 解析命令行参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -v|--verbose)
            VERBOSE=true
            shift
            ;;
        -r|--race)
            RACE=true
            shift
            ;;
        -c|--coverage)
            COVERAGE=true
            shift
            ;;
        -t|--test)
            SPECIFIC_TEST="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -v, --verbose    Enable verbose output"
            echo "  -r, --race       Enable race detector"
            echo "  -c, --coverage   Generate coverage report"
            echo "  -t, --test NAME  Run specific test pattern"
            echo "  -h, --help       Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                           # Run all tests"
            echo "  $0 -v -r                     # Run with verbose and race detection"
            echo "  $0 -c                        # Run with coverage report"
            echo "  $0 -t TestAuthFlow           # Run only auth tests"
            echo "  $0 -t TestTransactionFlow    # Run only transaction tests"
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            echo "Use -h or --help for usage information"
            exit 1
            ;;
    esac
done

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}🧪 NeverBroke Integration Tests${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 检查是否在正确的目录
if [ ! -f "setup_test.go" ]; then
    echo -e "${RED}❌ Error: Please run this script from the tests/integration directory${NC}"
    exit 1
fi

# 加载环境变量
if [ -f "../../.env.test" ]; then
    echo -e "${GREEN}📝 Loading environment variables from .env.test${NC}"
    export $(cat ../../.env.test | grep -v '^#' | xargs)
else
    echo -e "${YELLOW}⚠️  Warning: .env.test not found, using default values${NC}"
fi

# 检查数据库连接
echo -e "${BLUE}🔍 Checking database connection...${NC}"
DB_HOST="${NEVERBROKE_PG_HOST:-127.0.0.1}"
DB_PORT="${NEVERBROKE_PG_PORT:-5432}"
DB_NAME="${NEVERBROKE_TEST_DB:-neverbroke_test}"
DB_USER="${NEVERBROKE_PG_USER:-neverbroke}"

if ! pg_isready -h $DB_HOST -p $DB_PORT &> /dev/null; then
    echo -e "${RED}❌ PostgreSQL is not running${NC}"
    echo -e "${YELLOW}Please start PostgreSQL or run: ./setup_test_db.sh${NC}"
    exit 1
fi

# 检查测试数据库是否存在
if ! psql -h $DB_HOST -p $DB_PORT -U postgres -lqt | cut -d \| -f 1 | grep -qw $DB_NAME; then
    echo -e "${YELLOW}⚠️  Test database not found${NC}"
    echo -e "${BLUE}Creating test database...${NC}"
    ./setup_test_db.sh
fi

echo -e "${GREEN}✅ Database connection OK${NC}"
echo ""

# 构建测试命令
cd ../..
TEST_CMD="go test ./tests/integration/..."

# 添加标志
if [ "$VERBOSE" = true ]; then
    TEST_CMD="$TEST_CMD -v"
fi

if [ "$RACE" = true ]; then
    TEST_CMD="$TEST_CMD -race"
    echo -e "${BLUE}🏃 Running with race detector${NC}"
fi

if [ "$COVERAGE" = true ]; then
    TEST_CMD="$TEST_CMD -cover -coverprofile=tests/integration/coverage.out"
    echo -e "${BLUE}📊 Generating coverage report${NC}"
fi

if [ -n "$SPECIFIC_TEST" ]; then
    TEST_CMD="$TEST_CMD -run $SPECIFIC_TEST"
    echo -e "${BLUE}🎯 Running specific test: $SPECIFIC_TEST${NC}"
fi

echo ""
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}Running: $TEST_CMD${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# 运行测试
START_TIME=$(date +%s)

if $TEST_CMD; then
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✅ All tests passed! (${DURATION}s)${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    
    # 如果生成了覆盖率报告
    if [ "$COVERAGE" = true ]; then
        echo ""
        echo -e "${BLUE}📊 Generating HTML coverage report...${NC}"
        go tool cover -html=tests/integration/coverage.out -o tests/integration/coverage.html
        echo -e "${GREEN}Coverage report saved to: tests/integration/coverage.html${NC}"
        
        # 显示覆盖率摘要
        echo ""
        echo -e "${BLUE}Coverage Summary:${NC}"
        go tool cover -func=tests/integration/coverage.out | tail -5
    fi
    
    exit 0
else
    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    
    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}❌ Tests failed! (${DURATION}s)${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${YELLOW}Troubleshooting tips:${NC}"
    echo "  1. Check database connection and permissions"
    echo "  2. Review test output above for error details"
    echo "  3. Ensure test database is clean: psql -U $DB_USER -d $DB_NAME"
    echo "  4. Re-run setup: ./tests/integration/setup_test_db.sh"
    echo ""
    
    exit 1
fi
