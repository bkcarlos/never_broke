# NeverBroke - 认知翻身账

一个快速、简洁的记账应用，帮助你掌控财务，永不破产。

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
- **框架**: Gin
- **存储**: 内存存储（可扩展为数据库）

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

## 快速开始

### 前端启动

```bash
cd frontend-vue
npm install
npm run dev
```

前端将运行在 `http://localhost:3000`

### 后端启动

```bash
cd backend-go
go mod download
go run ./cmd/server
```

后端 API 将运行在 `http://localhost:8080`

## 功能特性

### 当前版本 (v0.1)

- ✅ **首页概览**: 今日支出、本月预算、资产总览
- ✅ **账单明细**: 按日期查看收支记录
- ✅ **操作中心**: 快速记账入口（占位）
- ✅ **月度报表**: 分类统计、每日趋势
- ✅ **个人设置**: 多语言切换、金额显隐

### 计划功能

- 🔲 记一笔（支出/收入/转账）
- 🔲 账户管理
- 🔲 预算设置
- 🔲 数据导入导出
- 🔲 未来现金流预测
- 🔲 多用户支持

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

## 开发说明

### 前端开发

```bash
# 开发模式
npm run dev

# 构建生产版本
npm run build

# 预览生产版本
npm run preview
```

### 后端开发

```bash
# 运行服务
go run ./cmd/server

# 构建二进制
go build -o neverbroke ./cmd/server

# 运行测试
go test ./...
```

## 国际化

支持语言：
- 🇨🇳 简体中文 (zh-CN)
- 🇺🇸 English (en-US)

语言文件位于 `frontend-vue/src/i18n/`

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

