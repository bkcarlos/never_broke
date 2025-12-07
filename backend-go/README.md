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

### 安装依赖

```bash
go mod download
```

### 运行服务

```bash
go run ./cmd/server
```

服务将启动在 `http://localhost:8080`

### 构建

```bash
go build -o neverbroke ./cmd/server
./neverbroke
```

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

- ✅ 统一响应格式
- ✅ CORS 支持
- ✅ Mock 数据接口
- ✅ 内存存储（用户设置）
- ✅ Zap 日志系统（线程安全，支持多输出）
- ✅ Gin 日志中间件

## 后续计划

- 🔲 数据库集成（SQLite/PostgreSQL）
- 🔲 用户认证（JWT）
- 🔲 真实数据 CRUD
- 🔲 数据验证
- 🔲 日志记录
- 🔲 单元测试

## 开发说明

### 添加新的 API

1. 在 `internal/` 下创建新模块目录
2. 实现 Handler
3. 在 `internal/http/router/router.go` 注册路由
4. 在 `internal/storage/models/models.go` 添加数据模型

### 添加公共组件

公共组件放在 `common/` 目录下：
- `common/logger/` - 日志系统
- 未来可添加：`common/utils/`、`common/middleware/` 等

### 代码规范

- 遵循 Go 官方代码规范
- 使用 `gofmt` 格式化代码
- 添加必要的注释

## 许可证

MIT License

