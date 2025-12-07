# NeverBroke Frontend (Vue 3)

NeverBroke 记账应用的前端项目，使用 Vue 3 + TypeScript + Vite 构建。

## 技术栈

- Vue 3 (Composition API)
- TypeScript
- Vite
- Vue Router 4
- Pinia
- Tailwind CSS

## 项目结构

```
frontend-vue/
├── src/
│   ├── views/              # 页面组件
│   │   ├── HomePage.vue
│   │   ├── DetailPage.vue
│   │   ├── ActionsPage.vue
│   │   ├── ReportsPage.vue
│   │   └── ProfilePage.vue
│   ├── components/         # 公共组件
│   ├── api/                # API 封装
│   │   ├── http.ts
│   │   ├── home.ts
│   │   ├── ledger.ts
│   │   ├── report.ts
│   │   └── user.ts
│   ├── types/              # TypeScript 类型定义
│   │   └── index.ts
│   ├── i18n/               # 国际化
│   │   ├── index.ts
│   │   ├── zh-CN.ts
│   │   └── en-US.ts
│   ├── router/             # 路由配置
│   │   └── index.ts
│   ├── App.vue
│   ├── main.ts
│   └── style.css
├── index.html
├── vite.config.ts
├── tailwind.config.js
└── package.json
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

应用将运行在 `http://localhost:3000`

### 构建生产版本

```bash
npm run build
```

### 预览生产版本

```bash
npm run preview
```

## 页面说明

### 首页 (/)
- 今日支出、今日可花
- 本月预算、已用、结余
- 资产总览（总资产、总负债、净资产）
- 支出分类 TOP 5

### 明细页 (/detail)
- 日期选择器
- 当日收支统计
- 账单列表（时间、分类、金额、备注）

### 操作页 (/actions)
- 记支出
- 记收入
- 转账
- 设置预算
- 新增账户
- 导入账单

### 报表页 (/reports)
- 月份选择器
- 月度收支汇总
- 分类统计（金额、占比）
- 每日趋势

### 我的 (/profile)
- 语言切换（中文/English）
- 金额显隐开关

## API 配置

后端 API 地址配置在 `vite.config.ts` 中：

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

## 国际化

支持中文和英文两种语言，语言文件位于 `src/i18n/`：

- `zh-CN.ts` - 简体中文
- `en-US.ts` - English

使用方式：

```typescript
import { t } from '@/i18n'

// 在模板中
{{ t('home.title') }}

// 在脚本中
const title = t('home.title')
```

## 样式

使用 Tailwind CSS 进行样式开发，配置文件：`tailwind.config.js`

## 开发说明

### 添加新页面

1. 在 `src/views/` 创建新的 Vue 组件
2. 在 `src/router/index.ts` 添加路由
3. 在 `src/App.vue` 底部导航添加入口（如需要）

### 添加新 API

1. 在 `src/types/index.ts` 定义类型
2. 在 `src/api/` 创建对应的 API 文件
3. 使用 `http.ts` 中的 `get`、`post`、`patch` 等方法

### 添加国际化文案

在 `src/i18n/zh-CN.ts` 和 `src/i18n/en-US.ts` 中添加对应的 key-value。

## 代码规范

- 使用 TypeScript，避免使用 `any`
- 使用 `<script setup lang="ts">` 语法
- 组件使用 Composition API
- 遵循 Vue 3 最佳实践

## 许可证

MIT License

