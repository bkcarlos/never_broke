# NeverBroke 项目文档索引

> 完整的文档导航和快速访问指南

---

## 📚 文档结构

```
NeverBroke/
├── README.md                              # 项目主页
│
├── docs/                                  # 📚 文档中心
│   ├── README.md                          # 文档入口
│   ├── DOCS_INDEX.md                      # 本文档
│   ├── FEATURE_DOCUMENTATION.md           # 完整功能文档
│   ├── TESTING_OVERVIEW.md                # 测试体系总览
│   │
│   ├── mage/                              # 🔨 Mage 构建工具
│   │   ├── MAGE_QUICKSTART.md             # 快速开始
│   │   └── MAGE_GUIDE.md                  # 完整指南
│   │
│   └── testing/                           # 🧪 测试文档
│       ├── TESTING_GUIDE.md               # 完整测试指南
│       ├── TEST_SUMMARY.md                # 测试总结报告
│       └── integration/                   # 集成测试
│           ├── QUICKSTART.md              # 快速开始
│           └── COMPLEX_SCENARIOS.md       # 复杂场景
│
├── backend-go/                            # 后端代码
│   ├── README.md                          # 后端入口
│   ├── common/logger/README.md            # 日志系统
│   └── tests/
│       ├── README.md                      # 测试中心
│       └── integration/README.md          # 集成测试说明
│
└── frontend-vue/                          # 前端代码
    └── README.md                          # 前端入口
```

---

## 🎯 快速导航

### 新手入门

**刚接触项目？从这里开始：**

1. [项目 README](../README.md) - 了解项目概况
2. [后端 README](../backend-go/README.md) - 后端快速开始
3. [Mage 快速开始](mage/MAGE_QUICKSTART.md) - 5分钟学会构建工具
4. [测试快速开始](testing/integration/QUICKSTART.md) - 5分钟运行测试

### 开发者指南

**准备开发？看这些：**

1. [后端 README](backend-go/README.md) - 完整开发指南
2. [Mage 完整指南](backend-go/MAGE_GUIDE.md) - 所有构建命令
3. [测试体系总览](TESTING_OVERVIEW.md) - 理解测试架构
4. [测试指南](backend-go/tests/TESTING_GUIDE.md) - 如何编写测试

### 测试相关

**想运行或编写测试？**

1. [测试体系总览](TESTING_OVERVIEW.md) - **入口文档**
2. [测试快速开始](testing/integration/QUICKSTART.md) - 快速上手
3. [复杂场景说明](testing/integration/COMPLEX_SCENARIOS.md) - 高级测试
4. [测试总结报告](testing/TEST_SUMMARY.md) - 完整测试覆盖

### 功能查询

**想了解具体功能？**

1. [功能文档](FEATURE_DOCUMENTATION.md) - 所有功能说明

---

## 📖 文档详解

### 项目级文档

#### [README.md](../README.md)
- **用途**: 项目主页和概览
- **包含**: 项目介绍、快速开始、技术栈
- **适合**: 所有人

#### [FEATURE_DOCUMENTATION.md](FEATURE_DOCUMENTATION.md)
- **用途**: 完整的功能文档
- **包含**: 所有功能的详细说明、API、数据模型
- **适合**: 开发者、产品经理

#### [TESTING_OVERVIEW.md](TESTING_OVERVIEW.md)
- **用途**: 测试体系总览（测试文档入口）
- **包含**: 测试架构、复杂场景、快速使用
- **适合**: 所有开发者

---

### 后端文档

#### [backend-go/README.md](../backend-go/README.md)
- **用途**: 后端开发完整指南
- **包含**: 
  - 快速开始（Mage + Go）
  - 开发工作流
  - API 端点
  - 测试运行
  - 相关文档链接
- **适合**: 后端开发者

#### [backend-go/common/logger/README.md](../backend-go/common/logger/README.md)
- **用途**: 日志系统使用说明
- **包含**: 日志 API、中间件、最佳实践
- **适合**: 后端开发者

---

### Mage 构建工具文档

#### [mage/MAGE_QUICKSTART.md](mage/MAGE_QUICKSTART.md)
- **用途**: Mage 快速开始（5分钟）
- **包含**: 
  - 安装 Mage
  - 常用命令
  - 使用场景
  - 故障排除
- **适合**: 新手、快速查询

#### [mage/MAGE_GUIDE.md](mage/MAGE_GUIDE.md)
- **用途**: Mage 完整使用指南
- **包含**: 
  - 所有命令详解
  - 命令示例
  - 最佳实践
  - 高级用法
- **适合**: 深度使用、参考查询

---

### 测试文档

#### [backend-go/tests/README.md](../backend-go/tests/README.md)
- **用途**: 测试中心导航
- **包含**: 
  - 测试体系概览
  - 快速导航链接
  - 常用命令
- **适合**: 测试文档导航

#### [testing/TESTING_GUIDE.md](testing/TESTING_GUIDE.md)
- **用途**: 完整测试指南
- **包含**: 
  - 测试架构
  - 单元测试
  - 集成测试
  - 编写测试
  - 最佳实践
- **适合**: 深入理解测试、编写测试

#### [testing/TEST_SUMMARY.md](testing/TEST_SUMMARY.md)
- **用途**: 测试总结报告
- **包含**: 
  - 所有测试用例
  - 覆盖率统计
  - 测试场景
  - 使用方式
- **适合**: 了解测试覆盖、快速查询

#### [testing/integration/QUICKSTART.md](testing/integration/QUICKSTART.md)
- **用途**: 集成测试快速开始（5分钟）
- **包含**: 
  - 环境设置
  - 运行测试
  - 常见问题
- **适合**: 新手、快速上手

#### [backend-go/tests/integration/README.md](../backend-go/tests/integration/README.md)
- **用途**: 集成测试详细文档
- **包含**: 
  - 测试内容
  - 运行方式
  - 环境配置
  - 故障排除
- **适合**: 深入使用集成测试

#### [testing/integration/COMPLEX_SCENARIOS.md](testing/integration/COMPLEX_SCENARIOS.md)
- **用途**: 复杂测试场景说明
- **包含**: 
  - 多用户并发
  - 真实场景
  - 压力测试
  - 场景详解
- **适合**: 了解高级测试、压力测试

---

### 前端文档

#### [frontend-vue/README.md](../frontend-vue/README.md)
- **用途**: 前端开发指南
- **包含**: 前端快速开始、开发说明
- **适合**: 前端开发者

---

## 🎨 文档分类

### 按角色分类

#### 👶 新手
1. [项目 README](../README.md)
2. [Mage 快速开始](mage/MAGE_QUICKSTART.md)
3. [测试快速开始](testing/integration/QUICKSTART.md)

#### 👨‍💻 后端开发者
1. [后端 README](../backend-go/README.md)
2. [功能文档](FEATURE_DOCUMENTATION.md)
3. [Mage 完整指南](mage/MAGE_GUIDE.md)
4. [测试指南](testing/TESTING_GUIDE.md)

#### 🧪 测试工程师
1. [测试体系总览](TESTING_OVERVIEW.md)
2. [测试指南](testing/TESTING_GUIDE.md)
3. [测试总结](testing/TEST_SUMMARY.md)
4. [复杂场景](testing/integration/COMPLEX_SCENARIOS.md)

#### 🎨 前端开发者
1. [前端 README](../frontend-vue/README.md)

---

### 按目的分类

#### 📖 学习理解
- [项目 README](../README.md) - 项目概览
- [功能文档](FEATURE_DOCUMENTATION.md) - 功能说明
- [测试体系总览](TESTING_OVERVIEW.md) - 测试架构
- [测试指南](testing/TESTING_GUIDE.md) - 测试详解

#### ⚡ 快速上手
- [Mage 快速开始](mage/MAGE_QUICKSTART.md) - 5分钟学会 Mage
- [测试快速开始](testing/integration/QUICKSTART.md) - 5分钟运行测试

#### 📚 参考查询
- [Mage 完整指南](mage/MAGE_GUIDE.md) - 所有命令
- [测试总结](testing/TEST_SUMMARY.md) - 所有测试
- [复杂场景](testing/integration/COMPLEX_SCENARIOS.md) - 高级测试

#### 🔧 实际操作
- [后端 README](../backend-go/README.md) - 开发指南
- [集成测试 README](../backend-go/tests/integration/README.md) - 运行测试

---

## 🔗 推荐阅读路径

### 路径1: 快速上手

```
../README.md (项目主页)
    ↓
../backend-go/README.md (后端入口)
    ↓
mage/MAGE_QUICKSTART.md (构建工具)
    ↓
开始开发！
```

### 路径2: 深入理解

```
../README.md
    ↓
FEATURE_DOCUMENTATION.md
    ↓
TESTING_OVERVIEW.md
    ↓
testing/TESTING_GUIDE.md
    ↓
完整掌握！
```

### 路径3: 测试专家

```
TESTING_OVERVIEW.md
    ↓
testing/integration/QUICKSTART.md
    ↓
testing/TESTING_GUIDE.md
    ↓
testing/integration/COMPLEX_SCENARIOS.md
    ↓
testing/TEST_SUMMARY.md
```

---

## 📝 文档维护

### 更新规则

- ✅ **功能更新**: 更新 `FEATURE_DOCUMENTATION.md`
- ✅ **测试新增**: 更新测试相关文档
- ✅ **构建变更**: 更新 Mage 文档
- ❌ **不要**: 创建临时文档（使用现有结构）

### 文档层次

```
Level 1: README (入口)
Level 2: 总览/指南 (架构)
Level 3: 详细文档 (细节)
Level 4: 快速开始 (实践)
```

---

## 🎉 快速命令

### 查看所有文档

```bash
# 列出所有文档
find . -name "*.md" -type f | grep -v node_modules

# 搜索文档内容
grep -r "关键词" --include="*.md"
```

### 常用文档

```bash
# 项目主页
cat README.md

# 后端开发
cat backend-go/README.md

# Mage 快速开始
cat backend-go/MAGE_QUICKSTART.md

# 测试快速开始
cat backend-go/tests/integration/QUICKSTART.md
```

---

**保持文档简洁、有序、易用！** 📚
