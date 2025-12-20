# 复杂场景测试说明

## 概述

本文档详细说明了新增的复杂测试场景，这些场景模拟了真实世界中的多用户、多设备并发操作情况。

## 测试文件

### 1. `concurrent_users_test.go` - 多用户并发测试

#### 场景 1.1: 同一用户多设备登录

**测试函数**: `TestMultiUser_SameUserDifferentDevices/SameUserLoginFromMultipleDevices`

**模拟场景**:
```
用户: zhangsan@example.com
设备: iPhone, iPad, MacBook, Web浏览器, Android手机
```

**测试流程**:
1. 同一用户从5个不同设备并发登录
2. 验证所有设备都获得唯一的 access_token
3. 验证数据库中有5个活跃的 refresh_token
4. 验证每个设备的 token 都是独立的

**验证点**:
- ✅ 所有5个设备都能成功登录
- ✅ 每个设备有独立的 token
- ✅ 数据库中有5条 refresh_token 记录
- ✅ 所有 token 都处于活跃状态

---

#### 场景 1.2: 多设备并发操作同一账户

**测试函数**: `TestMultiUser_SameUserDifferentDevices/MultiDevicesConcurrentOperations`

**模拟场景**:
```
用户: zhangsan@example.com
账户余额: 10,000元

Device 1 (手机):  创建3笔支出，每笔100元
Device 2 (平板):  创建2笔收入，每笔500元  
Device 3 (电脑):  创建4笔支出，每笔50元

并发执行，同时操作
```

**预期结果**:
```
初始余额: 10,000元
Device 1: -300元 (3 × 100)
Device 2: +1,000元 (2 × 500)
Device 3: -200元 (4 × 50)
------------------------
最终余额: 10,500元
```

**验证点**:
- ✅ 所有9笔交易都成功
- ✅ 最终余额准确 = 10,500元
- ✅ 数据库有9条交易记录
- ✅ 没有交易丢失或重复

---

#### 场景 1.3: 单设备登出不影响其他设备

**测试函数**: `TestMultiUser_SameUserDifferentDevices/OneDeviceLogoutOthersStillWork`

**测试流程**:
1. 用户从3个设备登录
2. Device 1 执行登出操作
3. Device 1 尝试创建交易 → 应该失败
4. Device 2 和 3 创建交易 → 应该成功

**验证点**:
- ✅ Device 1 登出后无法操作
- ✅ Device 2 和 3 仍然可以正常操作
- ✅ Device 1 的 token 在数据库中被标记为已撤销

---

### 2. `real_world_scenarios_test.go` - 真实场景测试

#### 场景 2.1: 家庭共享账户

**测试函数**: `TestRealWorldScenario_FamilySharedAccount`

**模拟场景**:
```
👨 丈夫 (husband@family.com)
  - 设备: iPhone
  - 在超市购物: 5笔小额交易 (50-150元随机)
  - 收到工资: 5,000元

👩 妻子 (wife@family.com)
  - 设备: Android
  - 网上购物: 3笔交易 (200-500元随机)

共享账户初始余额: 10,000元
所有操作并发执行
```

**时间线示例**:
```
T+0ms:    丈夫开始扫描商品
T+20ms:   妻子下第一个订单 (300元)
T+50ms:   丈夫收到工资 (+5000元)
T+80ms:   丈夫扫描第2件商品 (80元)
T+100ms:  妻子下第二个订单 (450元)
T+150ms:  丈夫继续扫描...
...
```

**验证逻辑**:
```go
最终余额 = 初始余额 + 收入 - 丈夫支出 - 妻子支出
        = 10,000 + 5,000 - (5笔随机) - (3笔随机)
```

**关键验证**:
- ✅ 所有交易都被记录
- ✅ 余额计算完全准确（精确到分）
- ✅ 交易时间线合理
- ✅ 无并发冲突导致的数据丢失

---

#### 场景 2.2: 企业日常运营

**测试函数**: `TestRealWorldScenario_BusinessDayOperations`

**模拟场景**:
```
公司账户初始余额: 500,000元

👤 财务部 (finance@company.com)
  - 批量发放工资: 10个员工
  - 每人工资: 5,000-6,000元
  
📊 销售部 (sales@company.com)
  - 记录销售收入: 5笔订单
  - 每笔: 10,000-15,000元
  
🛒 采购部 (procurement@company.com)
  - 记录采购支出: 3笔订单
  - 每笔: 20,000-30,000元

所有操作并发执行
```

**执行流程**:
```
阶段1 (0-100ms):   财务批量发工资
阶段2 (100-200ms): 销售记录收入
阶段3 (100-200ms): 采购记录支出 (与销售同时)
```

**验证点**:
- ✅ 18笔交易全部成功 (10+5+3)
- ✅ 账户余额保持正数
- ✅ 不同部门的操作互不干扰
- ✅ 所有交易都有完整记录

---

#### 场景 2.3: 移动端网络不稳定

**测试函数**: `TestRealWorldScenario_MobileAppUsage/UnstableNetworkScenario`

**模拟场景**:
```
用户在地铁上，网络信号不稳定
用户尝试记录午餐支出 35元
由于网络延迟，用户连续点击了5次提交按钮
```

**测试流程**:
```go
for i := 0; i < 5; i++ {
    go func(attemptNum int) {
        // 模拟网络延迟
        time.Sleep(attemptNum * 50ms)
        
        // 提交相同的交易
        createTransaction("午餐", 35.0)
    }(i)
}
```

**可能的结果**:

**情况A: 无去重机制**
- 5次请求都成功
- 记录了5笔交易
- 余额减少 175元 (5 × 35)

**情况B: 有去重机制**
- 只有1次成功
- 记录了1笔交易
- 余额减少 35元

**验证逻辑**:
```go
交易记录数 × 35元 == 实际余额变化
```

---

#### 场景 2.4: 高并发压力测试

**测试函数**: `TestRealWorldScenario_HighConcurrencyStress`

**压力参数**:
```
用户数:     10
设备数/用户: 3
交易数/设备: 10
------------------------
总交易数:   300 (并发)
```

**执行矩阵**:
```
        Device1  Device2  Device3
User1     10tx     10tx     10tx
User2     10tx     10tx     10tx
User3     10tx     10tx     10tx
...
User10    10tx     10tx     10tx
```

**性能指标**:
- **成功率**: ≥ 90%
- **TPS**: 取决于硬件
- **数据一致性**: 100%

**验证点**:
- ✅ 至少90%的交易成功
- ✅ 每个用户的余额准确
- ✅ 交易记录数正确
- ✅ 无数据竞态问题

**示例输出**:
```
=== Stress Test Results ===
Total transactions: 300
Successful: 287
Failed: 13
Duration: 2.5s
TPS: 114.8
```

---

## 运行方式

### 运行所有复杂场景测试

```bash
cd backend-go

# 多用户并发测试
make test-concurrent

# 真实场景测试
make test-stress

# 运行所有集成测试
make test-integration
```

### 运行特定场景

```bash
cd backend-go/tests/integration

# 同一用户多设备
./run_tests.sh -v -t TestMultiUser_SameUserDifferentDevices

# 家庭共享账户
./run_tests.sh -v -t TestRealWorldScenario_FamilySharedAccount

# 企业运营
./run_tests.sh -v -t TestRealWorldScenario_BusinessDayOperations

# 高并发压力测试
./run_tests.sh -v -t TestRealWorldScenario_HighConcurrencyStress
```

### 跳过压力测试

压力测试默认在 `go test -short` 时会被跳过：

```bash
go test ./tests/integration/... -short
```

---

## 测试价值

### 1. 真实性
所有场景都基于真实的用户使用模式：
- ✅ 家庭成员共享账户
- ✅ 企业多部门协作
- ✅ 移动端网络不稳定
- ✅ 用户多设备使用

### 2. 并发性
大量使用 goroutines 模拟真实并发：
- ✅ 同时多个用户操作
- ✅ 同一用户多个设备
- ✅ 随机时间延迟
- ✅ 竞态条件检测

### 3. 完整性
每个测试都验证多个层面：
- ✅ HTTP 响应正确
- ✅ 数据库记录完整
- ✅ 余额计算准确
- ✅ 关联数据一致

### 4. 可靠性
测试设计考虑了各种边界情况：
- ✅ 网络中断
- ✅ Token 过期
- ✅ 并发冲突
- ✅ 部分失败

---

## 常见问题

### Q: 压力测试运行很慢？

**A**: 这是正常的，300个并发交易需要时间。可以：
- 减少用户数/设备数/交易数
- 使用 `-short` 跳过压力测试
- 增加数据库连接池大小

### Q: 偶尔有交易失败？

**A**: 在高并发下可能出现：
- 数据库连接池耗尽
- 锁竞争超时
- 网络超时

检查失败率，只要 ≥90% 成功就是正常的。

### Q: 如何调试并发问题？

**A**: 使用竞态检测器：
```bash
./run_tests.sh -v -r -t TestMultiUser
```

### Q: 余额不一致怎么办？

**A**: 启用详细日志：
```go
t.Logf("Expected: %.2f, Actual: %.2f", expected, actual)
```

检查每笔交易是否都被正确记录。

---

## 扩展场景

可以基于这些测试添加更多场景：

### 场景 A: 跨境转账
- 多币种账户
- 实时汇率
- 并发转账

### 场景 B: 分期付款
- 创建分期计划
- 自动扣款
- 多个分期并发

### 场景 C: 预算超支
- 设置预算
- 并发支出
- 触发预算警告

### 场景 D: 账单导入
- 批量导入
- 数据验证
- 重复检测

---

## 总结

这些复杂场景测试确保了系统在真实的、高并发的使用环境下仍能保持：

1. **数据准确性** - 余额计算精确到分
2. **操作完整性** - 所有交易都被正确记录
3. **并发安全性** - 无数据竞态和死锁
4. **用户隔离性** - 用户之间数据互不干扰
5. **设备独立性** - 多设备独立又协同

通过这些测试，我们可以自信地说：**系统已经为生产环境做好了准备！** 🚀
