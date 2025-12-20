#!/bin/bash

# 复杂测试场景演示脚本
# 用途：逐个运行并展示各种复杂测试场景

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${BLUE}  🎭 NeverBroke 复杂场景测试演示  ${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

cd ../..

# 场景1: 同一用户多设备登录
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📱 场景1: 同一用户从多个设备登录${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟：用户从 iPhone, iPad, MacBook, Web, Android 5个设备登录"
echo "验证：所有设备都能成功登录，每个设备有独立的 token"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestMultiUser_SameUserDifferentDevices/SameUserLoginFromMultipleDevices

echo ""
read -p "按回车继续下一个场景..." 
echo ""

# 场景2: 多设备并发操作
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}🔄 场景2: 多设备并发操作同一账户${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟："
echo "  Device 1: 3笔支出 (每笔100元)"
echo "  Device 2: 2笔收入 (每笔500元)"
echo "  Device 3: 4笔支出 (每笔50元)"
echo ""
echo "预期结果："
echo "  初始余额: 10,000元"
echo "  最终余额: 10,500元 (10000 - 300 + 1000 - 200)"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestMultiUser_SameUserDifferentDevices/MultiDevicesConcurrentOperations

echo ""
read -p "按回车继续下一个场景..." 
echo ""

# 场景3: 家庭共享账户
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}👨‍👩‍👧 场景3: 家庭共享账户 - 夫妻同时消费${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟："
echo "  👨 丈夫: 在超市购物 (5笔随机小额交易)"
echo "  👩 妻子: 网上购物 (3笔随机交易)"
echo "  💰 同时收到工资 (5000元)"
echo ""
echo "验证："
echo "  - 所有交易都被正确记录"
echo "  - 最终余额计算准确"
echo "  - 交易时间线合理"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestRealWorldScenario_FamilySharedAccount/SimultaneousShoppingByCouple

echo ""
read -p "按回车继续下一个场景..." 
echo ""

# 场景4: 企业多部门运营
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}🏢 场景4: 企业多部门并发运营${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟："
echo "  👤 财务部: 批量发放10个员工工资"
echo "  📊 销售部: 记录5笔销售收入"
echo "  🛒 采购部: 记录3笔采购支出"
echo ""
echo "预期："
echo "  - 总计18笔交易 (10+5+3)"
echo "  - 所有交易并发执行"
echo "  - 最终余额保持一致"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestRealWorldScenario_BusinessDayOperations

echo ""
read -p "按回车继续下一个场景..." 
echo ""

# 场景5: 移动端网络不稳定
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}📱 场景5: 移动端网络不稳定场景${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟："
echo "  用户在地铁上，网络信号不稳定"
echo "  尝试记录午餐支出35元"
echo "  由于网络延迟，连续点击了5次提交"
echo ""
echo "验证："
echo "  - 记录的交易数 × 35元 = 实际余额变化"
echo "  - 数据保持一致性"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestRealWorldScenario_MobileAppUsage/UnstableNetworkScenario

echo ""
read -p "按回车继续下一个场景..." 
echo ""

# 场景6: 多用户并发操作
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}👥 场景6: 多个不同用户并发操作${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "模拟："
echo "  3个不同用户"
echo "  每个用户并发创建10笔交易"
echo "  总计30笔并发交易"
echo ""
echo "验证："
echo "  - 每个用户的数据完全隔离"
echo "  - 每个用户恰好10笔交易"
echo "  - 每个用户余额独立计算正确"
echo ""
read -p "按回车键开始测试..." 

go test ./tests/integration/... -v -run TestMultiUser_DifferentUsersConcurrentOperations/MultipleUsersCreateTransactionsConcurrently

echo ""
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}⚠️  下一个是高并发压力测试，会需要较长时间...${NC}"
echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
read -p "是否运行压力测试？(y/N): " run_stress
echo ""

if [[ $run_stress =~ ^[Yy]$ ]]; then
    # 场景7: 高并发压力测试
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}🔥 场景7: 高并发压力测试${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "压力参数:"
    echo "  - 10个用户"
    echo "  - 每人3个设备"
    echo "  - 每设备10笔交易"
    echo "  - 总计: 300笔并发交易"
    echo ""
    echo "性能指标:"
    echo "  - 成功率: ≥90%"
    echo "  - 数据一致性: 100%"
    echo ""
    
    go test ./tests/integration/... -v -run TestRealWorldScenario_HighConcurrencyStress -timeout 2m
    
    echo ""
else
    echo -e "${YELLOW}跳过压力测试${NC}"
fi

echo ""
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${GREEN}✨ 所有场景演示完成！${NC}"
echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""
echo "测试覆盖的场景:"
echo "  ✅ 同一用户多设备登录和操作"
echo "  ✅ 多设备并发操作同一账户"
echo "  ✅ 家庭共享账户夫妻同时消费"
echo "  ✅ 企业多部门并发运营"
echo "  ✅ 移动端网络不稳定场景"
echo "  ✅ 多用户数据隔离验证"
if [[ $run_stress =~ ^[Yy]$ ]]; then
    echo "  ✅ 高并发压力测试 (300+ 交易)"
fi
echo ""
echo "查看完整测试结果:"
echo "  ./run_tests.sh -v"
echo ""
echo "查看特定场景说明:"
echo "  cat COMPLEX_SCENARIOS.md"
echo ""
