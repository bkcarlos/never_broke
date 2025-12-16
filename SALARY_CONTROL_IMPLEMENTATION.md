# 薪资控制功能实施完成报告

## 概述

成功实现了「我的」Tab下的独立「薪资控制」页面,将薪资管理从周期收入中独立出来,提供完整的工资方案配置、薪资参数管理和税后收入预览功能。

## 实施内容

### 一、数据库扩展

#### 1. 新增城市社保比例表 (`city_social_rates`)
- **文件**: `backend-go/migrations/0002_add_salary_tables.sql`
- **功能**: 管理各城市的社保和公积金缴纳比例,支持详细拆分和历史记录
- **字段**:
  - 基本信息: id, city_code, city_name, effective_date
  - 社保比例: pension_rate(养老), medical_rate(医疗), unemployment_rate(失业), housing_fund_rate(公积金)
  - 总比例: total_rate
  - 状态: is_active

#### 2. 扩展周期收入计划表 (`recurring_income_plans`)
- **新增字段**:
  - `salary_pre_tax_monthly`: 税前月薪
  - `salary_social_rate`: 社保+公积金个人缴纳比例
  - `salary_extra_deductions`: 每月专项附加扣除合计
  - `salary_pay_day`: 发薪日(1-31)
  - `salary_start_date`: 工资方案开始日期

#### 3. 初始化数据
- **文件**: `backend-go/migrations/0003_seed_city_social_rates.sql`
- **包含城市**: 北京、上海、广州、深圳、其他城市(默认)
- **数据示例**:
  - 北京: 养老8% + 医疗2% + 失业0.2% + 公积金12% = 22.2%
  - 上海: 养老8% + 医疗2% + 失业0.5% + 公积金7% = 17.5%

### 二、后端实现

#### 1. 城市社保比例模块 (`citysocialrate`)
**新建文件**:
- `backend-go/internal/citysocialrate/model.go`: 数据模型定义
- `backend-go/internal/citysocialrate/store.go`: 数据访问层
- `backend-go/internal/citysocialrate/handler.go`: HTTP 处理器

**API 接口**:
- `GET /api/v1/city-social-rates`: 获取所有城市社保比例列表
- `GET /api/v1/city-social-rates/:cityCode`: 获取指定城市的社保比例详情

#### 2. 薪资方案管理
**扩展文件**:
- `backend-go/internal/recurringincome/model.go`: 扩展 RecurringIncomePlan 模型
- `backend-go/internal/recurringincome/handler.go`: 新增薪资方案 handlers
- `backend-go/internal/recurringincome/salary_tax.go`: 优化薪资计算逻辑

**新增 API 接口**:
- `GET /api/v1/me/salary-plan`: 获取当前用户的工资方案
- `POST /api/v1/me/salary-plan`: 保存/更新工资方案
- `DELETE /api/v1/me/salary-plan/:id`: 删除工资方案

**核心功能**:
- 支持自定义社保比例(覆盖城市默认值)
- 支持专项附加扣除(子女教育、房贷利息等)
- 自动计算税后金额并保存到周期收入计划
- 保存新方案时自动将旧方案设为 inactive

#### 3. 路由配置
**文件**: `backend-go/internal/http/router/router.go`
- 注册城市社保比例相关路由
- 注册薪资方案管理相关路由

#### 4. 初始化
**文件**: `backend-go/cmd/server/main.go`
- 在服务启动时初始化城市社保比例数据

### 三、前端实现

#### 1. 类型定义
**文件**: `frontend-vue/src/types/index.ts`
- 新增 `CitySocialRate`: 城市社保比例
- 新增 `CitySocialRateDetail`: 社保比例明细
- 新增 `SocialRateComponent`: 社保组成部分
- 新增 `SalaryPlanConfig`: 工资方案配置
- 扩展 `RecurringIncomePlan`: 添加薪资专用字段
- 扩展 `SalaryPreviewRequest`: 支持自定义参数

#### 2. API 封装
**新建文件**:
- `frontend-vue/src/api/citySocialRate.ts`: 城市社保比例 API
  - `listCitySocialRates()`: 获取所有城市列表
  - `getCitySocialRateDetail(cityCode)`: 获取城市详情

- `frontend-vue/src/api/meSalary.ts`: 薪资方案 API
  - `getSalaryPlan()`: 获取工资方案
  - `upsertSalaryPlan(payload)`: 保存/更新工资方案
  - `deleteSalaryPlan(planId)`: 删除工资方案

#### 3. 国际化文本
**文件**: 
- `frontend-vue/src/i18n/zh-CN.ts`: 中文文本(45+ 条)
- `frontend-vue/src/i18n/en-US.ts`: 英文文本(45+ 条)

**包含内容**:
- 页面标题和说明
- 表单字段标签和提示
- 预览表格标题
- 操作按钮文本
- 成功/错误提示信息

#### 4. 薪资控制页面
**文件**: `frontend-vue/src/views/MeSalaryPage.vue`

**页面结构**:

##### Section 1: 当前工资方案卡片
- 显示当前 active 的工资方案
- 展示: 税前月薪、发薪日、到账账户、纳税城市、状态
- 操作: 编辑、删除按钮

##### Section 2: 工资参数配置表单
表单字段:
- 税前月薪 (number input)
- 纳税城市 (select: 北京/上海/广州/深圳/其他)
  - 选择城市后自动加载社保比例详情
  - 显示社保比例明细表格(养老/医疗/失业/公积金)
- 社保+公积金比例 (number input, 0-100%)
  - 默认值为所选城市的总比例
  - 用户可自定义修改
  - 显示提示: "当前城市默认: XX%"
- 专项附加扣除 (number input)
- 开始日期 (date picker)
- 发薪日 (number input, 1-31)
- 到账账户 (select, 从 accounts API 获取)
- 操作按钮: "保存为工资方案" / "仅预览"

##### Section 3: 薪资计算器预览
- 顶部汇总卡片:
  - 全年税前总额
  - 全年个税总额
  - 全年税后总额
- 月度明细表格:
  - 列: 月份、税前、社保+公积金、个税、税后到手
- 底部提示: "仅供参考,以实际发薪与当地政策为准"

**技术特点**:
- Vue 3 Composition API + TypeScript
- 完整的表单验证
- 加载状态管理
- 错误处理和用户提示
- 响应式设计

#### 5. 路由配置
**文件**: `frontend-vue/src/router/index.ts`
- 新增路由: `/me/salary` -> `MeSalaryPage.vue`

#### 6. 入口集成
**文件**: `frontend-vue/src/views/ProfilePage.vue`
- 在「我的」页面新增「薪资控制」入口
- 位置: 在「周期收入配置」之前
- 显示标题和副标题说明

## 核心功能特性

### 1. 城市社保比例管理
- 支持多个城市的社保比例配置
- 详细拆分各项比例(养老/医疗/失业/公积金)
- 支持历史记录(通过 effective_date 和 is_active 管理)
- 前端自动展示社保比例明细

### 2. 工资方案配置
- 支持自定义社保比例(覆盖城市默认值)
- 支持专项附加扣除配置
- 自动计算税后金额
- 保存时自动将旧方案设为 inactive
- 只允许一个 active 的工资方案

### 3. 薪资预览计算
- 基于中国个税累进税率表
- 支持从开始日期到年底的逐月预览
- 显示每月详细: 税前/社保/个税/税后
- 显示全年汇总数据
- 计算累计个税(符合中国税法)

### 4. 数据一致性
- 工资方案的 `Amount` 字段存储税后金额
- 税前金额和计算参数存储在专用字段
- 与现金流预测系统无缝集成

## 数据流设计

### 保存工资方案流程
1. 用户填写工资参数
2. 前端调用 `POST /me/salary-plan`
3. 后端调用 `PreviewSalaryToYearEnd()` 计算税后金额
4. 取首月税后金额作为 `Amount`
5. 查询现有 salary 计划并设为 inactive
6. 保存新计划
7. 返回计划和预览结果

### 加载页面流程
1. 访问 `/me/salary`
2. 并行加载: 工资方案 + 账户列表
3. 如果存在工资方案,填充表单
4. 加载城市社保比例详情
5. 渲染页面

## 技术亮点

1. **模块化设计**: 城市社保比例独立为单独模块
2. **数据库设计**: 支持历史记录和多版本管理
3. **类型安全**: 全程 TypeScript,无 any 类型
4. **国际化**: 完整的中英文支持
5. **用户体验**: 
   - 自动加载城市默认比例
   - 实时显示社保明细
   - 表单验证和错误提示
   - 加载状态反馈

## 注意事项

1. **税法免责**: 页面显著位置注明"仅供参考"
2. **数据验证**: 
   - 税前月薪 > 0
   - 社保比例 0-100%
   - 发薪日 1-31
   - 日期格式正确
3. **向后兼容**: 新增字段使用 `omitempty`,不影响现有功能
4. **性能优化**: 城市社保数据在服务启动时初始化到内存

## 后续扩展建议

1. **社保比例管理界面**: 管理员可在后台更新城市社保比例
2. **多地工作**: 支持同时配置多个城市的工资
3. **年终奖**: 单独配置年终奖的税务计算
4. **历史记录**: 查看历史工资方案和实际发薪记录对比
5. **智能提醒**: 发薪日前一天推送提醒
6. **数据导出**: 导出全年薪资预览为 Excel/PDF
7. **社保基数上下限**: 支持配置各城市的社保缴费基数上下限

## 文件清单

### 后端文件 (Go)
- `backend-go/migrations/0002_add_salary_tables.sql` (新建)
- `backend-go/migrations/0003_seed_city_social_rates.sql` (新建)
- `backend-go/internal/citysocialrate/model.go` (新建)
- `backend-go/internal/citysocialrate/store.go` (新建)
- `backend-go/internal/citysocialrate/handler.go` (新建)
- `backend-go/internal/recurringincome/model.go` (修改)
- `backend-go/internal/recurringincome/handler.go` (修改)
- `backend-go/internal/recurringincome/salary_tax.go` (修改)
- `backend-go/internal/http/router/router.go` (修改)
- `backend-go/cmd/server/main.go` (修改)

### 前端文件 (Vue + TypeScript)
- `frontend-vue/src/types/index.ts` (修改)
- `frontend-vue/src/api/citySocialRate.ts` (新建)
- `frontend-vue/src/api/meSalary.ts` (新建)
- `frontend-vue/src/i18n/zh-CN.ts` (修改)
- `frontend-vue/src/i18n/en-US.ts` (修改)
- `frontend-vue/src/views/MeSalaryPage.vue` (新建)
- `frontend-vue/src/views/ProfilePage.vue` (修改)
- `frontend-vue/src/router/index.ts` (修改)

## 总结

成功实现了完整的薪资控制功能,包括:
- ✅ 数据库表设计和迁移
- ✅ 城市社保比例管理模块
- ✅ 薪资方案 CRUD 操作
- ✅ 薪资计算和预览功能
- ✅ 完整的前端页面和交互
- ✅ 国际化支持
- ✅ 与现有系统集成

所有功能已实现并可以投入使用。
