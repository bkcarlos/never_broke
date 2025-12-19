# NeverBroke - 功能实现详细文档

> **项目简介**: NeverBroke（认知翻身账）是一个现代化的个人财务管理应用，帮助用户掌控财务，实现财务自由。

**技术栈**: Go + Gin + PostgreSQL + Vue 3 + TypeScript + Tailwind CSS

**最后更新**: 2024年12月18日

---

## 目录

1. [系统架构](#系统架构)
2. [认证与授权系统](#认证与授权系统)
3. [账户管理](#账户管理)
4. [交易记录](#交易记录)
5. [分期付款](#分期付款)
6. [周期收入管理](#周期收入管理)
7. [薪资控制系统](#薪资控制系统)
8. [预算管理](#预算管理)
9. [现金流预测](#现金流预测)
10. [报表分析](#报表分析)
11. [时间线视图](#时间线视图)
12. [数据导入](#数据导入)
13. [日志系统](#日志系统)
14. [数据库设计](#数据库设计)

---

## 系统架构

### 后端架构 (Go)

```
backend-go/
├── cmd/
│   ├── server/          # 主服务入口
│   ├── keygen/          # RSA密钥生成工具
│   ├── migrate/         # 数据库迁移工具
│   └── test-gorm/       # GORM测试工具
├── common/
│   └── logger/          # 统一日志系统（基于Zap）
├── internal/
│   ├── accounts/        # 账户管理
│   ├── auth/            # 认证授权（JWT + OAuth）
│   ├── budget/          # 预算管理
│   ├── cashflow/        # 现金流预测
│   ├── citysocialrate/  # 城市社保比例
│   ├── home/            # 首页概览
│   ├── importer/        # 数据导入
│   ├── incomeplan/      # 收入计划
│   ├── installment/     # 分期付款
│   ├── ledger/          # 账本明细
│   ├── recurringincome/ # 周期收入（含薪资）
│   ├── reports/         # 报表分析
│   ├── timeline/        # 时间线
│   ├── user/            # 用户设置
│   ├── http/            # HTTP相关（路由、响应）
│   ├── repository/      # 数据访问层
│   │   ├── postgres/    # PostgreSQL实现
│   │   └── memoryadapter/ # 内存存储适配器
│   └── storage/         # 存储层抽象
├── migrations/          # SQL迁移脚本
└── keys/               # JWT密钥存储（已忽略）
```

### 前端架构 (Vue 3)

```
frontend-vue/
├── src/
│   ├── api/            # API封装层
│   │   ├── accounts.ts
│   │   ├── budget.ts
│   │   ├── cashflow.ts
│   │   ├── citySocialRate.ts
│   │   ├── home.ts
│   │   ├── http.ts     # Axios配置
│   │   ├── importer.ts
│   │   ├── incomePlans.ts
│   │   ├── installments.ts
│   │   ├── ledger.ts
│   │   ├── meSalary.ts
│   │   ├── recurringIncome.ts
│   │   ├── report.ts
│   │   ├── reports.ts
│   │   ├── timeline.ts
│   │   └── user.ts
│   ├── components/     # 公共组件
│   │   ├── Icon.vue
│   │   └── RadialActionMenu.vue
│   ├── i18n/          # 国际化（中英文）
│   ├── router/        # 路由配置
│   ├── types/         # TypeScript类型定义
│   ├── utils/         # 工具函数
│   └── views/         # 页面组件（16个页面）
└── public/            # 静态资源
```

### 技术特点

- **模块化设计**: 每个功能模块独立，职责清晰
- **类型安全**: 全程TypeScript，Go强类型
- **RESTful API**: 统一的API设计规范
- **响应式设计**: 支持移动端和桌面端
- **国际化**: 完整的中英文支持
- **安全性**: JWT认证、密码哈希、SQL注入防护

---

## 认证与授权系统

### 功能概述

实现了完整的现代化认证系统，支持：
- ✅ 邮箱密码注册/登录
- ✅ JWT Token认证（Access Token + Refresh Token）
- ✅ OAuth第三方登录（微信、Google、Apple等）
- ✅ 多端登录支持（Web、App、小程序）
- ✅ Token轮换和重放检测
- ✅ 设备管理和会话管理

### 技术实现

#### 1. Token方案设计

**Access Token**:
- 类型: JWT (HS256签名)
- 有效期: 10分钟
- 用途: API请求认证
- 存储: 前端内存（不持久化）
- 包含信息: user_id, email, nickname

**Refresh Token**:
- 类型: JWT (HS256签名)
- 有效期: 7天
- 用途: 刷新Access Token
- 存储: 
  - Web端: HttpOnly Cookie（安全）
  - App/小程序: 客户端安全存储
- 数据库存储: SHA-256哈希（禁止明文）

#### 2. 安全特性

**密码安全**:
```go
// 使用 argon2id 哈希算法
func HashPassword(password string) (string, error) {
    // 参数: time=1, memory=64MB, threads=4, keyLen=32
    hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    return base64.RawStdEncoding.EncodeToString(hash), nil
}
```

**Token轮换**:
- 每次刷新生成新的Refresh Token
- 旧Token标记为replaced_by（防重放）
- 检测到重放攻击时撤销整个设备

**设备管理**:
- 每个设备独立的Refresh Token
- 支持单设备登出/全部设备登出
- 记录设备信息（User-Agent、IP）

#### 3. 数据库表设计

**users表** (用户基本信息):
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,  -- argon2id哈希
    nickname VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

**refresh_tokens表** (刷新令牌):
```sql
CREATE TABLE refresh_tokens (
    id TEXT PRIMARY KEY,              -- UUID v4 (作为jti)
    user_id INTEGER NOT NULL,
    device_id TEXT NOT NULL,          -- 设备ID
    token_hash TEXT NOT NULL,         -- SHA-256哈希
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,             -- 撤销时间
    replaced_by TEXT,                 -- 轮换链路
    last_used_at TIMESTAMP,
    user_agent TEXT,
    ip_address TEXT
);
```

**oauth_identities表** (第三方登录):
```sql
CREATE TABLE oauth_identities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    provider VARCHAR(50) NOT NULL,    -- wechat, google, apple
    subject TEXT NOT NULL,            -- OIDC sub / 微信unionid
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(provider, subject)
);
```

**oauth_states表** (OAuth防CSRF):
```sql
CREATE TABLE oauth_states (
    state TEXT PRIMARY KEY,           -- 随机state（32字节）
    provider VARCHAR(50) NOT NULL,
    redirect_uri TEXT,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL     -- 5分钟过期
);
```

#### 4. API接口

**注册**:
```http
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "nickname": "张三",
  "device_id": "uuid-device-id"
}

Response 200:
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",  // App端返回，Web端在Cookie
  "expires_in": 600,
  "token_type": "Bearer",
  "user": {
    "email": "user@example.com",
    "nickname": "张三"
  }
}
```

**登录**:
```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "device_id": "uuid-device-id"
}
```

**刷新Token**:
```http
POST /api/auth/refresh
Content-Type: application/json
Cookie: refresh_token=xxx  // Web端

{
  "refresh_token": "xxx",  // App端传body
  "device_id": "uuid-device-id"
}
```

**登出**:
```http
POST /api/auth/logout
Authorization: Bearer <access_token>

{
  "device_id": "uuid-device-id"  // 可选，指定设备
}
```

**全部设备登出**:
```http
POST /api/auth/logout_all
Authorization: Bearer <access_token>
```

**获取当前用户信息**:
```http
GET /api/me
Authorization: Bearer <access_token>

Response 200:
{
  "email": "user@example.com",
  "nickname": "张三"
}
```

#### 5. OAuth第三方登录

**支持的提供商**:
- 微信（网站应用扫码登录）
- Google（OIDC，待实现）
- Apple（OIDC，待实现）
- GitHub（OIDC，待实现）

**微信登录流程**:

1. 开始OAuth流程:
```http
GET /api/oauth/wechat/start?redirect_uri=https://yourapp.com/callback

Response 200:
{
  "auth_url": "https://open.weixin.qq.com/connect/qrconnect?...",
  "state": "random-state-string"
}
```

2. 用户扫码授权后，微信回调:
```http
GET /api/oauth/wechat/callback?code=xxx&state=xxx&device_id=xxx

Response 200:
{
  "access_token": "eyJhbGc...",
  "expires_in": 600,
  "token_type": "Bearer",
  "user": {
    "email": "wechat_xxx@oauth.local",
    "nickname": "微信用户"
  }
}
```

**实现细节**:
```go
// 微信提供商实现
type WeChatProvider struct {
    appID       string  // 从环境变量读取
    appSecret   string
    redirectURI string
}

// 获取授权URL
func (p *WeChatProvider) GetAuthURL(state string) string {
    return "https://open.weixin.qq.com/connect/qrconnect?" +
           "appid=" + p.appID +
           "&redirect_uri=" + p.redirectURI +
           "&response_type=code" +
           "&scope=snsapi_login" +
           "&state=" + state
}

// 交换code获取用户信息
func (p *WeChatProvider) ExchangeCode(code string) (*OAuthUserInfo, error) {
    // 1. 用code换access_token
    // 2. 用access_token获取用户信息（unionid）
    // 3. 返回标准化的用户信息
}
```

#### 6. 认证中间件

```go
// Gin中间件：验证Access Token
func (h *TokenHandler) AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从Header获取token
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }
        
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // 2. 验证token
        claims, err := h.service.ValidateAccessToken(token)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
            return
        }
        
        // 3. 将用户信息存入context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        c.Next()
    }
}
```

#### 7. 安全最佳实践

**密钥管理**:
```bash
# 生成JWT密钥
go run cmd/keygen/main.go

# 环境变量配置
export JWT_SECRET="your-secret-key-min-32-bytes"
export WECHAT_APP_ID="wx..."
export WECHAT_APP_SECRET="..."
```

**防护措施**:
- ✅ 密码使用argon2id哈希（抗GPU破解）
- ✅ Refresh Token哈希存储（SHA-256）
- ✅ Token轮换防重放攻击
- ✅ OAuth State防CSRF攻击
- ✅ 定期清理过期数据（每小时）
- ✅ 限流保护（待实现）
- ✅ HTTPS传输（生产环境必须）

**定期清理**:
```go
// 每小时清理一次过期数据
go func() {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        tokenService.CleanupExpiredData()
        // 清理: 过期的refresh_tokens, oauth_states
    }
}()
```

---

## 账户管理

### 功能概述

支持多种账户类型的管理，包括现金、银行卡、电子钱包和信用卡。

### 账户类型

**1. 现金账户 (cash)**:
- 用途: 现金管理
- 特点: 简单余额记录

**2. 银行账户 (bank)**:
- 用途: 银行储蓄卡
- 支持机构: 工行、建行、农行、中行、交行、招行等15家银行
- 特点: 记录卡号尾4位

**3. 电子钱包 (wallet)**:
- 用途: 微信、支付宝等
- 支持: 微信、支付宝
- 特点: 实时余额

**4. 信用卡 (credit)**:
- 用途: 信用卡管理
- 特点: 
  - 记录总额度、临时额度
  - 已用额度（balance）
  - 可用额度 = 总额度 + 临时额度 - 已用额度

### 数据模型

```typescript
interface Account {
  id: string
  type: 'cash' | 'bank' | 'wallet' | 'credit'
  institution?: string  // 机构代码
  cardLast4?: string    // 卡号尾4位
  currency: string      // 货币（如CNY）
  balance: number       // 余额（信用卡为已用额度）
  creditLimit?: number  // 信用卡总额度
  tempLimit?: number    // 信用卡临时额度
  archived: boolean     // 是否归档
}
```

### 支持的银行机构

```typescript
type AccountInstitution =
  | 'icbc'   // 工商银行
  | 'ccb'    // 建设银行
  | 'abc'    // 农业银行
  | 'boc'    // 中国银行
  | 'bocom'  // 交通银行
  | 'cmb'    // 招商银行
  | 'spdb'   // 浦发银行
  | 'citic'  // 中信银行
  | 'ceb'    // 光大银行
  | 'hxb'    // 华夏银行
  | 'cmbc'   // 民生银行
  | 'cgb'    // 广发银行
  | 'pab'    // 平安银行
  | 'cib'    // 兴业银行
  | 'psbc'   // 邮储银行
  | 'wechat' // 微信
  | 'alipay' // 支付宝
  | 'other'  // 其他
```

### API接口

**获取账户列表**:
```http
GET /api/v1/accounts

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": [
    {
      "id": "acc_001",
      "type": "bank",
      "institution": "cmb",
      "cardLast4": "8888",
      "currency": "CNY",
      "balance": 50000.00,
      "archived": false
    },
    {
      "id": "acc_002",
      "type": "credit",
      "institution": "cmb",
      "cardLast4": "6666",
      "currency": "CNY",
      "balance": 3000.00,      // 已用额度
      "creditLimit": 20000.00, // 总额度
      "tempLimit": 5000.00,    // 临时额度
      "archived": false
    }
  ]
}
```

**创建账户**:
```http
POST /api/v1/accounts
Content-Type: application/json

{
  "type": "bank",
  "institution": "cmb",
  "cardLast4": "8888",
  "currency": "CNY",
  "balance": 50000.00
}

Response 200:
{
  "code": 0,
  "message": "Account created",
  "data": { ... }
}
```

**更新信用卡额度**:
```http
PATCH /api/v1/accounts/:id/credit-limits
Content-Type: application/json

{
  "creditLimit": 25000.00,
  "tempLimit": 10000.00
}
```

### 数据库设计

```sql
CREATE TABLE accounts (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,              -- cash/bank/wallet/credit
  institution TEXT,
  card_last4 TEXT,
  currency TEXT NOT NULL,
  balance DOUBLE PRECISION NOT NULL DEFAULT 0,
  credit_limit DOUBLE PRECISION,
  temp_limit DOUBLE PRECISION,
  archived BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_accounts_user ON accounts(user_id);
CREATE INDEX idx_accounts_user_archived ON accounts(user_id, archived);
```

### 前端页面

**账户管理页面** (`AccountManagementPage.vue`):
- 显示所有账户列表
- 按类型分组展示
- 显示可用额度/余额
- 支持归档/取消归档

**新建账户页面** (`NewAccountPage.vue`):
- 表单式创建账户
- 根据类型动态显示字段
- 实时验证

---

## 交易记录

### 功能概述

记录所有收入、支出和转账交易，支持分类、备注、关联账户等功能。

### 交易类型

**1. 支出 (expense)**:
- 从账户扣款
- 必须指定分类
- 可关联分期计划

**2. 收入 (income)**:
- 向账户入账
- 可选分类
- 支持周期收入自动生成

**3. 转账 (transfer)**:
- 从一个账户转到另一个账户
- 不计入收支统计
- 金额相等

### 数据模型

```typescript
interface Transaction {
  id: string
  type: 'income' | 'expense' | 'transfer'
  amount: number
  categoryKey?: string         // 分类键
  remark?: string             // 备注
  occurredAt: string          // 发生时间 ISO8601
  accountId?: string          // 账户ID（收入/支出）
  fromAccountId?: string      // 转出账户（转账）
  toAccountId?: string        // 转入账户（转账）
  installmentPlanId?: string  // 关联分期计划
  installmentPeriod?: number  // 第几期
  installmentTotal?: number   // 总期数
}
```

### API接口

**获取每日交易**:
```http
GET /api/v1/transactions/daily?date=2024-12-18

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "date": "2024-12-18",
    "stats": {
      "totalIncome": 10000.00,
      "totalExpense": 1500.00,
      "balance": 8500.00
    },
    "transactions": [
      {
        "id": "txn_001",
        "type": "income",
        "amount": 10000.00,
        "categoryKey": "salary",
        "remark": "12月工资",
        "occurredAt": "2024-12-18T10:00:00Z",
        "accountId": "acc_001"
      },
      {
        "id": "txn_002",
        "type": "expense",
        "amount": 1500.00,
        "categoryKey": "food",
        "remark": "超市购物",
        "occurredAt": "2024-12-18T15:30:00Z",
        "accountId": "acc_001"
      }
    ]
  }
}
```

**创建交易**:
```http
POST /api/v1/transactions
Content-Type: application/json

{
  "type": "expense",
  "amount": 150.00,
  "categoryKey": "food",
  "remark": "午餐",
  "occurredAt": "2024-12-18T12:00:00Z",
  "accountId": "acc_001"
}
```

### 分类系统

支持的分类（可扩展）:

**支出分类**:
- `food` - 餐饮
- `transport` - 交通
- `shopping` - 购物
- `entertainment` - 娱乐
- `housing` - 住房
- `healthcare` - 医疗
- `education` - 教育
- `other` - 其他

**收入分类**:
- `salary` - 工资
- `bonus` - 奖金
- `investment` - 投资收益
- `other` - 其他收入

### 数据库设计

```sql
CREATE TABLE transactions (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,              -- income/expense/transfer
  amount DOUBLE PRECISION NOT NULL,
  category_key TEXT,
  remark TEXT,
  occurred_at TIMESTAMPTZ NOT NULL,
  account_id TEXT,                 -- 单账户交易
  from_account_id TEXT,            -- 转账：转出
  to_account_id TEXT,              -- 转账：转入
  installment_plan_id TEXT,        -- 关联分期
  installment_period INT,
  installment_total INT
);

CREATE INDEX idx_transactions_user_time 
  ON transactions(user_id, occurred_at DESC);
CREATE INDEX idx_transactions_user_account 
  ON transactions(user_id, account_id);
CREATE INDEX idx_transactions_installment_plan 
  ON transactions(installment_plan_id) 
  WHERE installment_plan_id IS NOT NULL;
```

### 前端页面

**账本明细页面** (`DetailPage.vue`):
- 按日期查看交易
- 显示每日统计
- 支持筛选和搜索

**记账页面** (`RecordFormPage.vue`):
- 快速记账表单
- 支持三种交易类型
- 智能分类建议

**转账页面** (`TransferPage.vue`):
- 账户间转账
- 实时显示余额
- 验证转账金额

---

## 分期付款

### 功能概述

支持信用卡分期管理，包括等额本息和等额本金两种还款方式，自动生成还款计划和交易记录。

### 还款方式

**1. 等额本息 (equal_payment)**:
- 每月还款金额相同
- 前期利息多，后期本金多
- 计算公式: `月还款额 = [本金 × 月利率 × (1+月利率)^期数] / [(1+月利率)^期数 - 1]`

**2. 等额本金 (equal_principal)**:
- 每月本金相同
- 每月利息递减
- 总利息少于等额本息
- 计算公式: `月还款额 = 本金/期数 + (本金 - 已还本金) × 月利率`

### 数据模型

```typescript
interface InstallmentPlan {
  id: string
  accountId: string              // 信用卡账户
  principal: number              // 本金
  totalInterest: number          // 总利息
  totalAmount: number            // 总金额
  termMonths: number             // 期数
  monthlyRate: number            // 月利率（小数）
  repayMethod: 'equal_principal' | 'equal_payment'
  currency: string
  firstDueDate: string           // 首期还款日 YYYY-MM-DD
  status: 'active' | 'finished' | 'canceled'
  createdAt: string
}

interface InstallmentPreviewItem {
  period: number                 // 第几期
  dueDate: string               // 还款日期
  principal: number             // 本金
  interest: number              // 利息
  total: number                 // 总额
  remainingPrincipal: number    // 剩余本金
}
```

### API接口

**预览分期计划**:
```http
POST /api/v1/installments/preview
Content-Type: application/json

{
  "principal": 10000.00,
  "termMonths": 12,
  "monthlyRatePercent": 0.6,     // 月利率0.6%
  "repayMethod": "equal_payment",
  "firstDueDate": "2025-01-15",
  "currency": "CNY"
}

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "principal": 10000.00,
    "totalInterest": 395.04,
    "totalAmount": 10395.04,
    "termMonths": 12,
    "monthlyRate": 0.006,
    "repayMethod": "equal_payment",
    "schedule": [
      {
        "period": 1,
        "dueDate": "2025-01-15",
        "principal": 806.63,
        "interest": 60.00,
        "total": 866.63,
        "remainingPrincipal": 9193.37
      },
      {
        "period": 2,
        "dueDate": "2025-02-15",
        "principal": 811.47,
        "interest": 55.16,
        "total": 866.63,
        "remainingPrincipal": 8381.90
      },
      // ... 其余10期
    ]
  }
}
```

**创建支出分期**:
```http
POST /api/v1/installments/expense
Content-Type: application/json

{
  "accountId": "acc_credit_001",
  "principal": 10000.00,
  "termMonths": 12,
  "monthlyRatePercent": 0.6,
  "repayMethod": "equal_payment",
  "firstDueDate": "2025-01-15",
  "currency": "CNY",
  "categoryKey": "shopping",
  "remark": "iPhone 15 Pro",
  "purchaseDate": "2024-12-18"
}

Response 200:
{
  "code": 0,
  "message": "Installment created",
  "data": {
    "plan": { ... },
    "transactions": [
      {
        "id": "txn_001",
        "type": "expense",
        "amount": 10000.00,
        "categoryKey": "shopping",
        "remark": "iPhone 15 Pro (分期本金)",
        "occurredAt": "2024-12-18T00:00:00Z",
        "accountId": "acc_credit_001",
        "installmentPlanId": "plan_001",
        "installmentPeriod": 0,
        "installmentTotal": 12
      },
      {
        "id": "txn_002",
        "type": "expense",
        "amount": 866.63,
        "categoryKey": "shopping",
        "remark": "iPhone 15 Pro (第1期还款)",
        "occurredAt": "2025-01-15T00:00:00Z",
        "accountId": "acc_credit_001",
        "installmentPlanId": "plan_001",
        "installmentPeriod": 1,
        "installmentTotal": 12
      },
      // ... 其余11期交易
    ]
  }
}
```

### 计算逻辑

**等额本息计算** (`service.go`):
```go
func CalculateEqualPayment(principal float64, monthlyRate float64, termMonths int) []InstallmentPreviewItem {
    // 月还款额
    monthlyPayment := principal * monthlyRate * math.Pow(1+monthlyRate, float64(termMonths)) /
                     (math.Pow(1+monthlyRate, float64(termMonths)) - 1)
    
    remaining := principal
    schedule := make([]InstallmentPreviewItem, termMonths)
    
    for i := 0; i < termMonths; i++ {
        interest := remaining * monthlyRate
        principalPart := monthlyPayment - interest
        remaining -= principalPart
        
        schedule[i] = InstallmentPreviewItem{
            Period:             i + 1,
            DueDate:            calculateDueDate(firstDueDate, i),
            Principal:          principalPart,
            Interest:           interest,
            Total:              monthlyPayment,
            RemainingPrincipal: remaining,
        }
    }
    
    return schedule
}
```

**等额本金计算**:
```go
func CalculateEqualPrincipal(principal float64, monthlyRate float64, termMonths int) []InstallmentPreviewItem {
    // 每月本金
    monthlyPrincipal := principal / float64(termMonths)
    
    remaining := principal
    schedule := make([]InstallmentPreviewItem, termMonths)
    
    for i := 0; i < termMonths; i++ {
        interest := remaining * monthlyRate
        total := monthlyPrincipal + interest
        remaining -= monthlyPrincipal
        
        schedule[i] = InstallmentPreviewItem{
            Period:             i + 1,
            DueDate:            calculateDueDate(firstDueDate, i),
            Principal:          monthlyPrincipal,
            Interest:           interest,
            Total:              total,
            RemainingPrincipal: remaining,
        }
    }
    
    return schedule
}
```

### 数据库设计

```sql
CREATE TABLE installment_plans (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  account_id TEXT NOT NULL,
  principal DOUBLE PRECISION NOT NULL,
  total_interest DOUBLE PRECISION NOT NULL,
  total_amount DOUBLE PRECISION NOT NULL,
  term_months INT NOT NULL,
  monthly_rate DOUBLE PRECISION NOT NULL,
  repay_method TEXT NOT NULL,
  currency TEXT NOT NULL,
  first_due_date DATE NOT NULL,
  status TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_installment_user_account 
  ON installment_plans(user_id, account_id);
CREATE INDEX idx_installment_user_status 
  ON installment_plans(user_id, status);
```

### 前端页面

**分期设置页面** (`SalarySettingsPage.vue`):
- 预览分期计划
- 选择还款方式
- 查看详细还款表
- 创建分期计划

### 业务逻辑

1. **创建分期时**:
   - 创建分期计划记录
   - 生成本金支出交易（购买日）
   - 生成所有还款交易（未来日期）
   - 所有交易关联分期计划ID

2. **还款提醒**:
   - 可通过查询未来交易实现
   - 支持推送通知（待实现）

3. **提前还款**:
   - 修改分期状态为finished
   - 删除未来的还款交易
   - 创建提前还款交易

---

## 周期收入管理

### 功能概述

管理固定周期的收入，如工资、房租收入、副业收入等，支持自动生成收入交易。

### 收入类型

- `salary` - 工资（支持详细薪资计算）
- `rent` - 房租收入
- `side` - 固定副业
- `other` - 其他周期收入

### 收入频率

- `monthly` - 每月
- `bimonthly` - 每两个月
- `quarterly` - 每季度
- `semiannual` - 每半年
- `weekly` - 每周
- `yearly` - 每年

### 数据模型

```typescript
interface RecurringIncomePlan {
  id: string
  type: 'salary' | 'rent' | 'side' | 'other'
  label: string                    // 展示名称
  accountId: string                // 到账账户
  amount: number                   // 每次金额（工资为税后）
  currency: string
  frequency: RecurringIncomeFrequency
  dayOfMonth?: number              // 月频率：1-31
  weekday?: number                 // 周频率：1-7
  nextDueDate: string              // 下次应收日期
  lastFiredDate?: string           // 上次执行日期
  cityCode?: string                // 薪资：纳税城市
  active: boolean
  createdAt: string
  
  // 薪资专用字段
  salaryPreTaxMonthly?: number     // 税前月薪
  salarySocialRate?: number        // 社保比例
  salaryExtraDeductions?: number   // 专项附加扣除
  salaryPayDay?: number            // 发薪日
  salaryStartDate?: string         // 开始日期
}
```

### API接口

**获取周期收入列表**:
```http
GET /api/v1/recurring-incomes

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": [
    {
      "id": "plan_001",
      "type": "salary",
      "label": "工资",
      "accountId": "acc_001",
      "amount": 15000.00,
      "currency": "CNY",
      "frequency": "monthly",
      "dayOfMonth": 15,
      "nextDueDate": "2025-01-15",
      "cityCode": "beijing",
      "active": true,
      "salaryPreTaxMonthly": 20000.00,
      "salarySocialRate": 0.222,
      "salaryExtraDeductions": 1000.00
    }
  ]
}
```

**创建/更新周期收入**:
```http
POST /api/v1/recurring-incomes
Content-Type: application/json

{
  "type": "rent",
  "label": "房租收入",
  "accountId": "acc_001",
  "amount": 3000.00,
  "currency": "CNY",
  "frequency": "monthly",
  "dayOfMonth": 1,
  "nextDueDate": "2025-01-01",
  "active": true
}
```

**获取到期计划**:
```http
GET /api/v1/recurring-incomes/due

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": [
    {
      "id": "plan_001",
      "label": "工资",
      "amount": 15000.00,
      "nextDueDate": "2024-12-15"
    }
  ]
}
```

**实现收入**:
```http
POST /api/v1/recurring-incomes/:id/realize
Content-Type: application/json

{
  "date": "2024-12-15"  // 可选，默认今天
}

Response 200:
{
  "code": 0,
  "message": "Income realized",
  "data": {
    "plan": { ... },
    "transaction": {
      "id": "txn_001",
      "type": "income",
      "amount": 15000.00,
      "categoryKey": "salary",
      "remark": "工资",
      "occurredAt": "2024-12-15T00:00:00Z",
      "accountId": "acc_001"
    }
  }
}
```

**获取未来收入投射**:
```http
GET /api/v1/recurring-incomes/projected-transactions?days=90

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": [
    {
      "date": "2024-12-15",
      "accountId": "acc_001",
      "amount": 15000.00,
      "type": "income",
      "label": "工资",
      "planId": "plan_001",
      "isSalary": true
    },
    {
      "date": "2025-01-01",
      "accountId": "acc_001",
      "amount": 3000.00,
      "type": "income",
      "label": "房租收入",
      "planId": "plan_002",
      "isSalary": false
    }
  ]
}
```

### 数据库设计

```sql
CREATE TABLE recurring_income_plans (
  id TEXT PRIMARY KEY,
  user_id UUID NOT NULL,
  type TEXT NOT NULL,
  label TEXT NOT NULL,
  account_id TEXT NOT NULL,
  amount DOUBLE PRECISION NOT NULL,
  currency TEXT NOT NULL,
  frequency TEXT NOT NULL,
  day_of_month INT,
  weekday INT,
  next_due_date DATE NOT NULL,
  last_fired_date DATE,
  city_code TEXT,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL,
  
  -- 薪资专用字段
  salary_pre_tax_monthly DOUBLE PRECISION,
  salary_social_rate DOUBLE PRECISION,
  salary_extra_deductions DOUBLE PRECISION,
  salary_pay_day INT,
  salary_start_date DATE
);

CREATE INDEX idx_recurring_income_user 
  ON recurring_income_plans(user_id);
CREATE INDEX idx_recurring_income_user_active 
  ON recurring_income_plans(user_id, active);
CREATE INDEX idx_recurring_income_next_due_date 
  ON recurring_income_plans(next_due_date) 
  WHERE active = TRUE;
```

### 自动执行逻辑

```go
// 定时任务：每天检查到期的周期收入
func CheckAndRealizeIncomes() {
    today := time.Now().Format("2006-01-02")
    
    // 查询到期的计划
    plans := db.Query(`
        SELECT * FROM recurring_income_plans 
        WHERE active = TRUE 
        AND next_due_date <= ?
    `, today)
    
    for _, plan := range plans {
        // 创建收入交易
        CreateIncomeTransaction(plan)
        
        // 更新下次应收日期
        UpdateNextDueDate(plan)
    }
}
```

### 前端页面

**周期收入配置页面** (`RecurringIncomeSettingsPage.vue`):
- 查看所有周期收入
- 创建/编辑/删除计划
- 查看到期提醒
- 手动实现收入

---

## 薪资控制系统

### 功能概述

这是一个完整的薪资管理系统，支持：
- ✅ 城市社保比例管理
- ✅ 工资方案配置
- ✅ 税后收入计算
- ✅ 全年薪资预览
- ✅ 个税累进计算
- ✅ 与现金流预测集成

### 核心功能

#### 1. 城市社保比例管理

**支持城市**:
- 北京: 22.2% (养老8% + 医疗2% + 失业0.2% + 公积金12%)
- 上海: 17.5% (养老8% + 医疗2% + 失业0.5% + 公积金7%)
- 广州: 22.0%
- 深圳: 21.5%
- 其他: 20.0% (默认)

**数据模型**:
```typescript
interface CitySocialRate {
  id: string
  cityCode: string              // beijing, shanghai, etc.
  cityName: string
  effectiveDate: string
  pensionRate: number           // 养老保险
  medicalRate: number           // 医疗保险
  unemploymentRate: number      // 失业保险
  injuryRate: number            // 工伤保险（个人通常不缴）
  maternityRate: number         // 生育保险（个人通常不缴）
  housingFundRate: number       // 住房公积金
  totalRate: number             // 总比例
  isActive: boolean
  remark?: string
}

interface CitySocialRateDetail {
  cityCode: string
  cityName: string
  totalRate: number
  components: SocialRateComponent[]
}

interface SocialRateComponent {
  name: string                  // 养老/医疗/失业/公积金
  rate: number
}
```

**API接口**:
```http
GET /api/v1/city-social-rates

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": [
    {
      "id": "csr_001",
      "cityCode": "beijing",
      "cityName": "北京",
      "totalRate": 0.222,
      "effectiveDate": "2024-01-01",
      "isActive": true
    }
  ]
}
```

```http
GET /api/v1/city-social-rates/beijing

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "cityCode": "beijing",
    "cityName": "北京",
    "totalRate": 0.222,
    "components": [
      { "name": "养老保险", "rate": 0.08 },
      { "name": "医疗保险", "rate": 0.02 },
      { "name": "失业保险", "rate": 0.002 },
      { "name": "住房公积金", "rate": 0.12 }
    ]
  }
}
```

#### 2. 工资方案配置

**配置参数**:
```typescript
interface SalaryPlanConfig {
  preTaxMonthly: number         // 税前月薪
  cityCode: string              // 纳税城市
  startDate: string             // 开始日期
  payDay: number                // 发薪日（1-31）
  socialRate: number            // 社保比例（可自定义）
  extraDeductions: number       // 专项附加扣除
  accountId: string             // 到账账户
  currency: string
}
```

**API接口**:

获取工资方案:
```http
GET /api/v1/me/salary-plan

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "plan": {
      "id": "plan_001",
      "type": "salary",
      "label": "工资",
      "accountId": "acc_001",
      "amount": 15000.00,           // 税后金额
      "frequency": "monthly",
      "dayOfMonth": 15,
      "nextDueDate": "2025-01-15",
      "cityCode": "beijing",
      "active": true,
      "salaryPreTaxMonthly": 20000.00,
      "salarySocialRate": 0.222,
      "salaryExtraDeductions": 1000.00,
      "salaryPayDay": 15,
      "salaryStartDate": "2024-01-01"
    }
  }
}
```

保存工资方案:
```http
POST /api/v1/me/salary-plan
Content-Type: application/json

{
  "preTaxMonthly": 20000.00,
  "cityCode": "beijing",
  "startDate": "2024-01-01",
  "payDay": 15,
  "socialRate": 0.222,
  "extraDeductions": 1000.00,
  "accountId": "acc_001",
  "currency": "CNY"
}

Response 200:
{
  "code": 0,
  "message": "Salary plan saved",
  "data": {
    "plan": { ... },
    "preview": {
      "cityCode": "beijing",
      "year": 2024,
      "items": [ ... ],
      "totalPreTax": 240000.00,
      "totalTax": 18000.00,
      "totalPostTax": 180000.00
    }
  }
}
```

删除工资方案:
```http
DELETE /api/v1/me/salary-plan/:id

Response 200:
{
  "code": 0,
  "message": "Salary plan deleted"
}
```

#### 3. 薪资计算逻辑

**中国个税累进税率表**:
```go
var TaxBrackets = []TaxBracket{
    {Threshold: 0,      Rate: 0.03, QuickDeduction: 0},
    {Threshold: 36000,  Rate: 0.10, QuickDeduction: 2520},
    {Threshold: 144000, Rate: 0.20, QuickDeduction: 16920},
    {Threshold: 300000, Rate: 0.25, QuickDeduction: 31920},
    {Threshold: 420000, Rate: 0.30, QuickDeduction: 52920},
    {Threshold: 660000, Rate: 0.35, QuickDeduction: 85920},
    {Threshold: 960000, Rate: 0.45, QuickDeduction: 181920},
}
```

**计算流程**:
```go
func CalculateSalaryPreview(req SalaryPreviewRequest) SalaryPreviewResponse {
    // 1. 确定社保比例
    socialRate := req.SocialRate
    if socialRate == nil {
        // 使用城市默认比例
        cityRate := GetCitySocialRate(req.CityCode)
        socialRate = &cityRate.TotalRate
    }
    
    // 2. 生成发薪日期列表（从startDate到年底）
    dates := GeneratePayDates(req.StartDate, req.PayDay, req.Frequency)
    
    // 3. 逐月计算
    items := make([]SalaryPreviewItem, len(dates))
    cumulativePreTax := 0.0
    cumulativeTax := 0.0
    
    for i, date := range dates {
        // 3.1 扣除社保
        socialSecurity := req.PreTaxAmount * (*socialRate)
        
        // 3.2 计算应纳税所得额
        taxableIncome := req.PreTaxAmount - socialSecurity - 5000 - req.ExtraDeductions
        
        // 3.3 累计应纳税所得额
        cumulativePreTax += req.PreTaxAmount
        cumulativeTaxableIncome := cumulativePreTax - 
                                   socialSecurity*float64(i+1) - 
                                   5000*float64(i+1) - 
                                   req.ExtraDeductions*float64(i+1)
        
        // 3.4 查税率表计算累计税额
        totalTax := CalculateTaxByBracket(cumulativeTaxableIncome)
        
        // 3.5 本月应缴税额 = 累计税额 - 已缴税额
        currentMonthTax := totalTax - cumulativeTax
        cumulativeTax = totalTax
        
        // 3.6 税后金额
        postTax := req.PreTaxAmount - socialSecurity - currentMonthTax
        
        items[i] = SalaryPreviewItem{
            Date:             date,
            PreTax:           req.PreTaxAmount,
            SocialSecurity:   socialSecurity,
            TaxableIncome:    taxableIncome,
            TaxAmount:        currentMonthTax,
            PostTax:          postTax,
            CumulativePreTax: cumulativePreTax,
            CumulativeTax:    cumulativeTax,
        }
    }
    
    return SalaryPreviewResponse{
        CityCode:     req.CityCode,
        Year:         time.Now().Year(),
        Items:        items,
        TotalPreTax:  cumulativePreTax,
        TotalTax:     cumulativeTax,
        TotalPostTax: cumulativePreTax - cumulativeTax - socialSecurity*float64(len(dates)),
    }
}
```

**查税率表**:
```go
func CalculateTaxByBracket(taxableIncome float64) float64 {
    if taxableIncome <= 0 {
        return 0
    }
    
    for i := len(TaxBrackets) - 1; i >= 0; i-- {
        bracket := TaxBrackets[i]
        if taxableIncome > bracket.Threshold {
            return taxableIncome*bracket.Rate - bracket.QuickDeduction
        }
    }
    
    return 0
}
```

#### 4. 薪资预览

**预览请求**:
```http
POST /api/v1/recurring-incomes/salary/preview
Content-Type: application/json

{
  "preTaxAmount": 20000.00,
  "cityCode": "beijing",
  "startDate": "2024-12-01",
  "payDay": 15,
  "frequency": "monthly",
  "currency": "CNY",
  "socialRate": 0.222,
  "extraDeductions": 1000.00
}

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "cityCode": "beijing",
    "year": 2024,
    "items": [
      {
        "date": "2024-12-15",
        "preTax": 20000.00,
        "socialSecurity": 4440.00,
        "taxableIncome": 9560.00,
        "taxAmount": 286.80,
        "postTax": 15273.20,
        "cumulativePreTax": 20000.00,
        "cumulativeTax": 286.80
      }
    ],
    "totalPreTax": 20000.00,
    "totalTax": 286.80,
    "totalPostTax": 15273.20
  }
}
```

### 数据库设计

**城市社保比例表**:
```sql
CREATE TABLE city_social_rates (
  id TEXT PRIMARY KEY,
  city_code TEXT NOT NULL,
  city_name TEXT NOT NULL,
  effective_date DATE NOT NULL,
  pension_rate DOUBLE PRECISION NOT NULL,
  medical_rate DOUBLE PRECISION NOT NULL,
  unemployment_rate DOUBLE PRECISION NOT NULL,
  injury_rate DOUBLE PRECISION DEFAULT 0,
  maternity_rate DOUBLE PRECISION DEFAULT 0,
  housing_fund_rate DOUBLE PRECISION NOT NULL,
  total_rate DOUBLE PRECISION NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT TRUE,
  remark TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_city_social_rates_city_active 
  ON city_social_rates(city_code, is_active);
```

**周期收入计划表（扩展）**:
```sql
ALTER TABLE recurring_income_plans 
  ADD COLUMN salary_pre_tax_monthly DOUBLE PRECISION,
  ADD COLUMN salary_social_rate DOUBLE PRECISION,
  ADD COLUMN salary_extra_deductions DOUBLE PRECISION,
  ADD COLUMN salary_pay_day INT,
  ADD COLUMN salary_start_date DATE;
```

### 前端页面

**薪资控制页面** (`MeSalaryPage.vue`):

**Section 1: 当前工资方案卡片**
- 显示税前月薪、发薪日、到账账户
- 显示纳税城市、状态
- 编辑/删除按钮

**Section 2: 工资参数配置表单**
- 税前月薪输入
- 纳税城市选择（自动加载社保比例）
- 社保比例（显示默认值，可自定义）
- 社保比例明细表格
- 专项附加扣除
- 开始日期、发薪日
- 到账账户选择
- 保存/预览按钮

**Section 3: 薪资计算器预览**
- 顶部汇总卡片:
  - 全年税前总额
  - 全年个税总额
  - 全年税后总额
- 月度明细表格:
  - 列: 月份、税前、社保+公积金、个税、税后到手
- 底部提示: "仅供参考，以实际发薪与当地政策为准"

**页面特点**:
- 实时计算预览
- 自动加载城市社保比例
- 表单验证
- 加载状态管理
- 错误提示

### 业务流程

**保存工资方案流程**:
1. 用户填写工资参数
2. 前端调用预览API（可选）
3. 用户确认后调用保存API
4. 后端计算税后金额
5. 查询现有salary计划并设为inactive
6. 创建新的周期收入计划
7. 返回计划和预览结果

**数据一致性**:
- 工资方案的`amount`字段存储税后金额
- 税前金额和计算参数存储在专用字段
- 与现金流预测系统无缝集成
- 只允许一个active的工资方案

### 扩展建议

1. **年终奖计算**: 单独配置年终奖的税务计算
2. **多地工作**: 支持同时配置多个城市的工资
3. **历史记录**: 查看历史工资方案和实际发薪记录对比
4. **智能提醒**: 发薪日前一天推送提醒
5. **数据导出**: 导出全年薪资预览为Excel/PDF
6. **社保基数**: 支持配置各城市的社保缴费基数上下限

---

## 预算管理

### 功能概述

设置和管理月度预算，帮助用户控制支出。

### 数据模型

```typescript
interface MonthlyBudget {
  month: string       // 'YYYY-MM'
  totalBudget: number
}
```

### API接口

**获取月度预算**:
```http
GET /api/v1/budget/monthly?month=2024-12

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "month": "2024-12",
    "totalBudget": 5000.00
  }
}
```

**设置月度预算**:
```http
PUT /api/v1/budget/monthly
Content-Type: application/json

{
  "month": "2024-12",
  "totalBudget": 5000.00
}
```

### 前端页面

**预算设置页面** (`BudgetSettingsPage.vue`):
- 选择月份
- 设置总预算
- 查看当前支出
- 显示预算进度

---

## 现金流预测

### 功能概述

基于历史数据和未来计划，预测未来30/60/90天的现金流情况，帮助用户提前规避资金风险。

### 预测逻辑

**数据来源**:
1. 当前账户余额
2. 未来周期收入（工资、房租等）
3. 未来分期还款
4. 历史平均支出

**计算公式**:
```
每日余额 = 期初余额 + 当日流入 - 当日流出
```

### 数据模型

```typescript
interface CashflowForecast {
  startDate: string
  endDate: string
  base: 'total_assets' | 'account'
  accountId?: string
  safeAmount: number              // 安全垫金额
  points: DailyForecastPoint[]
  status: 'safe' | 'warning' | 'danger'
  dangerDates: string[]           // 余额<0的日期
  warningDates: string[]          // 余额<安全垫的日期
}

interface DailyForecastPoint {
  date: string
  openingBalance: number          // 期初余额
  inflow: number                  // 流入
  outflow: number                 // 流出
  closingBalance: number          // 期末余额
}
```

### API接口

```http
GET /api/v1/cashflow/forecast?startDate=2024-12-18&days=90&base=total_assets&safeAmount=5000

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "startDate": "2024-12-18",
    "endDate": "2025-03-17",
    "base": "total_assets",
    "safeAmount": 5000.00,
    "points": [
      {
        "date": "2024-12-18",
        "openingBalance": 50000.00,
        "inflow": 0.00,
        "outflow": 500.00,
        "closingBalance": 49500.00
      },
      {
        "date": "2024-12-19",
        "openingBalance": 49500.00,
        "inflow": 0.00,
        "outflow": 300.00,
        "closingBalance": 49200.00
      },
      {
        "date": "2025-01-15",
        "openingBalance": 35000.00,
        "inflow": 15000.00,        // 工资
        "outflow": 866.63,         // 分期还款
        "closingBalance": 49133.37
      }
    ],
    "status": "safe",
    "dangerDates": [],
    "warningDates": ["2024-12-30", "2024-12-31"]
  }
}
```

### 计算逻辑

```go
func ForecastCashflow(req CashflowForecastRequest) CashflowForecastResponse {
    // 1. 获取起始余额
    balance := GetCurrentBalance(req.Base, req.AccountID)
    
    // 2. 获取未来收入（周期收入）
    futureIncomes := GetProjectedIncomes(req.StartDate, req.Days)
    
    // 3. 获取未来支出（分期还款）
    futureExpenses := GetProjectedExpenses(req.StartDate, req.Days)
    
    // 4. 计算历史平均日支出
    avgDailyExpense := CalculateAvgDailyExpense()
    
    // 5. 逐日计算
    points := make([]DailyForecastPoint, req.Days)
    dangerDates := []string{}
    warningDates := []string{}
    
    for i := 0; i < req.Days; i++ {
        date := req.StartDate.AddDate(0, 0, i)
        
        // 当日流入
        inflow := GetInflowForDate(futureIncomes, date)
        
        // 当日流出
        outflow := GetOutflowForDate(futureExpenses, date)
        if outflow == 0 {
            outflow = avgDailyExpense  // 使用历史平均
        }
        
        // 计算余额
        closingBalance := balance + inflow - outflow
        
        points[i] = DailyForecastPoint{
            Date:           date.Format("2006-01-02"),
            OpeningBalance: balance,
            Inflow:         inflow,
            Outflow:        outflow,
            ClosingBalance: closingBalance,
        }
        
        // 判断状态
        if closingBalance < 0 {
            dangerDates = append(dangerDates, date.Format("2006-01-02"))
        } else if closingBalance < req.SafeAmount {
            warningDates = append(warningDates, date.Format("2006-01-02"))
        }
        
        balance = closingBalance
    }
    
    // 6. 确定整体状态
    status := "safe"
    if len(dangerDates) > 0 {
        status = "danger"
    } else if len(warningDates) > 0 {
        status = "warning"
    }
    
    return CashflowForecastResponse{
        StartDate:    req.StartDate.Format("2006-01-02"),
        EndDate:      req.StartDate.AddDate(0, 0, req.Days-1).Format("2006-01-02"),
        Base:         req.Base,
        AccountID:    req.AccountID,
        SafeAmount:   req.SafeAmount,
        Points:       points,
        Status:       status,
        DangerDates:  dangerDates,
        WarningDates: warningDates,
    }
}
```

### 前端页面

**现金流预测页面** (`CashflowForecastPage.vue`):
- 选择预测天数（30/60/90天）
- 选择预测基准（总资产/单个账户）
- 设置安全垫金额
- 显示预测曲线图
- 高亮危险日期和警告日期
- 显示关键事件（收入/支出）

---

## 报表分析

### 功能概述

提供多维度的财务报表分析，包括月度报表、分类统计、趋势分析等。

### 报表类型

#### 1. 月度报表

**数据模型**:
```typescript
interface MonthlyReport {
  month: string                   // 'YYYY-MM'
  totalIncome: number
  totalExpense: number
  balance: number
  categories: CategoryReportItem[]
  dailyTrend: DailyTrendPoint[]
}

interface CategoryReportItem {
  categoryKey: string
  amount: number
  percentage: number
  count: number
}

interface DailyTrendPoint {
  date: string                    // 'YYYY-MM-DD'
  expense: number
  income: number
}
```

**API接口**:
```http
GET /api/v1/reports/monthly?month=2024-12

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "month": "2024-12",
    "totalIncome": 15000.00,
    "totalExpense": 8500.00,
    "balance": 6500.00,
    "categories": [
      {
        "categoryKey": "food",
        "amount": 3000.00,
        "percentage": 35.3,
        "count": 45
      },
      {
        "categoryKey": "transport",
        "amount": 1500.00,
        "percentage": 17.6,
        "count": 30
      }
    ],
    "dailyTrend": [
      {
        "date": "2024-12-01",
        "expense": 200.00,
        "income": 0.00
      },
      {
        "date": "2024-12-15",
        "expense": 500.00,
        "income": 15000.00
      }
    ]
  }
}
```

#### 2. 月度汇总

**数据模型**:
```typescript
interface SummaryReport {
  period: 'monthly' | 'yearly'
  year: number
  month?: number
  totalIncome: number
  totalExpense: number
  balance: number
  categories: CategorySummary[]
  trend: TrendPoint[]
  compare: CompareSummary
}

interface CategorySummary {
  categoryKey: string
  amount: number
  percentage: number
}

interface TrendPoint {
  date: string
  income: number
  expense: number
  balance: number
}

interface CompareSummary {
  prevYear: number
  prevMonth: number
  yearOverYear: number
  monthOverMonth: number
}
```

**API接口**:
```http
GET /api/v1/reports/monthly-summary?year=2024&month=12

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "period": "monthly",
    "year": 2024,
    "month": 12,
    "totalIncome": 15000.00,
    "totalExpense": 8500.00,
    "balance": 6500.00,
    "categories": [ ... ],
    "trend": [ ... ],
    "compare": {
      "prevYear": 7200.00,
      "prevMonth": 6800.00,
      "yearOverYear": -9.7,
      "monthOverMonth": -4.4
    }
  }
}
```

### 前端页面

**报表页面** (`ReportsPage.vue`):
- 选择月份/年份
- 显示收支汇总
- 分类饼图
- 每日趋势折线图
- 同比/环比分析

---

## 时间线视图

### 功能概述

以时间线方式展示历史交易和未来预测，帮助用户全面了解财务状况。

### 模式

- `history` - 仅历史交易
- `with_future` - 历史 + 未来预测

### 数据模型

```typescript
interface DailyTimeline {
  fromDate: string
  toDate: string
  mode: 'history' | 'with_future'
  summaries: DailySummary[]
}

interface DailySummary {
  date: string
  income: number
  expense: number
  balance: number
  hasTransactions: boolean
}
```

### API接口

**获取每日汇总**:
```http
GET /api/v1/timeline/daily-summary?fromDate=2024-12-01&toDate=2024-12-31&mode=with_future

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "fromDate": "2024-12-01",
    "toDate": "2024-12-31",
    "mode": "with_future",
    "summaries": [
      {
        "date": "2024-12-18",
        "income": 0.00,
        "expense": 500.00,
        "balance": -500.00,
        "hasTransactions": true
      },
      {
        "date": "2025-01-15",
        "income": 15000.00,
        "expense": 866.63,
        "balance": 14133.37,
        "hasTransactions": true
      }
    ]
  }
}
```

**获取每日明细**:
```http
GET /api/v1/timeline/daily-detail?date=2024-12-18

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "date": "2024-12-18",
    "transactions": [ ... ],
    "projectedTransactions": [ ... ]
  }
}
```

### 前端页面

**时间线页面** (`TimelinePage.vue`):
- 日历视图
- 每日收支标记
- 点击查看详情
- 切换历史/预测模式

---

## 数据导入

### 功能概述

支持从银行账单、支付宝、微信等导入交易数据。

### 数据模型

```typescript
interface ImportPreviewItem {
  date: string
  categoryGuess: string
  amount: number
  remark: string
}

interface ImportPreviewResult {
  count: number
  items: ImportPreviewItem[]
}
```

### API接口

**预览导入**:
```http
POST /api/v1/import/preview
Content-Type: multipart/form-data

file: [CSV/Excel文件]

Response 200:
{
  "code": 0,
  "message": "OK",
  "data": {
    "count": 100,
    "items": [
      {
        "date": "2024-12-18",
        "categoryGuess": "food",
        "amount": 50.00,
        "remark": "肯德基"
      }
    ]
  }
}
```

### 前端页面

**导入账单页面** (`ImportBillsPage.vue`):
- 上传文件
- 预览数据
- 调整分类
- 确认导入

---

## 日志系统

### 功能概述

基于Uber Zap的高性能日志系统，支持结构化日志、多级别、多输出。

### 特性

- ✅ 结构化日志（JSON/Console格式）
- ✅ 多级别（Debug/Info/Warn/Error/Fatal）
- ✅ 多输出（控制台/文件/多文件）
- ✅ Gin中间件集成
- ✅ 彩色控制台输出
- ✅ 线程安全
- ✅ 零内存分配

### 配置

```go
logger.Init(logger.Config{
    Level: "info",
    OutputPaths: []string{
        "stdout",
        "./logs/app.log",
    },
    Format: "console",
})
defer logger.Sync()
```

### 使用示例

```go
// 简单日志
logger.Info("Server started")

// 带字段的日志
logger.Info("User logged in",
    zap.String("userID", "123"),
    zap.String("ip", "192.168.1.1"),
)

logger.Error("Database connection failed",
    zap.Error(err),
    zap.String("host", "localhost"),
    zap.Int("port", 5432),
)
```

### Gin中间件

```go
r := gin.New()
r.Use(logger.GinLogger())    // 记录HTTP请求
r.Use(logger.GinRecovery())  // Panic恢复
```

### 日志格式

**Console格式**:
```
2024-12-18T10:30:45.123+0800    INFO    home/handler.go:25    Home overview fetched successfully
```

**JSON格式**:
```json
{
  "time": "2024-12-18T10:30:45.123+0800",
  "level": "INFO",
  "caller": "home/handler.go:25",
  "msg": "Home overview fetched successfully"
}
```

---

## 数据库设计

### 表结构概览

```
users                    - 用户基本信息
refresh_tokens           - 刷新令牌
oauth_identities         - 第三方登录身份
oauth_states             - OAuth状态
accounts                 - 账户
transactions             - 交易记录
installment_plans        - 分期计划
recurring_income_plans   - 周期收入计划
city_social_rates        - 城市社保比例
```

### 索引策略

**高频查询优化**:
- `accounts`: user_id, (user_id, archived)
- `transactions`: (user_id, occurred_at DESC), (user_id, account_id)
- `recurring_income_plans`: (user_id, active), next_due_date
- `refresh_tokens`: user_id, device_id, token_hash, expires_at

### 数据隔离

**用户数据隔离**:
- 所有业务表包含`user_id`字段
- 查询时必须带上`user_id`条件
- 防止跨用户数据泄露

### 迁移管理

**迁移脚本**:
```
migrations/
├── 0001_init.sql                    - 初始化表结构
├── 0002_add_salary_tables.sql       - 薪资功能
├── 0003_seed_city_social_rates.sql  - 城市社保数据
├── 0004_add_auth_tables.sql         - 认证系统
└── 0005_add_token_based_auth.sql    - Token认证
```

**执行迁移**:
```bash
go run cmd/migrate/main.go
```

---

## API规范

### 统一响应格式

```typescript
interface ApiResponse<T> {
  code: number      // 0=成功，非0=错误
  message: string
  data?: T
}
```

### 错误码

```
0     - 成功
400   - 请求参数错误
401   - 未授权
403   - 禁止访问
404   - 资源不存在
500   - 服务器错误
```

### 认证方式

```http
Authorization: Bearer <access_token>
```

### CORS配置

```go
AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}
AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"}
AllowCredentials: true
```

---

## 国际化

### 支持语言

- 🇨🇳 简体中文 (zh-CN)
- 🇺🇸 English (en-US)

### 实现方式

**前端**:
```typescript
// i18n/zh-CN.ts
export default {
  home: {
    title: '首页',
    todayExpense: '今日支出',
    // ...
  }
}

// 使用
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
console.log(t('home.title'))  // "首页"
```

### 文本覆盖

- ✅ 所有页面标题
- ✅ 所有表单标签
- ✅ 所有按钮文本
- ✅ 所有提示信息
- ✅ 所有错误消息

---

## 部署指南

### 后端部署

**环境变量**:
```bash
# 数据库
DATABASE_URL=postgresql://user:pass@localhost:5432/neverbroke

# JWT密钥
JWT_SECRET=your-secret-key-min-32-bytes

# OAuth配置
WECHAT_APP_ID=wx...
WECHAT_APP_SECRET=...
WECHAT_REDIRECT_URI=https://yourdomain.com/api/oauth/wechat/callback

# 环境
ENV=production
```

**构建**:
```bash
cd backend-go
go build -o neverbroke ./cmd/server
```

**运行**:
```bash
./neverbroke
```

### 前端部署

**构建**:
```bash
cd frontend-vue
pnpm install
pnpm build
```

**Nginx配置**:
```nginx
server {
    listen 80;
    server_name yourdomain.com;
    
    root /var/www/neverbroke/dist;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

---

## 安全最佳实践

### 认证安全

- ✅ 密码使用argon2id哈希
- ✅ JWT密钥至少32字节
- ✅ Refresh Token哈希存储
- ✅ Token轮换防重放
- ✅ OAuth State防CSRF
- ✅ HTTPS传输（生产环境）

### 数据安全

- ✅ SQL参数化查询（防注入）
- ✅ 用户数据隔离
- ✅ 敏感字段不序列化
- ✅ 定期清理过期数据

### API安全

- ✅ 限流保护（待实现）
- ✅ CORS白名单
- ✅ 请求大小限制
- ✅ 错误信息脱敏

---

## 测试

### 单元测试

```bash
# 后端
cd backend-go
go test ./...

# 前端
cd frontend-vue
pnpm test
```

### 测试覆盖

- ✅ 账户管理
- ✅ 交易记录
- ✅ 分期计算
- ✅ 周期收入
- ✅ 薪资计算
- ✅ 日志系统

---

## 性能优化

### 后端优化

- ✅ 数据库索引优化
- ✅ 查询结果缓存（待实现）
- ✅ 连接池配置
- ✅ Zap零内存分配日志

### 前端优化

- ✅ 路由懒加载
- ✅ 组件按需加载
- ✅ API请求缓存
- ✅ 图片懒加载（待实现）

---

## 未来规划

### 短期计划

- [ ] 限流保护
- [ ] 数据导出（Excel/PDF）
- [ ] 智能分类建议
- [ ] 预算超支提醒
- [ ] 多设备同步

### 长期计划

- [ ] 移动App（React Native）
- [ ] 微信小程序
- [ ] 数据可视化增强
- [ ] AI财务建议
- [ ] 多人账本
- [ ] 投资组合管理

---

## 贡献指南

欢迎提交Issue和Pull Request！

### 开发流程

1. Fork项目
2. 创建功能分支
3. 提交代码
4. 编写测试
5. 提交PR

### 代码规范

- Go: 遵循官方规范，使用gofmt
- TypeScript: 使用ESLint + Prettier
- 提交信息: 使用Conventional Commits

---

## 许可证

MIT License

---

## 联系方式

- GitHub: [NeverBroke Repository]
- 邮箱: support@neverbroke.com

---

**最后更新**: 2024年12月18日
**版本**: v1.0.0
**文档维护**: NeverBroke Team
