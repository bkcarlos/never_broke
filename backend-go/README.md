# NeverBroke Backend (Go)

NeverBroke 记账应用的后端服务，使用 Go + Gin 构建。

## 技术栈

- Go 1.21+
- Gin Web Framework
- 内存存储（可扩展为数据库）

## 项目结构

```
backend-go/
├── cmd/
│   └── server/
│       └── main.go              # 主程序入口
├── common/
│   └── logger/
│       ├── logger.go            # 日志封装
│       ├── middleware.go        # Gin 中间件
│       ├── logger_test.go       # 单元测试
│       └── README.md            # 日志使用文档
├── internal/
│   ├── home/
│   │   └── handler.go           # 首页处理器
│   ├── ledger/
│   │   └── handler.go           # 账本处理器
│   ├── report/
│   │   └── handler.go           # 报表处理器
│   ├── user/
│   │   └── handler.go           # 用户处理器
│   ├── http/
│   │   ├── response/
│   │   │   └── response.go      # 统一响应格式
│   │   └── router/
│   │       └── router.go        # 路由配置
│   └── storage/
│       ├── models/
│       │   └── models.go        # 数据模型
│       └── memory/
│           └── store.go         # 内存存储实现
├── logs/
│   └── .gitkeep                 # 日志目录
├── go.mod
└── go.sum
```

## 快速开始

### 方式1: 使用 Mage（推荐）

```bash
# 安装 Mage
go install github.com/magefile/mage@latest

# 查看所有可用命令
mage -l

# 安装依赖
mage install

# 运行服务
mage run

# 或使用热重载
mage dev

# 构建应用
mage build
```

### 方式2: 使用 Go 命令

```bash
# 安装依赖
go mod download

# 运行服务
go run ./cmd/server

# 构建
go build -o neverbroke ./cmd/server
./neverbroke
```

服务将启动在 `http://localhost:8080`

## API 端点

### 首页
- `GET /api/v1/home/overview` - 获取首页概览数据

### 账本明细
- `GET /api/v1/transactions/daily?date=YYYY-MM-DD` - 获取指定日期的账单

### 报表
- `GET /api/v1/reports/monthly?month=YYYY-MM` - 获取月度报表

### 用户设置
- `GET /api/v1/user/settings` - 获取用户设置
- `PATCH /api/v1/user/settings` - 更新用户设置

## 响应格式

所有 API 返回统一的 JSON 格式：

```json
{
  "code": 0,
  "message": "OK",
  "data": { ... }
}
```

- `code`: 0 表示成功，非 0 表示错误
- `message`: 响应消息
- `data`: 响应数据

## 当前实现

### 核心功能
- ✅ 用户认证系统（JWT + Token Rotation）
- ✅ 账户管理（多账户、多币种）
- ✅ 交易记录（收入/支出/转账）
- ✅ 预算管理
- ✅ 分期计划
- ✅ 周期收入
- ✅ 现金流预测
- ✅ 统一响应格式
- ✅ CORS 支持

### 基础设施
- ✅ PostgreSQL 数据库
- ✅ GORM ORM
- ✅ Zap 日志系统
- ✅ Gin 框架
- ✅ 数据迁移系统

### 测试体系
- ✅ 完整的单元测试（80+ 用例）
- ✅ 完整的集成测试（35+ 用例）
- ✅ 多用户并发测试
- ✅ 真实场景压力测试（300+ 并发）
- ✅ 覆盖率报告
- ✅ 竞态检测

### 构建工具
- ✅ Mage 构建系统
- ✅ Makefile（兼容）
- ✅ 热重载支持（air）
- ✅ 多平台交叉编译

## 测试

### 快速运行测试

```bash
# 使用 Mage（推荐）
mage test:all           # 所有测试
mage test:unit          # 单元测试
mage test:integration   # 集成测试
mage test:concurrent    # 并发测试
mage test:stress        # 压力测试
mage test:coverage      # 覆盖率报告

# 查看所有测试命令
mage -l
```

### 测试文档

- [测试快速开始](tests/integration/QUICKSTART.md) - 5分钟上手
- [测试完整指南](tests/TESTING_GUIDE.md) - 详细说明
- [复杂场景测试](tests/integration/COMPLEX_SCENARIOS.md) - 高级场景
- [测试总结报告](tests/TEST_SUMMARY.md) - 完整报告

查看 [项目测试概览](../TESTING_OVERVIEW.md) 了解测试体系全貌。

## 构建工具

### Mage（推荐）

Mage 是基于 Go 的现代构建工具，提供更好的跨平台支持和 IDE 集成。

```bash
# 查看所有命令
mage -l

# 常用命令
mage build              # 构建
mage run                # 运行
mage dev                # 热重载
mage test:all           # 测试
mage ci                 # CI 流程
mage clean              # 清理
```

查看 [Mage 使用指南](../docs/mage/MAGE_GUIDE.md) 了解详细用法。

### Makefile（兼容）

```bash
# 查看所有命令
make help

# 常用命令
make build
make test-all
make clean
```

## 开发说明

### 开发工作流

```bash
# 1. 安装依赖和工具
mage install
mage installTools

# 2. 设置测试数据库
mage test:setupDB

# 3. 启动开发服务器（热重载）
mage dev

# 4. 运行测试
mage test:unit          # 开发时快速测试
mage test:all           # 提交前完整测试

# 5. 代码质量检查
mage format             # 格式化代码
mage vet                # 静态检查
mage lint               # Linter 检查

# 6. 提交前运行 CI
mage ci                 # 完整的 CI 流程
```

### 添加新的 API

1. 在 `internal/` 下创建新模块目录
2. 实现 Handler 和 Service
3. 添加单元测试 (`*_test.go`)
4. 在 `internal/http/router/router.go` 注册路由
5. 在 `internal/models/` 添加数据模型
6. 添加集成测试（如需要）

### 添加公共组件

公共组件放在 `common/` 目录下：
- `common/logger/` - 日志系统
- 其他公共工具和中间件

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `mage format` 格式化代码
- 添加必要的注释和文档
- 编写单元测试（覆盖率 ≥ 80%）
- 提交前运行 `mage ci`

## 相关文档

- [Mage 快速开始](../docs/mage/MAGE_QUICKSTART.md) - 5分钟学会 Mage
- [Mage 使用指南](../docs/mage/MAGE_GUIDE.md) - 构建工具完整说明
- [测试系统文档](../docs/testing/TESTING_GUIDE.md) - 完整的测试体系
- [功能文档](../docs/FEATURE_DOCUMENTATION.md) - 系统功能说明

## 许可证

MIT License

