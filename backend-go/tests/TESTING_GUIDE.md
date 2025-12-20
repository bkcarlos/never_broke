# NeverBroke 测试完整指南

## 📋 目录

- [概述](#概述)
- [测试架构](#测试架构)
- [快速开始](#快速开始)
- [测试类型](#测试类型)
- [使用说明](#使用说明)
- [最佳实践](#最佳实践)

## 概述

NeverBroke 项目包含两类测试：

1. **单元测试** - 测试单个模块和函数的逻辑
2. **集成测试** - 测试完整的 API 流程和数据库交互

### 测试覆盖范围

| 模块 | 单元测试 | 集成测试 | 覆盖率 |
|------|---------|---------|--------|
| 认证 (auth) | ✅ | ✅ | ~90% |
| 账本 (ledger) | ✅ | ✅ | ~95% |
| 账户 (accounts) | ✅ | ⚠️ | ~90% |
| 预算 (budget) | ✅ | ⚠️ | ~85% |
| 分期 (installment) | ✅ | ⚠️ | ~80% |
| 现金流 (cashflow) | ✅ | ⚠️ | ~85% |
| 周期收入 (recurring) | ✅ | ⚠️ | ~80% |

## 测试架构

```
backend-go/
├── internal/                    # 业务代码
│   ├── auth/
│   │   ├── handler.go          # 业务逻辑
│   │   └── *_test.go           # 单元测试
│   ├── ledger/
│   │   ├── handler.go
│   │   └── *_test.go
│   └── ...
│
├── tests/
│   ├── integration/            # 集成测试
│   │   ├── setup_test.go       # 测试环境设置
│   │   ├── auth_integration_test.go
│   │   ├── transaction_integration_test.go
│   │   ├── setup_test_db.sh    # 数据库设置脚本
│   │   ├── run_tests.sh        # 测试运行脚本
│   │   ├── README.md           # 详细文档
│   │   └── QUICKSTART.md       # 快速开始
│   │
│   └── TESTING_GUIDE.md        # 本文档
│
└── Makefile                     # 便捷命令
```

## 快速开始

### 方式1：使用 Makefile（推荐）

```bash
cd backend-go

# 运行所有单元测试
make test-unit

# 设置测试数据库并运行集成测试
make setup-test-db
make test-integration

# 运行所有测试
make test-all

# 生成覆盖率报告
make test-coverage

# 查看所有可用命令
make help
```

### 方式2：使用脚本

```bash
cd backend-go/tests/integration

# 一键设置测试数据库
./setup_test_db.sh

# 运行集成测试
./run_tests.sh -v

# 带覆盖率运行
./run_tests.sh -v -c
```

### 方式3：直接使用 go test

```bash
cd backend-go

# 单元测试
go test ./internal/... -v

# 集成测试
go test ./tests/integration/... -v

# 带竞态检测
go test ./internal/... -v -race
```

## 测试类型

### 1. 单元测试

位置：`internal/*/handler_test.go`

**特点：**
- ✅ 快速执行（毫秒级）
- ✅ 不依赖外部服务
- ✅ 使用 Mock 对象
- ✅ 测试单个功能点

**示例：**
```bash
# 测试认证模块
go test ./internal/auth/... -v

# 测试账本模块
go test ./internal/ledger/... -v

# 测试所有模块
go test ./internal/... -v -short
```

**已覆盖的测试场景：**
- ✅ 账户创建和管理
- ✅ 交易创建（收入/支出/转账）
- ✅ 预算设置和查询
- ✅ 分期计划预览
- ✅ 现金流预测
- ✅ 周期收入管理
- ✅ 各种错误场景

### 2. 集成测试

位置：`tests/integration/*_test.go`

**特点：**
- ✅ 测试完整流程
- ✅ 使用真实数据库
- ✅ 验证数据持久化
- ✅ 测试异常场景

**示例：**
```bash
# 运行所有集成测试
cd tests/integration
./run_tests.sh -v

# 只测试认证流程
./run_tests.sh -v -t TestAuthFlow

# 只测试交易流程
./run_tests.sh -v -t TestTransactionFlow

# 带竞态检测
./run_tests.sh -v -r
```

**已覆盖的测试场景：**

#### 认证流程（auth_integration_test.go）
- ✅ 完整注册-登录-登出流程
- ✅ 数据库持久化验证
- ✅ Token 生成和验证
- ✅ Token 刷新和 Rotation
- ✅ 并发注册防重复
- ✅ 密码错误处理

#### 交易流程（transaction_integration_test.go）
- ✅ 创建收入交易并查询
- ✅ 创建支出交易并查询
- ✅ 账户余额实时更新
- ✅ 账户间转账
- ✅ 每日交易汇总
- ✅ 并发交易处理
- ✅ 网络中断模拟
- ✅ 数据一致性验证
- ✅ 事务回滚测试

## 使用说明

### 环境要求

**必需：**
- Go 1.21+
- PostgreSQL 15+

**可选：**
- golangci-lint（代码检查）
- air（热重载开发）

### 测试数据库设置

集成测试使用独立的测试数据库：

```bash
# 自动设置（推荐）
make setup-test-db

# 或手动设置
cd tests/integration
./setup_test_db.sh
```

这会创建：
- 数据库：`neverbroke_test`
- 用户：`neverbroke`
- 配置文件：`.env.test`

### 查看测试数据

```bash
# 连接到测试数据库
psql -U neverbroke -d neverbroke_test

# 查看用户
SELECT * FROM users;

# 查看交易
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;

# 查看账户
SELECT * FROM accounts;

# 查看 token
SELECT user_id, device_id, revoked_at FROM refresh_tokens;
```

### 清理测试数据

```bash
# 清理测试数据库
make clean-test-db

# 清理所有构建产物
make clean
```

## 测试场景详解

### 异常场景测试

#### 1. 网络中断模拟

```go
// 测试客户端在请求过程中断开连接
ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "POST", "/api/ledger/transactions", body)
// ... 验证数据一致性
```

**验证点：**
- 要么交易完全成功（有记录+余额更新）
- 要么完全失败（无记录+余额不变）
- 不会出现只有记录但余额未更新的情况

#### 2. 并发操作测试

```go
// 并发执行10笔交易
for i := 0; i < 10; i++ {
    go func() {
        // 创建交易
    }()
}
// 验证最终余额正确
```

**验证点：**
- 所有交易都能成功
- 最终余额 = 初始余额 ± 所有交易金额
- 交易记录数量正确

#### 3. 并发注册测试

```go
// 5个goroutine同时注册相同邮箱
for i := 0; i < 5; i++ {
    go func() {
        // 尝试注册
    }()
}
```

**验证点：**
- 只有1个注册成功
- 4个返回"邮箱已存在"错误
- 数据库中只有1条记录

### 数据持久化验证

每个测试都会验证：

1. **HTTP 响应正确**
   ```go
   assert.Equal(t, http.StatusOK, w.Code)
   ```

2. **数据已写入数据库**
   ```go
   var user auth.User
   err := testDB.Where("email = ?", email).First(&user).Error
   require.NoError(t, err)
   ```

3. **关联数据已更新**
   ```go
   var account models.Account
   testDB.First(&account, accountID)
   assert.Equal(t, expectedBalance, account.Balance)
   ```

## 最佳实践

### 1. 测试隔离

```go
func TestSomething(t *testing.T) {
    // 每个测试开始前清理数据
    cleanupTestData(testDB)
    
    // 使用独立的测试数据
    user := createTestUser(t, "unique@email.com")
    
    // ...
}
```

### 2. 异步操作等待

```go
// 创建数据后等待数据库写入完成
waitForDB()

// 然后验证数据
var record Model
testDB.First(&record, id)
```

### 3. 使用子测试

```go
func TestFeature(t *testing.T) {
    t.Run("success_case", func(t *testing.T) {
        // 测试成功场景
    })
    
    t.Run("error_case", func(t *testing.T) {
        // 测试错误场景
    })
}
```

### 4. 资源清理

```go
func TestWithCleanup(t *testing.T) {
    resource := setup()
    
    t.Cleanup(func() {
        cleanup(resource)
    })
    
    // 测试代码
}
```

### 5. 表格驱动测试

```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"case1", "input1", "expected1"},
    {"case2", "input2", "expected2"},
}

for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        result := function(tt.input)
        assert.Equal(t, tt.expected, result)
    })
}
```

## 性能测试

### 基准测试

```bash
# 运行性能测试
make test-benchmark

# 或直接使用 go test
go test ./internal/... -bench=. -benchmem
```

### 压力测试

可以使用工具如 `hey` 或 `wrk`：

```bash
# 安装 hey
go install github.com/rakyll/hey@latest

# 压力测试注册接口
hey -n 1000 -c 10 -m POST \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"Test123!","nickname":"Test"}' \
    http://localhost:8080/api/auth/register
```

## 持续集成

### GitHub Actions 示例

创建 `.github/workflows/test.yml`：

```yaml
name: Tests

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
      
      - name: Run tests
        env:
          NEVERBROKE_TEST_DB: neverbroke_test
          NEVERBROKE_PG_HOST: localhost
          JWT_SECRET: test-secret-key
        run: |
          cd backend-go
          make test-all
```

## 常见问题

### Q: 测试失败怎么办？

1. 检查 PostgreSQL 是否运行
2. 确认测试数据库已创建
3. 查看详细错误信息：`./run_tests.sh -v`
4. 清理并重新设置：`make clean-test-db && make setup-test-db`

### Q: 如何调试测试？

```bash
# 使用 -v 查看详细输出
go test ./internal/auth/... -v

# 运行特定测试
go test ./internal/auth/... -v -run TestLogin

# 使用 dlv 调试器
dlv test ./internal/auth/... -- -test.run TestLogin
```

### Q: 测试覆盖率如何提高？

```bash
# 查看当前覆盖率
make test-coverage

# 查看未覆盖的代码
go tool cover -html=coverage.out
```

## 下一步

- [ ] 添加更多集成测试场景
- [ ] 增加性能基准测试
- [ ] 添加压力测试
- [ ] 实现端到端(E2E)测试
- [ ] 集成到CI/CD流程

## 相关文档

- [集成测试 README](./integration/README.md) - 详细的集成测试文档
- [快速开始](./integration/QUICKSTART.md) - 5分钟快速上手
- [Makefile](../Makefile) - 所有可用命令

## 贡献

欢迎贡献新的测试用例！请确保：
- ✅ 测试有清晰的描述
- ✅ 测试是独立的（不依赖其他测试）
- ✅ 测试覆盖正常和异常场景
- ✅ 测试能够可靠地重复运行
