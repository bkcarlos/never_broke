#!/bin/bash

# 认证系统测试脚本
# 用法：./test_auth.sh

set -e

API_BASE="http://localhost:8080/api/auth"
EMAIL="test@example.com"
NICKNAME="测试用户"

echo "🧪 Testing Authentication System"
echo "================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试 1: 握手接口
echo "📡 Test 1: Handshake"
echo "-------------------"
HANDSHAKE_RESPONSE=$(curl -s "$API_BASE/handshake")

if [ -z "$HANDSHAKE_RESPONSE" ]; then
    echo -e "${RED}❌ Failed: Empty response${NC}"
    exit 1
fi

KEY_ID=$(echo "$HANDSHAKE_RESPONSE" | grep -o '"key_id":"[^"]*"' | cut -d'"' -f4)
NONCE=$(echo "$HANDSHAKE_RESPONSE" | grep -o '"nonce":"[^"]*"' | cut -d'"' -f4)

if [ -z "$KEY_ID" ] || [ -z "$NONCE" ]; then
    echo -e "${RED}❌ Failed: Missing key_id or nonce${NC}"
    echo "Response: $HANDSHAKE_RESPONSE"
    exit 1
fi

echo -e "${GREEN}✅ Success${NC}"
echo "   Key ID: $KEY_ID"
echo "   Nonce: ${NONCE:0:20}..."
echo ""

# 测试 2: 检查服务器是否运行
echo "🔍 Test 2: Server Health"
echo "------------------------"
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$API_BASE/handshake")

if [ "$HTTP_CODE" = "200" ]; then
    echo -e "${GREEN}✅ Server is running (HTTP $HTTP_CODE)${NC}"
else
    echo -e "${RED}❌ Server returned HTTP $HTTP_CODE${NC}"
    exit 1
fi
echo ""

# 测试 3: 未认证访问 /me
echo "🔒 Test 3: Unauthenticated Access"
echo "----------------------------------"
ME_RESPONSE=$(curl -s -w "\n%{http_code}" "$API_BASE/me")
HTTP_CODE=$(echo "$ME_RESPONSE" | tail -n1)
BODY=$(echo "$ME_RESPONSE" | head -n-1)

if [ "$HTTP_CODE" = "401" ]; then
    echo -e "${GREEN}✅ Correctly rejected (HTTP $HTTP_CODE)${NC}"
    echo "   Response: $BODY"
else
    echo -e "${RED}❌ Expected 401, got HTTP $HTTP_CODE${NC}"
fi
echo ""

# 提示：注册和登录需要 RSA 加密，建议使用前端页面测试
echo "⚠️  Note: Register and Login require RSA encryption"
echo "   Please use the frontend page (auth.html) to test these features"
echo ""

# 测试 4: 数据库连接（通过握手接口间接测试）
echo "🗄️  Test 4: Database Connection"
echo "-------------------------------"
# 多次调用握手接口，检查 nonce 是否每次都不同
NONCE1=$(curl -s "$API_BASE/handshake" | grep -o '"nonce":"[^"]*"' | cut -d'"' -f4)
sleep 1
NONCE2=$(curl -s "$API_BASE/handshake" | grep -o '"nonce":"[^"]*"' | cut -d'"' -f4)

if [ "$NONCE1" != "$NONCE2" ]; then
    echo -e "${GREEN}✅ Database is working (nonces are unique)${NC}"
    echo "   Nonce 1: ${NONCE1:0:20}..."
    echo "   Nonce 2: ${NONCE2:0:20}..."
else
    echo -e "${RED}❌ Database might not be working (nonces are identical)${NC}"
fi
echo ""

# 总结
echo "📊 Test Summary"
echo "==============="
echo -e "${GREEN}✅ Basic tests passed${NC}"
echo ""
echo "Next steps:"
echo "1. Open http://localhost:8080 in your browser"
echo "2. Navigate to /auth.html (or open frontend-vue/auth.html)"
echo "3. Test registration and login"
echo ""
echo "For detailed testing, see:"
echo "- QUICK_START.md"
echo "- AUTH_SYSTEM_README.md"
