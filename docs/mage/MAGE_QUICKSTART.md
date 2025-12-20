# Mage 快速开始

## 📦 安装 Mage

### 方式1: 使用 go install（推荐）

```bash
go install github.com/magefile/mage@latest
```

### 方式2: 使用 Homebrew (macOS)

```bash
brew install mage
```

### 验证安装

```bash
mage -version
```

应该输出类似：`Mage Build Tool v1.15.0`

---

## 🚀 快速体验

### 1. 查看所有命令

```bash
cd backend-go
mage -l
```

你会看到类似输出：

```
Targets:
  build              编译二进制文件
  buildAll           为多个平台交叉编译
  ci                 CI 流程（验证、格式化、检查、测试）
  clean              清理构建产物
  cleanAll           清理所有（包括测试数据库）
  database:migrate   运行数据库迁移
  database:migrateDown  回滚数据库迁移
  dev                使用 air 热重载运行开发服务器
  format             格式化代码
  info               显示项目信息和统计
  install            安装依赖
  installTools       安装开发工具
  lint               运行代码检查（需要安装 golangci-lint）
  run                运行开发服务器
  test:all           运行所有测试（单元测试 + 集成测试）
  test:auth          只运行认证相关测试
  test:benchmark     运行性能基准测试
  test:cleanDB       清理测试数据库
  test:concurrent    运行多用户并发测试
  test:coverage      运行测试并生成覆盖率报告
  test:coverageHTML  生成 HTML 格式的覆盖率报告
  test:demo          演示所有复杂测试场景
  test:integration   运行集成测试
  test:integrationQuick  快速运行集成测试（不显示详细输出）
  test:ledger        只运行账本相关测试
  test:race          运行测试并启用竞态检测
  test:setupDB       设置测试数据库
  test:stress        运行压力测试和真实场景测试
  test:unit          运行单元测试
  version            显示版本信息
  vet                运行 go vet
```

### 2. 运行第一个命令

```bash
# 显示项目信息
mage info

# 或运行版本命令
mage version
```

### 3. 运行测试

```bash
# 运行单元测试（快速）
mage test:unit

# 运行所有测试
mage test:all
```

---

## 🎯 常用场景

### 场景1: 开发环境设置

```bash
cd backend-go

# 1. 安装依赖
mage install

# 2. 安装开发工具（air, golangci-lint 等）
mage installTools

# 3. 设置测试数据库
mage test:setupDB

# 4. 启动热重载开发服务器
mage dev
```

### 场景2: 运行测试

```bash
# 单元测试（快速）
mage test:unit

# 集成测试
mage test:integration

# 所有测试
mage test:all

# 特定模块测试
mage test:auth          # 认证模块
mage test:ledger        # 账本模块

# 复杂场景测试
mage test:concurrent    # 多用户并发
mage test:stress        # 压力测试
```

### 场景3: 代码质量检查

```bash
# 格式化代码
mage format

# 静态检查
mage vet

# Linter 检查
mage lint

# 或运行完整 CI 流程
mage ci
```

### 场景4: 构建

```bash
# 构建当前平台
mage build

# 构建所有平台
mage buildAll

# 运行
mage run
```

### 场景5: 清理

```bash
# 清理构建产物
mage clean

# 清理所有（包括测试数据库）
mage cleanAll
```

---

## 💡 使用技巧

### 1. 命令补全

某些 shell 支持 Mage 命令补全。查看 [Mage 文档](https://magefile.org) 了解如何为你的 shell 设置。

### 2. 详细输出

使用 `-v` 查看执行的实际命令：

```bash
mage -v test:unit
```

### 3. 调试模式

使用 `-d` 查看调试信息：

```bash
mage -d test:unit
```

### 4. 运行多个命令

```bash
# 依次执行多个命令
mage format vet test:unit
```

---

## 📚 文档资源

- [完整 Mage 使用指南](MAGE_GUIDE.md) - 详细的使用说明
- [Makefile 迁移指南](MIGRATION_TO_MAGE.md) - 从 Makefile 迁移
- [Mage 官方文档](https://magefile.org/) - 官方文档

---

## 🐛 常见问题

### Q: 命令未找到：mage

**A**: Mage 未安装或未在 PATH 中。运行：

```bash
go install github.com/magefile/mage@latest

# 确保 $GOPATH/bin 在你的 PATH 中
echo $PATH | grep "go/bin"
```

如果不在 PATH 中，添加到你的 shell 配置文件：

```bash
# Bash/Zsh
echo 'export PATH="$PATH:$HOME/go/bin"' >> ~/.bashrc
source ~/.bashrc
```

### Q: 找不到某个命令

**A**: 先运行 `mage -l` 查看所有可用命令。注意命名空间用 `:` 分隔：

```bash
# 错误
mage test-unit

# 正确
mage test:unit
```

### Q: 命令执行失败

**A**: 使用 `-v` 查看详细输出：

```bash
mage -v test:unit
```

### Q: 如何自定义命令？

**A**: 编辑 `magefile.go`，添加新函数。查看 [MAGE_GUIDE.md](MAGE_GUIDE.md) 了解详情。

---

## ✅ 验证安装

运行以下命令确认一切正常：

```bash
# 1. 检查 Mage 版本
mage -version

# 2. 列出所有命令
mage -l

# 3. 显示项目信息
mage info

# 4. 运行快速测试
mage test:unit
```

如果所有命令都能正常运行，说明 Mage 已正确安装！ 🎉

---

## 🚀 下一步

1. 阅读 [完整 Mage 使用指南](MAGE_GUIDE.md)
2. 查看 [测试系统文档](tests/)
3. 开始开发！

Happy coding! 🎉
