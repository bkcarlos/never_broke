# Mage 构建工具使用指南

> Mage 是一个基于 Go 的构建工具，可以替代 Makefile，让你用 Go 代码来定义构建任务。

## 🚀 快速开始

### 安装 Mage

```bash
# 方式1: 使用 go install
go install github.com/magefile/mage@latest

# 方式2: 使用 homebrew (macOS)
brew install mage

# 验证安装
mage -version
```

### 查看所有可用命令

```bash
# 列出所有命令
mage -l

# 查看详细帮助
mage -h
```

---

## 📋 命令清单

### 🧪 测试相关命令

所有测试命令都在 `test:` 命名空间下：

```bash
# 运行单元测试
mage test:unit

# 运行集成测试
mage test:integration

# 快速运行集成测试（不显示详细输出）
mage test:integrationQuick

# 运行所有测试（单元 + 集成）
mage test:all

# 生成覆盖率报告
mage test:coverage

# 生成 HTML 覆盖率报告
mage test:coverageHTML

# 运行竞态检测
mage test:race

# 运行认证相关测试
mage test:auth

# 运行账本相关测试
mage test:ledger

# 运行多用户并发测试
mage test:concurrent

# 运行压力测试
mage test:stress

# 运行基准测试
mage test:benchmark

# 设置测试数据库
mage test:setupDB

# 清理测试数据库
mage test:cleanDB

# 演示所有复杂测试场景
mage test:demo
```

### 🗄️ 数据库相关命令

所有数据库命令都在 `database:` 命名空间下：

```bash
# 运行数据库迁移
mage database:migrate

# 回滚数据库迁移
mage database:migrateDown
```

### 🔨 构建相关命令

```bash
# 构建应用（当前平台）
mage build

# 构建所有平台
mage buildAll

# 运行开发服务器
mage run

# 使用 air 热重载运行
mage dev
```

### 🧹 清理命令

```bash
# 清理构建产物
mage clean

# 清理所有（包括测试数据库）
mage cleanAll
```

### 📦 依赖管理

```bash
# 安装依赖
mage install
```

### ✨ 代码质量

```bash
# 格式化代码
mage format

# 运行 linter
mage lint

# 运行 go vet
mage vet
```

### 🚀 CI/CD

```bash
# 运行完整的 CI 流程
mage ci
```

### ℹ️ 信息命令

```bash
# 显示版本信息
mage version

# 显示项目信息和统计
mage info
```

### 🔧 开发工具

```bash
# 安装所有开发工具
mage installTools
```

---

## 🎯 常用场景

### 场景1: 开始开发

```bash
# 1. 安装依赖
mage install

# 2. 设置测试数据库
mage test:setupDB

# 3. 启动热重载开发服务器
mage dev
```

### 场景2: 运行测试

```bash
# 快速运行单元测试
mage test:unit

# 运行所有测试
mage test:all

# 运行特定模块测试
mage test:auth          # 认证测试
mage test:ledger        # 账本测试
mage test:concurrent    # 并发测试
mage test:stress        # 压力测试
```

### 场景3: 测试覆盖率

```bash
# 生成覆盖率报告
mage test:coverage

# 生成 HTML 报告并在浏览器打开
mage test:coverageHTML
open coverage_unit.html
```

### 场景4: 体验复杂测试场景

```bash
# 演示所有复杂场景
mage test:demo

# 或者运行特定场景
mage test:concurrent    # 多用户并发场景
mage test:stress        # 压力测试场景
```

### 场景5: CI/CD 流程

```bash
# 运行完整的 CI 检查
mage ci
```

### 场景6: 构建发布

```bash
# 构建当前平台
mage build

# 构建所有平台
mage buildAll
```

### 场景7: 清理环境

```bash
# 清理构建产物
mage clean

# 清理所有（包括测试数据库）
mage cleanAll
```

---

## 🆚 Mage vs Makefile

### 优势

✅ **使用 Go 语法** - 不需要学习 Make 的特殊语法  
✅ **跨平台** - 在 Windows, macOS, Linux 上都能正常工作  
✅ **类型安全** - 编译时检查错误  
✅ **IDE 支持** - 代码补全、跳转等功能  
✅ **命名空间** - 更好地组织命令  
✅ **依赖管理** - `mg.Deps()` 自动处理依赖  

### 对比示例

**Makefile 方式:**
```makefile
test-unit:
	@echo "Running unit tests..."
	@go test ./internal/... -v -short

test-integration:
	@echo "Running integration tests..."
	@cd tests/integration && ./run_tests.sh -v
```

**Mage 方式:**
```go
// Unit 运行单元测试
func (Test) Unit() error {
	fmt.Println("🧪 Running unit tests...")
	return sh.RunV("go", "test", "./internal/...", "-v", "-short")
}

// Integration 运行集成测试
func (Test) Integration() error {
	fmt.Println("🧪 Running integration tests...")
	return sh.RunV("bash", "-c", "cd tests/integration && ./run_tests.sh -v")
}
```

---

## 🎨 Mage 特性

### 1. 命名空间

使用命名空间组织相关命令：

```go
type Test mg.Namespace

func (Test) Unit() error { ... }
func (Test) Integration() error { ... }
```

使用：
```bash
mage test:unit
mage test:integration
```

### 2. 依赖管理

使用 `mg.Deps()` 自动运行依赖任务：

```go
func Build() error {
    mg.Deps(Install)  // 自动先运行 Install
    // ... 构建逻辑
}
```

### 3. 变量和常量

在代码顶部定义配置：

```go
const (
    appName  = "neverbroke"
    mainPath = "./cmd/server"
    version  = "v0.1.0"
)
```

### 4. 辅助函数

可以定义辅助函数：

```go
func getGitCommit() string {
    commit, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
    return strings.TrimSpace(commit)
}
```

---

## 📝 添加自定义命令

### 1. 添加简单命令

在 `magefile.go` 中添加函数：

```go
// Hello 打印欢迎信息
func Hello() error {
    fmt.Println("Hello from NeverBroke!")
    return nil
}
```

使用：
```bash
mage hello
```

### 2. 添加命名空间命令

创建命名空间类型并添加方法：

```go
type Docker mg.Namespace

// Build 构建 Docker 镜像
func (Docker) Build() error {
    return sh.RunV("docker", "build", "-t", "neverbroke-api", ".")
}

// Run 运行 Docker 容器
func (Docker) Run() error {
    return sh.RunV("docker", "run", "-p", "8080:8080", "neverbroke-api")
}
```

使用：
```bash
mage docker:build
mage docker:run
```

---

## 🔍 调试

### 查看实际执行的命令

使用 `-v` 标志：

```bash
mage -v test:unit
```

### 查看命令解析

```bash
mage -d test:unit
```

---

## 📚 学习资源

- [Mage 官方文档](https://magefile.org/)
- [Mage GitHub](https://github.com/magefile/mage)
- [本项目的 magefile.go](./magefile.go)

---

## 🎓 最佳实践

### 1. 使用有意义的命令名

```go
// ✅ 好
func (Test) Unit() error { ... }

// ❌ 不好
func (Test) T1() error { ... }
```

### 2. 添加清晰的输出

```go
func Build() error {
    fmt.Println("🔨 Building application...")
    // ...
    fmt.Println("✅ Build successful")
}
```

### 3. 使用命名空间组织

```go
type Test mg.Namespace
type Database mg.Namespace
type Docker mg.Namespace
```

### 4. 处理错误

```go
func Build() error {
    if err := sh.Run("go", "build", ...); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }
    return nil
}
```

### 5. 使用 mg.Deps 管理依赖

```go
func Build() error {
    mg.Deps(Install)  // 确保依赖已安装
    // 构建逻辑
}
```

---

## 🚀 迁移说明

### 从 Makefile 迁移

如果你之前使用 Makefile，迁移到 Mage 很简单：

**之前 (Makefile):**
```bash
make test-unit
make test-integration
make test-all
```

**现在 (Mage):**
```bash
mage test:unit
mage test:integration
mage test:all
```

主要区别：
1. 命令前缀从 `make` 改为 `mage`
2. 命令中的 `-` 改为 `:`（命名空间分隔符）
3. 所有测试命令都在 `test:` 命名空间下

### 保留 Makefile

Makefile 仍然保留在项目中，你可以选择使用：
- ✅ **推荐**: `mage` - 更现代、跨平台
- ✅ **可选**: `make` - 如果你熟悉或有特殊需求

两者功能完全相同！

---

## 🎉 总结

Mage 让构建任务管理变得更加简单和强大：

✅ Go 语法，无需学习新语言  
✅ 完整的 IDE 支持  
✅ 跨平台兼容  
✅ 类型安全  
✅ 易于扩展  

**开始使用 Mage，让你的构建流程更加优雅！** 🚀
