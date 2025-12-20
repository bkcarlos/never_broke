# 集成测试文档

## 概述

集成测试用于验证前端API接口与数据库的完整交互流程，确保：
- 用户操作能正确写入数据库
- 数据能被再次查询并保持一致性
- 异常情况（如网络中断）不会导致数据不一致
- 并发操作的数据完整性

## 测试覆盖

### 1. 认证流程测试 (`auth_integration_test.go`)
- ✅ 完整的注册-登录-登出流程
- ✅ 数据库持久化验证
- ✅ Token 刷新和 Token Rotation
- ✅ 并发注册防重复
- ✅ 错误场景（错误密码、过期token等）

### 2. 交易流程测试 (`transaction_integration_test.go`)
- ✅ 创建收入/支出交易并验证持久化
- ✅ 账户余额更新验证
- ✅ 账户间转账
- ✅ 每日交易查询
- ✅ 事务回滚测试
- ✅ 网络中断场景模拟
- ✅ 并发交易测试

## 前置条件

### 1. 安装 PostgreSQL

```bash
# macOS
brew install postgresql@15
brew services start postgresql@15

# Ubuntu/Debian
sudo apt-get install postgresql-15
sudo systemctl start postgresql
```

### 2. 创建测试数据库

```bash
# 使用提供的脚本
./tests/integration/setup_test_db.sh

# 或手动创建
psql -U postgres
CREATE DATABASE neverbroke_test;
CREATE USER neverbroke WITH PASSWORD 'neverbroke';
GRANT ALL PRIVILEGES ON DATABASE neverbroke_test TO neverbroke;
\q
```

### 3. 配置环境变量

创建 `.env.test` 文件或设置环境变量：

```bash
export NEVERBROKE_TEST_DB=neverbroke_test
export NEVERBROKE_PG_HOST=127.0.0.1
export NEVERBROKE_PG_PORT=5432
export NEVERBROKE_PG_USER=neverbroke
export NEVERBROKE_PG_PASSWORD=neverbroke
export JWT_SECRET=test-secret-key-for-integration-tests-min-32-chars
```

## 运行测试

### 运行所有集成测试

```bash
cd backend-go
go test ./tests/integration/... -v
```

### 运行特定测试

```bash
# 只运行认证测试
go test ./tests/integration/... -v -run TestAuthFlow

# 只运行交易测试
go test ./tests/integration/... -v -run TestTransactionFlow

# 运行特定测试用例
go test ./tests/integration/... -v -run TestAuthFlow_RegisterLoginLogout
```

### 带竞态检测运行

```bash
go test ./tests/integration/... -v -race
```

### 查看测试覆盖率

```bash
go test ./tests/integration/... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 测试特点

### 1. 数据隔离
- 每个测试前清理数据库
- 使用独立的测试数据库（`neverbroke_test`）
- 不影响开发或生产数据

### 2. 真实环境
- 使用真实的 PostgreSQL 数据库
- 完整的 HTTP 请求-响应流程
- 真实的认证和授权流程

### 3. 异常场景覆盖
- ✅ 网络中断模拟
- ✅ 并发操作测试
- ✅ 事务回滚验证
- ✅ 数据一致性检查
- ✅ 重复操作防护

### 4. 性能测试
- ✅ 并发请求处理
- ✅ 数据库连接池验证
- ✅ 事务锁测试

## 测试数据管理

### 清理测试数据

测试会自动清理数据，但如需手动清理：

```bash
psql -U neverbroke -d neverbroke_test
TRUNCATE TABLE transactions CASCADE;
TRUNCATE TABLE accounts CASCADE;
TRUNCATE TABLE users CASCADE;
\q
```

### 查看测试数据

```bash
psql -U neverbroke -d neverbroke_test

-- 查看用户
SELECT * FROM users;

-- 查看交易
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;

-- 查看账户
SELECT * FROM accounts;

-- 查看 token
SELECT user_id, device_id, revoked_at FROM refresh_tokens;
```

## 故障排查

### 1. 连接数据库失败

```bash
# 检查 PostgreSQL 是否运行
pg_isready -h 127.0.0.1 -p 5432

# 检查数据库是否存在
psql -U postgres -c "\l" | grep neverbroke_test

# 检查用户权限
psql -U postgres -d neverbroke_test -c "SELECT current_user;"
```

### 2. 测试超时

增加超时时间：
```bash
go test ./tests/integration/... -v -timeout 5m
```

### 3. 端口占用

确保测试使用的端口未被占用：
```bash
lsof -i :8080
```

## 持续集成

### GitHub Actions 配置示例

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: neverbroke_test
          POSTGRES_USER: neverbroke
          POSTGRES_PASSWORD: neverbroke
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run integration tests
        env:
          NEVERBROKE_TEST_DB: neverbroke_test
          NEVERBROKE_PG_HOST: localhost
          NEVERBROKE_PG_PORT: 5432
          NEVERBROKE_PG_USER: neverbroke
          NEVERBROKE_PG_PASSWORD: neverbroke
          JWT_SECRET: test-secret-key-for-ci
        run: |
          cd backend-go
          go test ./tests/integration/... -v -race
```

## 最佳实践

1. **测试前清理数据**：每个测试开始前调用 `cleanupTestData()`
2. **使用事务隔离**：对需要事务的测试使用 `setupTestTransaction()`
3. **等待异步操作**：使用 `waitForDB()` 等待数据库写入完成
4. **验证数据一致性**：总是验证数据库中的实际数据，不只依赖HTTP响应
5. **测试异常路径**：不只测试成功场景，也要测试失败场景
6. **并发测试**：使用goroutines测试并发场景
7. **清理资源**：使用 `t.Cleanup()` 确保资源被正确清理

## 复杂测试场景示例

### 场景1：家庭共享账户
```bash
# 测试夫妻两人同时在不同地方消费
go test ./tests/integration/... -v -run TestRealWorldScenario_FamilySharedAccount
```

**测试内容：**
- 丈夫在超市购物（5笔小额交易）
- 妻子在网上购物（3笔较大交易）
- 同时收到工资（1笔收入）
- 验证最终余额准确性

### 场景2：企业日常运营
```bash
# 测试公司多部门并发操作
go test ./tests/integration/... -v -run TestRealWorldScenario_BusinessDayOperations
```

**测试内容：**
- 财务部门批量发工资（10人）
- 销售部门记录收入（5笔）
- 采购部门记录支出（3笔）
- 验证所有交易正确记录

### 场景3：高并发压力测试
```bash
# 测试 300+ 并发交易
go test ./tests/integration/... -v -run TestRealWorldScenario_HighConcurrencyStress
```

**测试内容：**
- 10个用户
- 每人3个设备
- 每设备10笔交易
- 总计300笔并发交易
- 验证成功率和数据一致性

### 场景4：同一用户多设备操作
```bash
# 测试用户从多个设备登录和操作
go test ./tests/integration/... -v -run TestMultiUser_SameUserDifferentDevices
```

**测试内容：**
- 从5个不同设备登录
- 多设备并发创建交易
- 单设备登出不影响其他设备
- 验证 Token 隔离

## 扩展测试

可以添加更多测试场景：

- [ ] 分期计划创建和还款测试
- [ ] 周期收入计划测试  
- [ ] 预算管理测试
- [ ] 现金流预测测试
- [ ] 数据导入导出测试
- [✅] 性能基准测试
- [✅] 压力测试
- [✅] 多用户并发测试

## 联系方式

如有问题，请联系开发团队或提交 Issue。
