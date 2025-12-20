# 集成测试快速开始指南

## 5分钟快速上手

### 1. 安装 PostgreSQL（如果还没安装）

**macOS:**
```bash
brew install postgresql@15
brew services start postgresql@15
```

**Ubuntu/Debian:**
```bash
sudo apt-get install postgresql-15
sudo systemctl start postgresql
```

### 2. 设置测试数据库

```bash
cd backend-go/tests/integration
./setup_test_db.sh
```

这个脚本会自动：
- ✅ 创建测试数据库 `neverbroke_test`
- ✅ 创建数据库用户
- ✅ 配置权限
- ✅ 生成 `.env.test` 文件

### 3. 运行测试

**最简单的方式：**
```bash
./run_tests.sh
```

**带详细输出：**
```bash
./run_tests.sh -v
```

**带竞态检测：**
```bash
./run_tests.sh -v -r
```

**生成覆盖率报告：**
```bash
./run_tests.sh -c
# 然后打开 coverage.html 查看
open coverage.html
```

**运行特定测试：**
```bash
./run_tests.sh -t TestAuthFlow              # 只运行认证测试
./run_tests.sh -t TestTransactionFlow       # 只运行交易测试
./run_tests.sh -v -t TestAuthFlow_Register  # 只运行注册测试
```

## 测试内容

### ✅ 认证流程
- 用户注册并写入数据库
- 登录验证和 Token 生成
- Token 刷新和 Token Rotation
- 登出和 Token 撤销
- 并发注册防重复

### ✅ 交易流程  
- 创建收入/支出交易
- 账户余额实时更新
- 账户间转账
- 每日交易查询
- 并发交易处理
- 数据一致性验证

### ✅ 多用户并发
- 同一用户多设备登录
- 多设备并发操作
- 多用户并发交易
- 设备间隔离
- Token 刷新竞态

### ✅ 真实场景
- 家庭共享账户
- 企业多部门运营
- 移动端使用模拟
- 高并发压力测试（300+ 交易）

### ✅ 异常场景
- 网络中断模拟
- 事务回滚验证
- 并发冲突处理
- 数据完整性检查

## 常见问题

### Q: 测试失败提示"database not found"
**A:** 运行 `./setup_test_db.sh` 创建测试数据库

### Q: 测试失败提示"connection refused"
**A:** 确保 PostgreSQL 正在运行：
```bash
# macOS
brew services restart postgresql@15

# Ubuntu
sudo systemctl restart postgresql
```

### Q: 想清空测试数据
**A:** 测试会自动清理，但如需手动清理：
```bash
psql -U neverbroke -d neverbroke_test -c "DROP SCHEMA public CASCADE; CREATE SCHEMA public;"
```

### Q: 如何查看测试数据库内容？
**A:** 
```bash
psql -U neverbroke -d neverbroke_test

-- 查看所有表
\dt

-- 查看用户
SELECT * FROM users;

-- 查看交易
SELECT * FROM transactions ORDER BY created_at DESC LIMIT 10;
```

### Q: 测试运行很慢
**A:** 
- 使用 `-t` 参数只运行特定测试
- 检查数据库性能
- 考虑增加数据库连接池大小

## 测试结果示例

```
🧪 NeverBroke Integration Tests
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

✅ Database connection OK

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Running: go test ./tests/integration/... -v
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

=== RUN   TestAuthFlow_RegisterLoginLogout
=== RUN   TestAuthFlow_RegisterLoginLogout/Register
=== RUN   TestAuthFlow_RegisterLoginLogout/Login_Success
=== RUN   TestAuthFlow_RegisterLoginLogout/Logout
--- PASS: TestAuthFlow_RegisterLoginLogout (0.25s)
    --- PASS: TestAuthFlow_RegisterLoginLogout/Register (0.12s)
    --- PASS: TestAuthFlow_RegisterLoginLogout/Login_Success (0.08s)
    --- PASS: TestAuthFlow_RegisterLoginLogout/Logout (0.05s)

=== RUN   TestTransactionFlow_CreateAndQuery
=== RUN   TestTransactionFlow_CreateAndQuery/CreateIncomeTransaction
=== RUN   TestTransactionFlow_CreateAndQuery/CreateExpenseTransaction
=== RUN   TestTransactionFlow_CreateAndQuery/QueryDailyTransactions
--- PASS: TestTransactionFlow_CreateAndQuery (0.18s)
    --- PASS: TestTransactionFlow_CreateAndQuery/CreateIncomeTransaction (0.06s)
    --- PASS: TestTransactionFlow_CreateAndQuery/CreateExpenseTransaction (0.06s)
    --- PASS: TestTransactionFlow_CreateAndQuery/QueryDailyTransactions (0.06s)

PASS
ok      neverbroke/tests/integration    0.856s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✅ All tests passed! (1s)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

## 下一步

- 📖 阅读完整文档：[README.md](./README.md)
- 🔨 添加更多测试用例
- 📊 查看覆盖率报告
- 🚀 集成到 CI/CD 流程

## 获取帮助

```bash
./run_tests.sh --help
```

或查看完整文档：`cat README.md`
