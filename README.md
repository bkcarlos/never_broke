# NeverBroke - 认知翻身账

一个功能完整、测试完善的记账应用，帮助你掌控财务，永不破产。

## ✨ 特点

- 🚀 **完整的后端系统** - Go + Gin + PostgreSQL
- 🎨 **现代化前端** - Vue 3 + TypeScript + Tailwind CSS
- 🧪 **完善的测试** - 120+ 测试用例，多用户并发测试
- 🔨 **现代构建工具** - Mage 构建系统
- 📚 **完整的文档** - 从入门到精通

## 📚 文档导航

**新手？** 从这里开始：
- 📖 [文档中心](docs/README.md) - 文档入口
- 📚 [文档索引](docs/DOCS_INDEX.md) - 完整文档导航
- ⚡ [后端快速开始](backend-go/README.md) - 5分钟开始开发
- 🧪 [测试快速开始](docs/testing/integration/QUICKSTART.md) - 5分钟运行测试

**开发者？** 看这些：
- 🔨 [Mage 构建工具](docs/mage/MAGE_GUIDE.md) - 所有构建命令
- 📋 [功能文档](docs/FEATURE_DOCUMENTATION.md) - 完整功能说明
- 🧪 [测试体系](docs/TESTING_OVERVIEW.md) - 测试架构和场景

## 技术栈

### 前端
- **框架**: Vue 3 (Composition API)
- **语言**: TypeScript
- **构建工具**: Vite
- **路由**: Vue Router 4
- **状态管理**: Pinia
- **样式**: Tailwind CSS

### 后端
- **语言**: Go 1.21+
- **框架**: Gin Web Framework
- **数据库**: PostgreSQL + GORM
- **认证**: JWT (Token Rotation)
- **日志**: Zap
- **构建**: Mage

## 项目结构

```
NeverBroke/
├── frontend-vue/          # Vue 前端项目
│   ├── src/
│   │   ├── views/        # 页面组件
│   │   ├── components/   # 公共组件
│   │   ├── api/          # API 封装
│   │   ├── types/        # TypeScript 类型
│   │   ├── i18n/         # 国际化
│   │   └── router/       # 路由配置
│   └── package.json
│
└── backend-go/           # Go 后端项目
    ├── cmd/server/       # 主程序入口
    ├── common/           # 公共组件
    │   └── logger/       # 日志系统
    ├── internal/
    │   ├── home/         # 首页模块
    │   ├── ledger/       # 账本模块
    │   ├── report/       # 报表模块
    │   ├── user/         # 用户模块
    │   ├── http/         # HTTP 相关
    │   └── storage/      # 存储层
    └── logs/             # 日志文件目录
```

## 🚀 快速开始

### 前端启动

```bash
cd frontend-vue
npm install
npm run dev
```

前端将运行在 `http://localhost:3000`

### 后端启动（使用 Mage）

```bash
cd backend-go

# 安装 Mage
go install github.com/magefile/mage@latest

# 安装依赖
mage install

# 启动开发服务器（热重载）
mage dev

# 或直接运行
mage run
```

后端 API 将运行在 `http://localhost:8080`

查看 [后端文档](backend-go/README.md) 和 [构建工具文档](docs/mage/MAGE_QUICKSTART.md) 了解更多。

## 🎯 功能特性

### 核心功能

- ✅ **用户认证** - JWT + Token Rotation
- ✅ **账户管理** - 多账户、多币种、信用卡
- ✅ **交易记录** - 收入/支出/转账
- ✅ **预算管理** - 月度预算设置和追踪
- ✅ **分期计划** - 分期付款管理
- ✅ **周期收入** - 工资等周期性收入
- ✅ **现金流预测** - 未来资金预测
- ✅ **报表分析** - 多维度数据分析

### 技术特性

- ✅ **完整的测试** - 120+ 测试用例
- ✅ **多用户并发** - 支持高并发场景
- ✅ **数据一致性** - 完善的事务处理
- ✅ **API 文档** - 清晰的接口说明
- ✅ **日志系统** - 完整的日志追踪

查看 [功能文档](docs/FEATURE_DOCUMENTATION.md) 了解详情。

## API 接口

### 首页
- `GET /api/v1/home/overview` - 获取首页概览

### 账本
- `GET /api/v1/transactions/daily?date=YYYY-MM-DD` - 获取每日账单

### 报表
- `GET /api/v1/reports/monthly?month=YYYY-MM` - 获取月度报表

### 用户
- `GET /api/v1/user/settings` - 获取用户设置
- `PATCH /api/v1/user/settings` - 更新用户设置

## 🧪 测试

项目拥有完整的测试体系：

```bash
cd backend-go

# 运行所有测试
mage test:all

# 运行单元测试
mage test:unit

# 运行集成测试
mage test:integration

# 运行多用户并发测试
mage test:concurrent

# 运行压力测试（300+ 并发）
mage test:stress
```

查看 [测试体系总览](docs/TESTING_OVERVIEW.md) 了解更多。

## 🔨 构建命令

使用 Mage 构建工具：

```bash
cd backend-go

# 查看所有命令
mage -l

# 构建应用
mage build

# 运行 CI 检查
mage ci

# 清理构建产物
mage clean
```

查看 [Mage 指南](docs/mage/MAGE_GUIDE.md) 了解所有命令。

## 国际化

支持语言：
- 🇨🇳 简体中文 (zh-CN)
- 🇺🇸 English (en-US)

语言文件位于 `frontend-vue/src/i18n/`

## 贡献

欢迎提交 Issue 和 Pull Request！

## 📖 文档

完整的文档体系：

- 📚 [文档中心](docs/README.md) - 文档入口
- 📋 [功能文档](docs/FEATURE_DOCUMENTATION.md) - 完整功能说明
- 🧪 [测试体系](docs/TESTING_OVERVIEW.md) - 测试架构
- 🔨 [Mage 指南](docs/mage/MAGE_GUIDE.md) - 构建工具
- 🔧 [后端开发](backend-go/README.md) - 后端指南
- 🎨 [前端开发](frontend-vue/README.md) - 前端指南

## 📊 项目统计

- **代码行数**: ~15,000+ 行
- **测试用例**: 120+ 个
- **测试覆盖**: 85%+
- **文档页数**: 10+ 个主要文档

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

开发前请阅读：
1. [后端开发指南](backend-go/README.md)
2. [测试指南](docs/testing/TESTING_GUIDE.md)
3. 提交前运行 `mage ci` 检查

## 许可证

MIT License

